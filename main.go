package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
)

type apiConfig struct {
	fileserverHits int
	DB             *DB
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits++
		next.ServeHTTP(w, r)
	})
}

const database string = "./database.json"

func main() {

	const filepathRoot = "."
	const port = "8080"
	var apiCfg apiConfig
	var err error
	apiCfg.DB, err = NewDB(database)

	dbg := flag.Bool("debug", false, "Enable debug mode")
	flag.Parse()
	if *dbg {
		err := os.Remove(database)
		if err != nil {
			log.Fatal(err)
		}
	}

	if err != nil {
		fmt.Printf("Error when loading DB File: %s", err.Error())
	}

	mux := http.NewServeMux()
	handler := http.StripPrefix("/app", http.FileServer(http.Dir(filepathRoot)))
	mux.Handle("/", apiCfg.middlewareMetricsInc(handler))
	mux.HandleFunc("GET /api/healthz", healthz)
	mux.HandleFunc("GET /admin/metrics", apiCfg.metrics)
	mux.HandleFunc("/api/reset", apiCfg.reset)
	mux.HandleFunc("POST /api/chirps", apiCfg.PostChirps)
	mux.HandleFunc("GET /api/chirps/{chirpID}", apiCfg.GetChirpID)
	mux.HandleFunc("GET /api/chirps", apiCfg.GetChirps)
	mux.HandleFunc("POST /api/users", apiCfg.PostUsers)
	mux.HandleFunc("POST /healthz", posthealthz)

	s := http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}
	log.Printf("Serving files from %s on port: %s\n", filepathRoot, port)
	log.Fatal(s.ListenAndServe())

}

func healthz(w http.ResponseWriter, req *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK\n"))
}

func posthealthz(w http.ResponseWriter, req *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusMethodNotAllowed)
	w.Write([]byte("Method not allowed\n"))
}
