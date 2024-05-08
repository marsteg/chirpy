package main

import (
	"encoding/json"
	"net/http"
	"strings"
)

type PolkaPost struct {
	// these tags indicate how the keys in the JSON should be mapped to the struct fields
	Event     string    `json:"event"`
	PolkaData PolkaData `json:"data"`
}
type PolkaData struct {
	UserID int `json:"user_id"`
}

func (cfg *apiConfig) PostPolkaWebhooks(w http.ResponseWriter, req *http.Request) {
	key := req.Header.Get("Authorization")
	if key == "" {
		respondWithError(w, 401, "Unauthorized - cannot find Authorization Header")
		return
	}
	key = strings.Split(key, "ApiKey ")[1]
	if key != cfg.PolkaAPIKey {
		respondWithError(w, 401, "Unauthorized")
		return
	}
	decoder := json.NewDecoder(req.Body)
	params := PolkaPost{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, 500, err.Error())
		return
	}

	if params.Event != "user.upgraded" {
		respondWithError(w, 200, "user not upgraded")
		return
	}

	user, err := cfg.DB.GetUserbyID(params.PolkaData.UserID)
	if err != nil {
		respondWithError(w, 404, err.Error())
		return
	}

	user.IsChirpyRed = true
	cfg.DB.UpdateUser(user.ID, user)

	respondWithJSON(w, 200, "User upgraded")
}
