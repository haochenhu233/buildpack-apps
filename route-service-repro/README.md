# Route-service egress_blocklist repro (viper-auth simulation)

Reproduces the post-upgrade **502** that viper-auth causes, and validates the fix.
Two tiny Go apps:

- **`target-app`** — the protected app. `GET /` returns `TARGET OK …` if the request
  made it through the route service.
- **`viper-sim`** — a minimal CF **route service** mimicking viper-auth:
  - no/invalid `X-Viper-Auth` header → **403** (like viper-ia),
  - valid header → forwards to the app via `X-CF-Forwarded-Url` → **200 TARGET OK**.

## Why the 502 happens (recap)
A route service sits in the path: **client → gorouter → `route_service_url` → gorouter → app**.
Since **routing-release 0.372.0**, gorouter's secure-by-default **`egress_blocklist`** returns
**502** when the `route_service_url` resolves to a **private/internal IP** (RFC1918 / link-local)
— *before* the route service is even reached. That's viper-mo (resolves private → 502);
viper-ia resolves non-private → reachable → 403.

## Deploy
```bash
cf push -f target-app/manifest.yml
cf push -f viper-sim/manifest.yml
APPS_DOMAIN=<your apps domain>          # e.g. apps.ocfp.ics-ocf-int-np.aws.travp.net
```

## A) Prove the route-service workflow works (no blocklist yet)
Put viper-sim on a **normal** (reachable) route, wire it up, and confirm auth + forwarding:
```bash
cf create-user-provided-service viper-rs -r https://viper-sim.$APPS_DOMAIN
cf bind-route-service $APPS_DOMAIN viper-rs --hostname target-app

curl -s https://target-app.$APPS_DOMAIN/                         # -> 403 (viper-sim rejects: no auth header)
curl -s https://target-app.$APPS_DOMAIN/ -H 'X-Viper-Auth: let-me-in'   # -> "TARGET OK …" (forwarded)
```
A 403 here already proves the route service is in the path and reachable.

## B) Reproduce the client 502 (route service on a PRIVATE IP)
The blocklist triggers only when `route_service_url` resolves to a private IP. Pick ONE:

**B1 — faithful: a private-resolving route.** Map viper-sim onto a domain whose host resolves
to a private IP (the same property viper-mo has). First find one:
```bash
# which of your domains resolve to a PRIVATE IP (10.x / 172.16-31 / 192.168 / 169.254)?
nslookup viper-sim.<some-internal-domain>
```
Then re-point the route service at that private route and re-bind:
```bash
cf map-route viper-sim <internal-private-domain> --hostname viper-sim
cf delete-service viper-rs -f        # recreate pointing at the private route
cf create-user-provided-service viper-rs -r https://viper-sim.<internal-private-domain>
cf bind-route-service $APPS_DOMAIN viper-rs --hostname target-app
curl -s -o /dev/null -w '%{http_code}\n' https://target-app.$APPS_DOMAIN/ -H 'X-Viper-Auth: let-me-in'
# -> 502  (gorouter refuses to dial the private route service; target-app never reached)
```

**B2 — deterministic (no private domain needed): add viper-sim's IP to the blocklist.**
This simulates viper-mo by making gorouter treat viper-sim's route as blocked, then the fix
removes it — validating the exact `egress_blocklist` override mechanism:
```bash
nslookup viper-sim.$APPS_DOMAIN          # note the IP, e.g. 203.0.113.10 -> /32
# temporarily set the blocklist to include that IP (ops file -> deploy; router-only roll):
#   router.route_services.egress_blocklist: [ 203.0.113.10/32 ]
curl -s -o /dev/null -w '%{http_code}\n' https://target-app.$APPS_DOMAIN/ -H 'X-Viper-Auth: let-me-in'
# -> 502
```

**Confirm it's the blocklist (not another 502):**
```bash
bosh -d cf ssh router -c 'grep -iE "route_service|egress|blocklist|RouteServiceDialControl" /var/vcap/sys/log/gorouter/gorouter.log | tail'
```

## C) Apply the fix and validate
Override the blocklist (router-only roll — see the ops-file guidance), then redeploy and retest:
```bash
# ops file sets router.route_services.egress_blocklist = []  (or omit the viper CIDR)
genesis deploy <sbx-env>      # confirm the diff touches ONLY router/*

curl -s https://target-app.$APPS_DOMAIN/ -H 'X-Viper-Auth: let-me-in'
# -> "TARGET OK …"   (502 gone: gorouter now dials the route service, which forwards to the app)
```
**Pass criteria:** the **502 disappears** and the gorouter egress-refusal log line is **absent**.
The before/after on that log line is the definitive proof the fix is correct.

## Cleanup
```bash
cf unbind-route-service $APPS_DOMAIN viper-rs --hostname target-app
cf delete-service viper-rs -f
cf delete viper-sim -f && cf delete target-app -f
```

## Notes
- `go_buildpack` builds both from source at stage time (Go stdlib only — offline-safe).
- The whole point: **403 = route service reachable; 502 = gorouter blocked from reaching it.**
  Your fix flips the 502 back to a real response.
