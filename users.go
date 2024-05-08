package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type userparameters struct {
	// these tags indicate how the keys in the JSON should be mapped to the struct fields
	// the struct fields must be exported (start with a capital letter) if you want them parsed
	Email            string `json:"email"`
	Password         string `json:"password"`
	ExpiresInSeconds int    `json:"expires_in_seconds,omitempty"`
}

type User struct {
	Email             string    `json:"email"`
	ID                int       `json:"id"`
	Password          string    `json:"password"`
	RefreshToken      string    `json:"refresh_token"`
	RefreshExpiration time.Time `json:"refesh_expiration"`
	IsChirpyRed       bool      `json:"is_chirpy_red"`
}

func (cfg *apiConfig) PostUsers(w http.ResponseWriter, req *http.Request) {
	decoder := json.NewDecoder(req.Body)
	params := userparameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, 500, err.Error())
		return
	}

	pw, err := HashPassword(params.Password)
	if err != nil {
		respondWithError(w, 500, err.Error())
		return
	}

	newUser, err := cfg.DB.CreateUser(params.Email, pw)
	if err != nil {
		fmt.Printf("error: %s", err.Error())
		respondWithError(w, 500, err.Error())
		return
	}

	respondWithJSON(w, 201, newUser)
}

func (cfg *apiConfig) PostRevoke(w http.ResponseWriter, req *http.Request) {
	//Revoke the token in the database that matches the token that was passed in
	// the header of the request. Respond with a 200 status code.
	// --> revoke means to delete the request token i assume.... might need to check with Boots
	refreshtoken := req.Header.Get("Authorization")
	refreshtoken = strings.Split(refreshtoken, "Bearer ")[1]

	user, err := cfg.DB.GetUserbyRefresh(refreshtoken)
	if err != nil {
		fmt.Printf("could not find user id in db via refresh token: %s", err.Error())
		respondWithError(w, 401, "cannot get user via refresh token")
		return
	}
	if user.RefreshExpiration.Before(time.Now().UTC()) {
		fmt.Printf("Refresh Timer expired: %v", user.RefreshExpiration)
		respondWithError(w, 401, "Unauthorized - expired refresh token")
		return
	}
	err = cfg.DB.DelRefreshToken(user.ID)
	if err != nil {
		fmt.Printf("error revoking token: %s", err.Error())
		respondWithError(w, 503, "error revoking token")
		return
	}

}

func (cfg *apiConfig) PostRefresh(w http.ResponseWriter, req *http.Request) {
	refreshtoken := req.Header.Get("Authorization")
	refreshtoken = strings.Split(refreshtoken, "Bearer ")[1]

	user, err := cfg.DB.GetUserbyRefresh(refreshtoken)
	if err != nil {
		fmt.Printf("could not find user id in db via refresh token: %s", err.Error())
		respondWithError(w, 401, "cannot get user via refresh token")
		return
	}
	if user.RefreshExpiration.Before(time.Now().UTC()) {
		fmt.Printf("Refresh Timer expired: %v", user.RefreshExpiration)
		respondWithError(w, 401, "Unauthorized - expired refresh token")
		return
	}

	type response struct {
		Token string `json:"token"`
	}

	// generate new access tken and send response
	token, err := MakeJWT(user.ID, cfg.JWT_SECRET, time.Duration(60*60)*time.Second)
	if err != nil {
		fmt.Printf("cannot Make JWT: %v", err.Error())
		respondWithError(w, 401, "cannot Make JWT")
		return
	}

	respondWithJSON(w, 200, response{
		Token: token,
	})

}

func (cfg *apiConfig) ValidateHeader(req *http.Request) (userid int, err error) {
	token := req.Header.Get("Authorization")
	if token == "" {
		return 0, errors.New("cannot find Authorization Header")
	}
	token = strings.Split(token, "Bearer ")[1]

	suserid, err := ValidateJWT(token, cfg.JWT_SECRET)
	if err != nil {
		return 0, err
	}
	userid, err = strconv.Atoi(suserid)
	if err != nil {
		fmt.Printf("cannot convert userid: %v\n", err)
		return 0, err
	}
	return userid, nil
}

