package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type userparameters struct {
	// these tags indicate how the keys in the JSON should be mapped to the struct fields
	// the struct fields must be exported (start with a capital letter) if you want them parsed
	Email string `json:"email"`
}

func (cfg *apiConfig) PostUsers(w http.ResponseWriter, req *http.Request) {
	decoder := json.NewDecoder(req.Body)
	params := userparameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, 500, err.Error())
		return
	}

	newUser, err := cfg.DB.CreateUser(params.Email)
	if err != nil {
		fmt.Printf("error: %s", err.Error())
	}

	respondWithJSON(w, 201, newUser)
}
