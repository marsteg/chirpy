package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
)

type parameters struct {
	// these tags indicate how the keys in the JSON should be mapped to the struct fields
	// the struct fields must be exported (start with a capital letter) if you want them parsed
	Body string `json:"body"`
}

type Chirp struct {
	Body string `json:"body"`
	ID   int    `json:"id"`
}

func (cfg *apiConfig) GetChirpID(w http.ResponseWriter, req *http.Request) {
	db, err := cfg.DB.loadDB()
	if err != nil {
		respondWithError(w, 500, "cannot load db")
	}
	id := req.PathValue("chirpID")
	iid, err := strconv.Atoi(id)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "id could not be parsed")
	}

	Chirp, exists := db.Chirps[iid]
	// If the key exists
	if !exists {
		respondWithError(w, 404, "404 page not found")
	}

	respondWithJSON(w, 200, Chirp)
}

func (cfg *apiConfig) GetChirps(w http.ResponseWriter, req *http.Request) {
	db, err := cfg.DB.loadDB()
	if err != nil {
		respondWithError(w, 500, "cannot load db")
	}
	Chirps := []Chirp{}
	for _, chirp := range db.Chirps {
		Chirps = append(Chirps, chirp)
	}

	respondWithJSON(w, 200, Chirps)
}

func (cfg *apiConfig) PostChirps(w http.ResponseWriter, req *http.Request) {
	decoder := json.NewDecoder(req.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		msg := "cannot decode json"
		respondWithError(w, 500, msg)
		return
	}
	if len(params.Body) > 140 {
		msg := "Chirp is too long"
		respondWithError(w, 400, msg)
		return
	}
	cleaned_body := replaceProfane(params.Body)

	validChirp, err := cfg.DB.CreateChirp(cleaned_body)
	if err != nil {
		fmt.Printf("error: %s", err.Error())
	}

	respondWithJSON(w, 201, validChirp)
}

func replaceProfane(body string) (cleaned_body string) {
	badWords := map[string]struct{}{
		"kerfuffle": {},
		"sharbert":  {},
		"fornax":    {},
	}
	words := strings.Split(body, " ")

	for i, word := range words {
		loweredWord := strings.ToLower(word)
		if _, ok := badWords[loweredWord]; ok {
			words[i] = "****"
		}
	}

	cleaned_body = strings.Join(words, " ")
	return cleaned_body
}

func respondWithError(w http.ResponseWriter, code int, msg string) {
	type error struct {
		Error string `json:"error"`
	}
	Error := error{
		Error: msg,
	}
	respondWithJSON(w, code, Error)

}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {

	w.Header().Set("Content-Type", "application/json")
	dat, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}
	w.WriteHeader(code)
	w.Write(dat)
}
