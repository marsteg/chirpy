package main

import (
	"net/http"
)

func main() {
	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer((http.Dir("."))))
	mux.HandleFunc("/assets/logo.png", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "chirpy.png")
	})
	corsMux := middlewareCors(mux)
	s := http.Server{
		Addr:    ":8080",
		Handler: corsMux,
	}
	s.ListenAndServe()
}

func middlewareCors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "*")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func logo(w http.ResponseWriter, r *http.Request) {

}
