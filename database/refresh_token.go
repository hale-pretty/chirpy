package database

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
