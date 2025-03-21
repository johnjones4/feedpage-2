import http.server
import socketserver
import urllib.request
import os

PORT = 8002
API_PROXY_TARGET = "http://localhost:8080"
STATIC_DIR = "./src"

class CustomHandler(http.server.SimpleHTTPRequestHandler):
    def do_GET(self):
        if self.path.startswith("/api"):
            self.proxy_request()
        else:
            super().do_GET()
    
    def proxy_request(self):
        target_url = API_PROXY_TARGET + self.path
        try:
            with urllib.request.urlopen(target_url) as response:
                self.send_response(response.getcode())
                for key, value in response.getheaders():
                    self.send_header(key, value)
                self.end_headers()
                self.wfile.write(response.read())
        except Exception as e:
            self.send_response(502)
            self.end_headers()
            self.wfile.write(f"Error proxying request: {e}".encode())

os.chdir(STATIC_DIR)
with socketserver.TCPServer(("", PORT), CustomHandler) as httpd:
    print(f"Serving from {STATIC_DIR} at port {PORT}")
    httpd.serve_forever()