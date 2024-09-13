package database

import "errors"

// create new Chirp and write new DB.data to disk
func (db *DB) CreateChirp(msg string, authorID int) (Chirp, error) {
	db.mux.Lock()
	defer db.mux.Unlock()
	newChirp := Chirp{
		ID:       len(db.Data.Chirps) + 1,
		Body:     msg,
		AuthorID: authorID,
	}
	db.Data.Chirps[newChirp.ID] = newChirp
	err := db.writeDBtoDisk()
	if err != nil {
		return Chirp{}, err
	}
	return newChirp, nil
}

func (db *DB) DeleteChirp(authorID, chirpID int) error {
	for id, chirp := range db.Data.Chirps {
		if chirp.AuthorID != authorID {
			return errors.New("cannot delete chirps of others")
		}
		if id == chirpID {
			db.Data.Chirps[id] = Chirp{}
			return nil
		}
	}
	return errors.New("chirp not found")
}

func (db *DB) GetChirpByChirpId(chirpID int) (Chirp, error) {
	for id, chirp := range db.Data.Chirps {
		if id != chirpID {
			return chirp, nil
		}
	}
	return Chirp{}, ErrNotExist
}
