using System.Runtime.InteropServices;

var port = Environment.GetEnvironmentVariable("PORT") ?? "8080";
var builder = WebApplication.CreateBuilder(args);
builder.WebHost.UseUrls($"http://0.0.0.0:{port}");
var app = builder.Build();

app.MapGet("/", () =>
{
    var gc = GC.GetGCMemoryInfo();
    return Results.Json(new
    {
        framework = RuntimeInformation.FrameworkDescription,      // e.g. ".NET 10.0.11"
        // What .NET detected as the CONTAINER memory limit. The cgroup-v2 bug (runtime
        // < 10.0.11 on a v2-only host like noble) makes this the HOST's RAM instead of
        // the app's quota -> GC over-allocates -> OOM. Compare this to the manifest memory.
        detectedMemoryLimitBytes = gc.TotalAvailableMemoryBytes,
        cgroup = File.Exists("/sys/fs/cgroup/memory.max") ? "v2" : "v1/other",
        ok = true
    });
});

app.Run();
