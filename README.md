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

## Prebuilt artifacts (air-gap)
Most apps stage from source, but **java** is compiled. The client runs in an air-gapped
environment and cannot build, so the Spring Boot fat JAR is committed as a prebuilt
artifact at `java/target/bptest-java-1.0.0.jar` (force-added past `.gitignore`). It is a
build output, not source. Rebuild it from `java/pom.xml` with
`mvn -DskipTests package` (or the Maven Docker image) whenever the pom or `src/` changes,
then re-add it with `git add -f`. The app's `GET /` reports the JRE the buildpack picked,
which is the number for the proof table.

## Vendored gems (air-gap)
The **ruby** app would otherwise run `bundle install` against rubygems.org at stage time,
which the air-gap blocks. So `ruby/Gemfile.lock` and `ruby/vendor/cache/*.gem` are
committed, and the buildpack installs from that cache with no network. The gems are
resolved for the `x86_64-linux` cell platform. puma and nio4r are cached as source gems
and compiled on the cell. Refresh them with `bundle lock` + `bundle cache` in a
`linux/amd64` ruby:3.4 container after any Gemfile change.

## hwc (Windows)
The **hwc** app covers the Windows stemcell path, which nothing else in the suite
exercises. It is a minimal ASP.NET app (`Web.config` + `default.aspx`) hosted by HWC
(Hostable Web Core) on a Windows cell. No Windows build is needed: the `.aspx` uses
inline code that ASP.NET compiles on the cell at first request. HWC installs no runtime
of its own, so `GET /` reports the **.NET Framework** version the Windows stemcell
provides. Push it onto the Windows isolation segment (oneagent-windows):
```
cf push bptest-hwc -s windows -b hwc_buildpack
```

## Not included here
- **binary**: the `binary_buildpack` installs no runtime, it just runs a precompiled
  executable, so there is no runtime version to prove. If you need the check, build the
  `go/` app static (`GOOS=linux CGO_ENABLED=0 go build -o server`) and push it with
  `-b binary_buildpack`. No separate app is kept here.
