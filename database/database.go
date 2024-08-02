package database

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"golang.org/x/crypto/bcrypt"
)

type Chirp struct {
	ID   int    `json:"id"`
	Body string `json:"body"`
}

type User struct {
	ID           int    `json:"id"`
	Email        string `json:"email"`
	Password     []byte `json:"password"`
	RefreshToken string `json:"refresh_token"`
}

type UserWithoutPW struct {
	ID    int    `json:"id"`
	Email string `json:"email"`
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

// create new User and write new DB.data to disk
func (db *DB) CreateUser(email string, password string) (UserWithoutPW, error) {
	db.mux.Lock()
	defer db.mux.Unlock()
	hashedPassword, err1 := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err1 != nil {
		return UserWithoutPW{}, err1
	}
	newUser := User{
		ID:       len(db.Data.Users) + 1,
		Password: hashedPassword,
		Email:    email,
	}
	db.Data.Users[newUser.ID] = newUser
	err2 := db.writeDBtoDisk()
	if err2 != nil {
		return UserWithoutPW{}, err2
	}
	newUserWoPW := UserWithoutPW{
		ID:    newUser.ID,
		Email: newUser.Email,
	}
	return newUserWoPW, nil
}

func (db *DB) IdentifyUser(password string) (UserWithoutPW, bool) {
	usersMap := db.Data.Users
	for _, user := range usersMap {
		err := bcrypt.CompareHashAndPassword(user.Password, []byte(password))
		if err == nil {
			return UserWithoutPW{
				ID:    user.ID,
				Email: user.Email,
			}, true
		}
	}
	return UserWithoutPW{}, false
}

func (db *DB) UpdateUser(userID int, email, password string) (UserWithoutPW, bool) {
	for id, user := range db.Data.Users {
		if id == userID {
			hashedPassword, err1 := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
			if err1 != nil {
				return UserWithoutPW{}, false
			}
			db.Data.Users[id] = User{
				ID:           user.ID,
				Email:        email,
				Password:     hashedPassword,
				RefreshToken: user.RefreshToken,
			}

			err := db.writeDBtoDisk()
			if err != nil {
				panic(err)
			}
			return UserWithoutPW{
				ID:    userID,
				Email: email,
			}, true
		}
	}
	return UserWithoutPW{}, false
}

func (db *DB) LoginUser(userID int, refreshToken string) {
	for id, user := range db.Data.Users {
		if id == userID {
			db.Data.Users[id] = User{
				ID:           user.ID,
				Email:        user.Email,
				Password:     user.Password,
				RefreshToken: refreshToken,
			}
			err := db.writeDBtoDisk()
			if err != nil {
				panic(err)
			}
		}
	}
}

func (db *DB) RefreshNewAccessToken(refreshToken string) (int, bool) {
	for _, user := range db.Data.Users {
		if user.RefreshToken == refreshToken {
			return user.ID, true
		}
	}
	return 0, false
}

func (db *DB) RevokeRefreshToken(refreshToken string) bool {
	for id, user := range db.Data.Users {
		if user.RefreshToken == refreshToken {
			db.Data.Users[id] = User{
				ID:           user.ID,
				Email:        user.Email,
				Password:     user.Password,
				RefreshToken: "",
			}
			err := db.writeDBtoDisk()
			if err != nil {
				panic(err)
			}
			return true
		}
	}
	return false
}
