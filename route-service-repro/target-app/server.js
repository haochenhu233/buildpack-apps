// The "protected" app. A route service (viper-sim) is bound in front of its route.
const http = require('http');
const port = process.env.PORT || 8080;

http.createServer((req, res) => {
  // If this responds, the request made it THROUGH the route service to the app.
  res.end('TARGET OK — reached the app. via-route-service="' + (req.headers['x-viper-sim'] || '') + '"\n');
}).listen(port, '0.0.0.0', () => console.log('target-app listening on ' + port));
