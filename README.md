# Buildpack test apps

Minimal apps to validate each CF buildpack after a rebuild/upgrade — especially on a new
stemcell (jammy→noble). Each app is tiny but exercises the **core path**: the buildpack
installs the runtime, the app binds `$PORT`, starts, and serves `GET /` returning the
**resolved runtime version** (dotnet also reports the detected cgroup memory limit).

## The #1 rule: target the LATEST runtime
The point is to catch bugs in the *newest* versions a buildpack ships (e.g. the .NET 10
cgroup-v2 OOM fixed in **10.0.11**). An app pinned to an **old** runtime is a false
negative — the `net8.0` test app that "passed" never exercised .NET 10 at all. So every
app here rolls forward to the newest version its buildpack bundles:

| App | How it targets latest |
|---|---|
| dotnet | `net10.0` + `<RollForward>LatestMinor</RollForward>` → newest 10.x (e.g. 10.0.11) |
| java | Spring Boot latest + `java.version` current LTS |
| nodejs | `engines.node` = current major |
| python | `runtime.txt` = current 3.x |
| go | `go.mod` go = current |
| php | `{PHP_84_LATEST}` placeholder → newest bundled 8.4 |
| ruby | `Gemfile` ruby = current |

To test a *specific* older version instead, pin it (comment in each app) and re-push.
After a buildpack version bump, **bump these targets** so the apps keep tracking the latest.

## Verify after pushing
`GET /` returns the resolved runtime — confirm it's the version you expect the buildpack
to install (not an old one):
```
curl https://<route>/     # dotnet → {"framework":".NET 10.0.11","detectedMemoryLimitBytes":...,"cgroup":"v2"}
```
For **dotnet**, compare `detectedMemoryLimitBytes` to the manifest `memory` (256M). If it
reports the host's RAM (GBs) instead of ~256M → the cgroup-v2 misdetection bug (runtime
< 10.0.11) → it will OOM on noble.

## Push
```
cf push -f <app>/manifest.yml
```
Each manifest sets the buildpack + stack. For the air-gap/offline buildpacks, set the
buildpack name to your `*_buildpack` and make sure the latest is the system default.

## Not included here
- **binary**: build the `go/` app (`GOOS=linux go build -o server`) and push the binary
  with `-b binary_buildpack` — no separate app needed.
- **hwc (windows)**: needs a Windows build; push a minimal ASP.NET app with
  `-s windows -b hwc_buildpack` separately.
