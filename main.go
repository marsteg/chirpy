package main

import (
	"fmt"
	"net/http"
)

type apiConfig struct {
	fileserverHits int
}

func main() {
	mux := http.NewServeMux()
	apiCfg := apiConfig{
		fileserverHits: 0,
	}
	mux.Handle("/app/*", apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer((http.Dir("."))))))

	mux.HandleFunc("/assets/logo.png", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "assets/logo.png")
	})
	mux.HandleFunc(("/healthz"), func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK\n"))
	})
	mux.HandleFunc(("/metrics"), func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Hits: %v\n", apiCfg.fileserverHits)
	})
	mux.HandleFunc(("/reset"), func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		apiCfg.fileserverHits = 0
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

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits++
		next.ServeHTTP(w, r)
	})
}
