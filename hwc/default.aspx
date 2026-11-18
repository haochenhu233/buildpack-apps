<%@ Page Language="C#" %>
<%
    // Smoke test for the hwc_buildpack on a Windows cell. Reports the .NET Framework
    // runtime the Windows stemcell provides (HWC does not install a runtime itself).
    Response.ContentType = "application/json";
    string framework;
    try
    {
        framework = System.Runtime.InteropServices.RuntimeInformation.FrameworkDescription;
    }
    catch
    {
        framework = ".NET Framework (version unavailable on this runtime)";
    }
    Response.Write(
        "{\"runtime\":\"" + framework + "\"," +
        "\"clr\":\"" + System.Environment.Version + "\"," +
        "\"host\":\"HWC\",\"ok\":true}");
%>
