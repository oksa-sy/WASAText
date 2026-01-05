package database

import "errors"

// controllo se un id utente è effettivamente presente nel database
func (db *appdbimpl) SearchUser(requestId string) (bool, error) {
	var id string
	err := db.c.QueryRow(`SELECT id FROM users WHERE id = ?`, requestId).Scan(&id)
	if err != nil {
		return false, err
	}
	return true, nil
}

// creazione utente, controllo se esiste gia, se no lo creo, restituisco il suo id
func (db *appdbimpl) CreateUser(name string) (string, error) {
	// se l'utente esiste gia, ritorno il suo id
	var id string
	err := db.c.QueryRow(`SELECT id FROM users WHERE username = ?`, name).Scan(&id)
	if err == nil {
		return id, nil
	}
	// se non esiste, genero l'id e aggiungo l'utente alla tabella
	userId, err := createId()
	if err != nil {
		return "", err
	}
	_, err = db.c.Exec("INSERT INTO users (id, username, photo) VALUES (?, ?, '')", userId, name)
	if err != nil {
		return "", err
	}
	return userId, nil
}

// cerca nella tabella tutte gli username che contengono name e restituisce possibili corrispondenze
func (db *appdbimpl) GetUsersByName(name string) ([]User, error) {
	users := make([]User, 0)
	searchName := "%" + name + "%"                                                                            // nome contenuto nella stringa
	userResults, err := db.c.Query(`SELECT id, username, photo from users WHERE username LIKE ?`, searchName) // si cercano users con uno username che contiene name
	if err != nil {
		return nil, err
	}
	defer userResults.Close() // chiudo la ricerca per liberare la connessione
	// itero sul risultato per costruire la struct User e restituire la lista di possibili match
	for userResults.Next() {
		var user User
		err := userResults.Scan(&user.Id, &user.Username, &user.Photo)
		if err != nil {
			return nil, err
		}
		// aggiungo l'utente al risultato finale
		users = append(users, user)
	}
	if err = userResults.Err(); err != nil {
		return nil, err
	}
	if len(users) == 0 {
		return users, errors.New(RowNotFound)
	}
	return users, nil
}

// inserire il nome dell'utente, il controllo sull'unicità viene fatto direttamente sulla tabella tramite UNIQUE
// il controllo sulla validità del nome (lunghezza, pattern ecc...) viene fatto nell'handler
func (db *appdbimpl) SetUserName(id string, name string) error {
	// modifico un utente esistente
	_, err := db.c.Exec("UPDATE users SET username = ? WHERE id = ?", name, id)
	return err
}

// inserire il path della foto
func (db *appdbimpl) SetUserPhoto(id string, photo string) error {
	_, err := db.c.Exec("UPDATE users SET photo = ? WHERE id = ?", photo, id)
	return err
}
