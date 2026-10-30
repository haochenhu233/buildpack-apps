// viper-sim: a minimal CF ROUTE SERVICE that mimics viper-auth.
//
// gorouter forwards the request HERE first, with:
//   x-cf-forwarded-url                      -> the original app URL to forward to after auth
//   x-cf-proxy-signature / x-cf-proxy-metadata -> MUST be echoed back so gorouter delivers
//                                               the forwarded request to the app (no loop)
//
// Behavior (mimics viper):
//   - no/invalid X-Viper-Auth  -> 403  (like viper-ia)
//   - valid X-Viper-Auth       -> forward to x-cf-forwarded-url -> return app response (200)
//
// The 502 you're chasing happens BEFORE this app runs: if viper-sim's route resolves to a
// PRIVATE IP, gorouter's egress_blocklist (routing 0.372.0+) refuses the dial -> 502.
const http = require('http');
const https = require('https');
const { URL } = require('url');
const port = process.env.PORT || 8080;

http.createServer((req, res) => {
  const fwd = req.headers['x-cf-forwarded-url'];
  if (!fwd) {
    res.statusCode = 400;
    return res.end('viper-sim: no X-CF-Forwarded-Url (not invoked as a route service)\n');
  }

  // --- "auth" step ---
  if (req.headers['x-viper-auth'] !== 'let-me-in') {
    res.statusCode = 403;
    return res.end('viper-sim: 403 Forbidden (missing/invalid X-Viper-Auth)\n');
  }

  // --- forward to the real app via x-cf-forwarded-url ---
  const u = new URL(fwd);
  const lib = u.protocol === 'https:' ? https : http;
  const headers = Object.assign({}, req.headers, { 'x-viper-sim': 'forwarded' }); // keeps x-cf-proxy-* (required)
  delete headers.host; // let the client use the forwarded URL's host

  const preq = lib.request(fwd, { method: req.method, headers }, (pres) => {
    res.writeHead(pres.statusCode, pres.headers);
    pres.pipe(res);
  });
  preq.on('error', (e) => {
    res.statusCode = 502;
    res.end('viper-sim: upstream error: ' + e.message + '\n');
  });
  req.pipe(preq);
}).listen(port, '0.0.0.0', () => console.log('viper-sim (route service) listening on ' + port));
