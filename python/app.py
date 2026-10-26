import os, sys, json
from http.server import BaseHTTPRequestHandler, HTTPServer

class H(BaseHTTPRequestHandler):
    def do_GET(self):
        self.send_response(200)
        self.send_header("Content-Type", "application/json")
        self.end_headers()
        self.wfile.write(json.dumps({"runtime": "python " + sys.version.split()[0], "ok": True}).encode())
    def log_message(self, *a):
        pass

port = int(os.environ.get("PORT", 8080))
HTTPServer(("0.0.0.0", port), H).serve_forever()
