package main

import (
	"encoding/json"
	"errors"
	"os"
	"sync"
)

type DB struct {
	path string
	mux  *sync.RWMutex
}

type DBStructure struct {
	Chirps map[int]Chirp `json:"chirps"`
	Users  map[int]User  `json:"users"`
}

type User struct {
	Email string `json:"email"`
	ID    int    `json:"id"`
}

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
func (db *DB) CreateChirp(body string) (Chirp, error) {
	DBStructure, err := db.loadDB()
	if err != nil {
		return Chirp{}, err
	}
	id := len(DBStructure.Chirps)
	id++

	newChirp := Chirp{
		Body: body,
		ID:   id,
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
		errors.New("error marshalling new db")
		return err
	}
	db.mux.Lock()
	defer db.mux.Unlock()
	err = os.WriteFile(db.path, bytedb, 0644)
	if err != nil {
		errors.New("error writing db")
		return err
	}

	return nil

}

func (db *DB) CreateUser(email string) (User, error) {
	DBStructure, err := db.loadDB()
	if err != nil {
		return User{}, err
	}
	id := len(DBStructure.Users)
	id++

	newUser := User{
		Email: email,
		ID:    id,
	}
	DBStructure.Users[id] = newUser
	db.writeDB(DBStructure)
	return newUser, nil
}
