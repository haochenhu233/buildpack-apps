package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
)

// viper-sim: a minimal CF ROUTE SERVICE that mimics viper-auth.
//
// When a route service is bound to an app's route, gorouter forwards the request
// HERE first, with these headers:
//   X-CF-Forwarded-Url  -> the original app URL to forward to after auth
//   X-CF-Proxy-Signature / X-CF-Proxy-Metadata -> MUST be echoed back so gorouter
//                          delivers the forwarded request to the app (not loop here)
//
// Behavior (mimics what you saw with viper):
//   - no/!ok auth header            -> 403  (like viper-ia returned)
//   - auth ok                       -> forward to X-CF-Forwarded-Url -> return app response (200)
// The 502 you're chasing happens BEFORE this app is even reached: if viper-sim's
// route resolves to a PRIVATE IP, gorouter's egress_blocklist (routing 0.372.0+)
// refuses the dial and returns 502 -> this app never runs.
func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fwd := r.Header.Get("X-Cf-Forwarded-Url")
		if fwd == "" {
			http.Error(w, "viper-sim: no X-CF-Forwarded-Url (not invoked as a route service)", http.StatusBadRequest)
			return
		}

		// --- "auth" step (mimic viper) ---
		if r.Header.Get("X-Viper-Auth") != "let-me-in" {
			http.Error(w, "viper-sim: 403 Forbidden (missing/invalid X-Viper-Auth)\n", http.StatusForbidden)
			return
		}

		// --- forward to the real app via X-CF-Forwarded-Url ---
		req, err := http.NewRequest(r.Method, fwd, r.Body)
		if err != nil {
			http.Error(w, "viper-sim: build req: "+err.Error(), http.StatusInternalServerError)
			return
		}
		req.Header = r.Header.Clone() // carries X-CF-Proxy-Signature / X-CF-Proxy-Metadata (required)
		req.Header.Set("X-Viper-Sim", "forwarded")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			http.Error(w, "viper-sim: upstream error: "+err.Error(), http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()
		for k, vv := range resp.Header {
			for _, v := range vv {
				w.Header().Add(k, v)
			}
		}
		w.WriteHeader(resp.StatusCode)
		_, _ = io.Copy(w, resp.Body)
	})

	fmt.Println("viper-sim (route service) listening on " + port)
	_ = http.ListenAndServe("0.0.0.0:"+port, nil)
}
