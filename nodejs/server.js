const http = require('http');
const port = process.env.PORT || 8080;
http.createServer((req, res) => {
  res.setHeader('Content-Type', 'application/json');
  res.end(JSON.stringify({ runtime: 'node ' + process.version, ok: true }));
}).listen(port, '0.0.0.0', () => console.log('listening on ' + port));