func (cfg *apiConfig) PutUsers(w http.ResponseWriter, req *http.Request) {

	id, err := cfg.ValidateHeader(req)
	if err != nil {
		fmt.Printf("error validating Auth Header: %s\n", err.Error())
		respondWithError(w, 401, "cannot get user by id:")
	}

	user, err := cfg.DB.GetUserbyID(id)
	if err != nil {
		fmt.Printf("cannot get user by id: %v\n", err)
		respondWithError(w, 401, "cannot get user by id:")
		return
	}
	decoder := json.NewDecoder(req.Body)
	params := userparameters{}
	err = decoder.Decode(&params)
	if err != nil {
		respondWithError(w, 500, err.Error())
		return
	}

	pw, err := HashPassword(params.Password)
	if err != nil {
		fmt.Printf("cannot hash password: %s", err.Error())
		respondWithError(w, http.StatusBadRequest, "cannot hash password")
		return
	}
	newUser := User{
		Email:             params.Email,
		Password:          pw,
		ID:                id,
		RefreshToken:      user.RefreshToken,
		RefreshExpiration: user.RefreshExpiration,
		IsChirpyRed:       user.IsChirpyRed,
	}
	err = cfg.DB.UpdateUser(id, newUser)
	if err != nil {
		fmt.Printf("could not update user. error: %v\n", err.Error())
		respondWithError(w, 500, err.Error())
		return
	}
	type response struct {
		User
	}
	responseUser := response{
		User: User{
			Email:             params.Email,
			ID:                user.ID,
			RefreshToken:      user.RefreshToken,
			RefreshExpiration: user.RefreshExpiration,
			IsChirpyRed:       user.IsChirpyRed,
		},
	}

	respondWithJSON(w, 200, responseUser)

}

func (cfg *apiConfig) PostLogin(w http.ResponseWriter, req *http.Request) {
	decoder := json.NewDecoder(req.Body)
	params := userparameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, 500, err.Error())
		return
	}
	user, err := cfg.DB.GetUserbyMail(params.Email)
	if err != nil {
		fmt.Printf("error cannot find user %s\n", err.Error())
		respondWithError(w, 401, "Unauthorized- cannot find user")
		return
	}
	err = CheckPasswordHash(params.Password, user.Password)
	if err != nil {
		fmt.Printf("error Password doesn't match: %s\n", err.Error())
		respondWithError(w, 401, "Unauthorized- passwords don't match")
		return
	}
	defaultExpiration := 60 * 60
	if params.ExpiresInSeconds == 0 {
		params.ExpiresInSeconds = defaultExpiration
	} else if params.ExpiresInSeconds > defaultExpiration {
		params.ExpiresInSeconds = defaultExpiration
	}
	token, err := MakeJWT(user.ID, cfg.JWT_SECRET, time.Duration(params.ExpiresInSeconds)*time.Second)
	if err != nil {
		fmt.Printf("cannot Make JWT: %v", err.Error())
		respondWithError(w, 401, "cannot Make JWT")
		return
	}
	//expires_in_seconds is an optional parameter. If it's specified by the client, use it as the expiration time. If it's not specified,
	// use a default expiration time of 24 hours. If the client specified a number over 24 hours, use 24 hours as the expiration time.

	refresh_token := GetRefreshToken()

	err = cfg.DB.SetRefreshToken(user.ID, refresh_token, time.Duration(time.Hour*24*60))
	if err != nil {
		respondWithError(w, 401, "cannot update refresh token in db")
		return
	}

	type returnUser struct {
		ID           int    `json:"id"`
		Email        string `json:"email"`
		Token        string `json:"token"`
		RefreshToken string `json:"refresh_token"`
		IsChirpyRed  bool   `json:"is_chirpy_red"`
	}

	rUser := returnUser{
		ID:           user.ID,
		Email:        user.Email,
		Token:        token,
		RefreshToken: refresh_token,
		IsChirpyRed:  user.IsChirpyRed,
	}

	respondWithJSON(w, 200, rUser)
}

// Generates a refresh_token
func GetRefreshToken() (refresh_token string) {
	c := 256
	b := make([]byte, c)
	inttoken, err := rand.Read(b)
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	stringtoken := strconv.Itoa(inttoken)
	src := []byte(stringtoken)
	refresh_token = hex.EncodeToString(src)

	return refresh_token
}

// ErrNoAuthHeaderIncluded -
var ErrNoAuthHeaderIncluded = errors.New("not auth header included in request")

// HashPassword -
func HashPassword(password string) (string, error) {
	dat, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(dat), nil
}

// CheckPasswordHash -
func CheckPasswordHash(password, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

// MakeJWT -
func MakeJWT(userID int, tokenSecret string, expiresIn time.Duration) (string, error) {
	signingKey := []byte(tokenSecret)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer:    "chirpy",
		IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(expiresIn)),
		Subject:   fmt.Sprintf("%d", userID),
	})
	return token.SignedString(signingKey)
}

// ValidateJWT -
func ValidateJWT(tokenString, tokenSecret string) (string, error) {
	claimsStruct := jwt.RegisteredClaims{}
	token, err := jwt.ParseWithClaims(
		tokenString,
		&claimsStruct,
		func(token *jwt.Token) (interface{}, error) { return []byte(tokenSecret), nil },
	)
	if err != nil {
		return "", err
	}

	userIDString, err := token.Claims.GetSubject()
	if err != nil {
		return "", err
	}

	issuer, err := token.Claims.GetIssuer()
	if err != nil {
		return "", err
	}
	if issuer != string("chirpy") {
		return "", errors.New("invalid issuer")
	}

	return userIDString, nil
}
