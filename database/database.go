package database

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

type Chirp struct {
	ID       int    `json:"id"`
	Body     string `json:"body"`
	AuthorID int    `json:"author_id"`
}

type User struct {
	ID           int    `json:"id"`
	Email        string `json:"email"`
	Password     []byte `json:"password"`
	RefreshToken string `json:"refresh_token"`
	IsChirpyRed  bool   `json:"is_chirpy_red"`
}

type UserWithoutPW struct {
	ID          int    `json:"id"`
	Email       string `json:"email"`
	IsChirpyRed bool   `json:"is_chirpy_red"`
}

type DB struct {
	path string
	mux  *sync.RWMutex
	Data *DbData `json:"data"`
}

type DbData struct {
	Chirps map[int]Chirp `json:"chirps"`
	Users  map[int]User  `json:"users"`
}

// NewDB creates a new database connection
// and creates the database file if it doesn't exist
func NewDB(path string) (*DB, error) {
	chirpsMap := make(map[int]Chirp)
	usersMap := make(map[int]User)
	db := &DB{
		path: path,
		mux:  &sync.RWMutex{},
		Data: &DbData{
			Chirps: chirpsMap,
			Users:  usersMap,
		},
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		err := db.writeDBtoDisk()
		if err != nil {
			return nil, fmt.Errorf("cannot create DB file: %w", err)
		}
	}
	err := db.loadDB()
	if err != nil {
		return nil, fmt.Errorf("cannot load DB: %w", err)
	}
	return db, nil
}

// write DB.data to Disk
func (db *DB) writeDBtoDisk() error {
	jsonData, err1 := json.MarshalIndent(db.Data, "", " ")
	if err1 != nil {
		return err1
	}
	err2 := os.WriteFile(db.path, []byte(jsonData), 0644)
	if err2 != nil {
		return err2
	}
	return nil
}

// load DB.data to memory
func (db *DB) loadDB() error {
	data, err1 := os.ReadFile(db.path)
	if err1 != nil {
		return err1
	}
	err2 := json.Unmarshal(data, db.Data)
	if err2 != nil {
		return err2
	}
	return nil
}
