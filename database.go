package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"
)

type DB struct {
	path string
	mux  *sync.RWMutex
}

type DBStructure struct {
	Chirps map[int]Chirp `json:"chirps"`
	Users  map[int]User  `json:"users"`
}

var ErrAlreadyExists = errors.New("already exists")
var ErrNotExist = errors.New("does not exist")

// NewDB creates a new database connection
// and creates the database file if it doesn't exist
func NewDB(path string) (*DB, error) {
	db := DB{
		path: path,
		mux:  &sync.RWMutex{},
	}

	db.ensureDB()

	return &db, nil

}

// CreateChirp creates a new chirp and saves it to disk
func (db *DB) CreateChirp(body string, userid int) (Chirp, error) {
	DBStructure, err := db.loadDB()
	if err != nil {
		return Chirp{}, err
	}
	id := len(DBStructure.Chirps)
	id++

	newChirp := Chirp{
		Body:     body,
		ID:       id,
		AuthorID: userid,
	}
	DBStructure.Chirps[id] = newChirp
	db.writeDB(DBStructure)
	return newChirp, nil
}

// GetChirps returns all chirps in the database
func (db *DB) GetChirps() ([]Chirp, error) {

	DBStructure, err := db.loadDB()
	if err != nil {
		return []Chirp{}, err
	}
	Chirps := []Chirp{}
	for _, chirp := range DBStructure.Chirps {
		Chirps = append(Chirps, chirp)
	}

	return Chirps, nil

}

// ensureDB creates a new database file if it doesn't exist
func (db *DB) ensureDB() error {
	_, err := os.ReadFile(db.path)
	if os.IsNotExist(err) {
		ChirpsDB := DBStructure{
			Chirps: make(map[int]Chirp),
			Users:  make(map[int]User),
		}
		db.writeDB(ChirpsDB)
	}
	return nil
}

// loadDB reads the database file into memory
func (db *DB) loadDB() (DBStructure, error) {
	db.ensureDB()
	dat, err := os.ReadFile(db.path)
	if err != nil {
		return DBStructure{}, err
	}
	db.mux.RLock()
	defer db.mux.RUnlock()
	newStructure := DBStructure{}
	err = json.Unmarshal(dat, &newStructure)
	if err != nil {
		return DBStructure{}, err
	}

	return newStructure, nil
}

// writeDB writes the database file to disk
func (db *DB) writeDB(dbStructure DBStructure) error {
	bytedb, err := json.Marshal(dbStructure)
	if err != nil {
		return err
	}
	db.mux.Lock()
	defer db.mux.Unlock()
	err = os.WriteFile(db.path, bytedb, 0644)
	if err != nil {
		return err
	}

	return nil

}

func (db *DB) CreateUser(email string, password string) (User, error) {
	DBStructure, err := db.loadDB()
	if err != nil {
		return User{}, err
	}
	//if _, err := db.GetUser(email); !errors.Is(err, ErrNotExist) {
	//	return User{}, ErrAlreadyExists
	//}
	id := len(DBStructure.Users)
	id++

	newUser := User{
		Email:    email,
		ID:       id,
		Password: password,
	}
	DBStructure.Users[id] = newUser
	db.writeDB(DBStructure)
	return newUser, nil
}

// GetUser returns a user from the database
func (db *DB) GetUserbyID(id int) (User, error) {
	DBStructure, err := db.loadDB()
	if err != nil {
		return User{}, err
	}
	user, exists := DBStructure.Users[id]
	if !exists {
		err := errors.New("user not found in DB")
		return User{}, err
	}

	return user, nil

}

// GetUser returns a user from the database
func (db *DB) GetUserbyMail(email string) (User, error) {
	DBStructure, err := db.loadDB()
	if err != nil {
		return User{}, err
	}
	var foundUser User
	for _, user := range DBStructure.Users {
		if user.Email == email {
			foundUser = user
			return foundUser, nil
		}
	}
	return User{}, errors.New("User does not exist in DB")

}

// GetUser returns a user from the database
func (db *DB) GetUserbyRefresh(refreshtoken string) (User, error) {
	DBStructure, err := db.loadDB()
	if err != nil {
		return User{}, err
	}
	var foundUser User
	for _, user := range DBStructure.Users {
		if user.RefreshToken == refreshtoken {
			foundUser = user
			return foundUser, nil
		}
	}
	return User{}, errors.New("refresh Token does not exist in DB")

}

// UpdateUser updates a user in the database
func (db *DB) UpdateUser(id int, newUser User) error {
	DBStructure, err := db.loadDB()
	if err != nil {
		return err
	}

	user, exists := DBStructure.Users[id]
	if !exists {
		err := errors.New("user not found in DB")
		return err
	}
	fmt.Printf("updating user with mail: %s\n", user.Email)
	DBStructure.Users[id] = newUser
	db.writeDB(DBStructure)
	return nil
}

// Set a new refreshToken for a User
func (db *DB) SetRefreshToken(id int, newtoken string, expiresIn time.Duration) error {
	DBStructure, err := db.loadDB()
	if err != nil {
		return err
	}

	user, exists := DBStructure.Users[id]
	if !exists {
		err := errors.New("user not found in DB")
		return err
	}
	fmt.Printf("setting new refresh_token for user with mail: %s\n", user.Email)
	user.RefreshToken = newtoken
	user.RefreshExpiration = time.Now().UTC().Add(expiresIn)

	DBStructure.Users[id] = user
	db.writeDB(DBStructure)
	return nil
}

// Delete refreshToken for a User
func (db *DB) DelRefreshToken(id int) error {
	DBStructure, err := db.loadDB()
	if err != nil {
		return err
	}

	user, exists := DBStructure.Users[id]
	if !exists {
		err := errors.New("user not found in DB")
		return err
	}
	fmt.Printf("revoking refresh_token for user with mail: %s\n", user.Email)
	user.RefreshToken = ""
	user.RefreshExpiration = time.Now().UTC()

	DBStructure.Users[id] = user
	db.writeDB(DBStructure)
	return nil
}
