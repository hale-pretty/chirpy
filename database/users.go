package database

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
)

var ErrNotExist = errors.New("resources not found")

// create new User and write new DB.data to disk
func (db *DB) CreateUser(email string, password string) (UserWithoutPW, error) {
	db.mux.Lock()
	defer db.mux.Unlock()
	hashedPassword, err1 := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err1 != nil {
		return UserWithoutPW{}, err1
	}
	newUser := User{
		ID:          len(db.Data.Users) + 1,
		Password:    hashedPassword,
		Email:       email,
		IsChirpyRed: false,
	}
	db.Data.Users[newUser.ID] = newUser
	err2 := db.writeDBtoDisk()
	if err2 != nil {
		return UserWithoutPW{}, err2
	}
	newUserWoPW := UserWithoutPW{
		ID:          newUser.ID,
		Email:       newUser.Email,
		IsChirpyRed: newUser.IsChirpyRed,
	}
	return newUserWoPW, nil
}

// Login user
func (db *DB) IdentifyUser(password string) (UserWithoutPW, bool) {
	usersMap := db.Data.Users
	for _, user := range usersMap {
		err := bcrypt.CompareHashAndPassword(user.Password, []byte(password))
		if err == nil {
			return UserWithoutPW{
				ID:          user.ID,
				Email:       user.Email,
				IsChirpyRed: user.IsChirpyRed,
			}, true
		}
	}
	return UserWithoutPW{}, false
}

// Update user info
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
				ID:          userID,
				Email:       email,
				IsChirpyRed: user.IsChirpyRed,
			}, true
		}
	}
	return UserWithoutPW{}, false
}

// Login user and save refresh token to database
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

func (db *DB) IsChirpyRed(userID int) error {
	for id, user := range db.Data.Users {
		if id == userID {
			db.Data.Users[id] = User{
				ID:           user.ID,
				Email:        user.Email,
				Password:     user.Password,
				RefreshToken: user.RefreshToken,
				IsChirpyRed:  true,
			}
			err := db.writeDBtoDisk()
			if err != nil {
				panic(err)
			}
			return nil
		}
	}
	return ErrNotExist
}
