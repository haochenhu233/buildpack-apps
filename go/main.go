package main

import (
	"encoding/json"
	"net/http"
	"os"
	"runtime"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"runtime": runtime.Version(), "ok": true})
	})
	_ = http.ListenAndServe("0.0.0.0:"+port, nil)
}
