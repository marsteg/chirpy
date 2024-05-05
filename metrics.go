package main

import (
	"fmt"
	"net/http"
)

func (cfg *apiConfig) metrics(w http.ResponseWriter, req *http.Request) {
	//reply := fmt.Sprintf("Hits: %d\n", cfg.fileserverHits)
	//text := fmt.Sprintf("<html><body><h1>Welcome, Chirpy Admin</h1><p>Chirpy has been visited %d times!</p></body></html>", cfg.fileserverHits)
	reply := fmt.Sprintf(`
	<html>
	
	<body>
		<h1>Welcome, Chirpy Admin</h1>
		<p>Chirpy has been visited %d times!</p>
	</body>
	
	</html>
		`, cfg.fileserverHits)
	w.Write([]byte(reply))
}
func (cfg *apiConfig) reset(w http.ResponseWriter, req *http.Request) {
	cfg.fileserverHits = 0
	w.Write([]byte("Hits resetted\n"))
}
