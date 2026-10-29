package main

import (
	"fmt"
	"net/http"
	"os"
)

// The "protected" app. A route service (viper-sim) is bound in front of its route.
func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// If this prints, the request made it THROUGH the route service to the app.
		fmt.Fprintf(w, "TARGET OK — reached the app. via-route-service=%q\n",
			r.Header.Get("X-Viper-Sim"))
	})
	fmt.Println("target-app listening on " + port)
	_ = http.ListenAndServe("0.0.0.0:"+port, nil)
}
