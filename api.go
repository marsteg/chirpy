package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"
)

type parameters struct {
	// these tags indicate how the keys in the JSON should be mapped to the struct fields
	// the struct fields must be exported (start with a capital letter) if you want them parsed
	Body string `json:"body"`
}

type Chirp struct {
	Body     string `json:"body"`
	ID       int    `json:"id"`
	AuthorID int    `json:"author_id"`
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
	/* Update the GET /api/chirps endpoint. It should accept an optional query parameter called sort. It can have 2 possible values:

	asc - Sort the chirps in the response by id in ascending order
	desc - Sort the chirps in the response by id in descending order
	asc is the default if no sort query parameter is provided.
	*/
	ssort := req.URL.Query().Get("sort")
	if ssort != "asc" && ssort != "desc" {
		ssort = "asc"
	}
	// sort by id in ascending order

	author_id := req.URL.Query().Get("author_id")
	if author_id != "" {
		author_id_int, err := strconv.Atoi(author_id)
		if err != nil {
			respondWithError(w, 400, "author_id could not be parsed")
		}
		Chirps := []Chirp{}

		for _, chirp := range db.Chirps {
			if chirp.AuthorID == author_id_int {
				Chirps = append(Chirps, chirp)
			}
		}
		Chirps = SortingChirps(Chirps, ssort)
		respondWithJSON(w, 200, Chirps)
	}

	Chirps := []Chirp{}
	for _, chirp := range db.Chirps {
		Chirps = append(Chirps, chirp)
	}
	Chirps = SortingChirps(Chirps, ssort)
	respondWithJSON(w, 200, Chirps)
}

func SortingChirps(Chirps []Chirp, ssort string) []Chirp {
	if ssort == "desc" {
		sort.Slice(Chirps, func(i, j int) bool {
			return Chirps[i].ID > Chirps[j].ID
		})
	} else {
		sort.Slice(Chirps, func(i, j int) bool {
			return Chirps[i].ID < Chirps[j].ID
		})
	}
	return Chirps
}

// deletes a Chirp from the database
func (cfg *apiConfig) DelChirpID(w http.ResponseWriter, req *http.Request) {
	userid, err := cfg.ValidateHeader(req)
	if err != nil {
		respondWithError(w, 401, "Unauthorized")
		return
	}
	db, err := cfg.DB.loadDB()
	if err != nil {
		respondWithError(w, 500, "cannot load db")
	}
	sid := req.PathValue("chirpID")
	chirpid, err := strconv.Atoi(sid)
	if err != nil {
		respondWithError(w, 400, "Chirp id could not be parsed")
	}
	chirp, exists := db.Chirps[chirpid]
	if !exists {
		respondWithError(w, 404, "Chirp does not exist")
		return
	}
	user, err := cfg.DB.GetUserbyID(userid)
	if err != nil {
		respondWithError(w, 500, "cannot get user by id")
		return
	}

	if chirp.AuthorID != user.ID {
		respondWithError(w, 403, "Unauthorized - different user")
		return
	}

	delete(db.Chirps, chirpid)
	cfg.DB.writeDB(db)
	respondWithJSON(w, 200, "Chirp deleted")
}

func (cfg *apiConfig) PostChirps(w http.ResponseWriter, req *http.Request) {
	userid, err := cfg.ValidateHeader(req)
	if err != nil {
		respondWithError(w, 401, "Unauthorized")
		return
	}

	decoder := json.NewDecoder(req.Body)
	params := parameters{}
	err = decoder.Decode(&params)
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

	validChirp, err := cfg.DB.CreateChirp(cleaned_body, userid)
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
