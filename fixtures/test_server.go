package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/manage/jolokia", metrics)
	log.Fatal(http.ListenAndServe(":3000", nil))
}

func metrics(w http.ResponseWriter, r *http.Request) {
	if u, p, ok := r.BasicAuth(); !ok {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, "Unauthorized")
		return
	} else if u != "admin" || p != "secret" {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, "Wrong credentials")
		return
	}

	w.Header().Add("Content-Type", "application/json")
	http.ServeFile(w, r, "metrics.json")
}
