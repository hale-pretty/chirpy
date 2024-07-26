package database

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

type Chirp struct {
	ID   int    `json:"id"`
	Body string `json:"body"`
}

type DB struct {
	path string
	mux  *sync.RWMutex
	Data *DbData `json:"data"`
}

type DbData struct {
	Chirps map[int]Chirp `json:"chirps"`
}

// NewDB creates a new database connection
// and creates the database file if it doesn't exist
func NewDB(path string) (*DB, error) {
	chirpsMap := make(map[int]Chirp)
	db := &DB{
		path: path,
		mux:  &sync.RWMutex{},
		Data: &DbData{
			Chirps: chirpsMap,
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

// create new Chirp and write new DB.data to disk
func (db *DB) CreateChirp(msg string) (Chirp, error) {
	db.mux.Lock()
	defer db.mux.Unlock()
	newChirp := Chirp{
		ID:   len(db.Data.Chirps) + 1,
		Body: msg,
	}
	db.Data.Chirps[newChirp.ID] = newChirp
	err := db.writeDBtoDisk()
	if err != nil {
		return Chirp{}, err
	}
	return newChirp, nil
}
