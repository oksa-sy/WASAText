package database

import (
	"database/sql"
	"errors"
)

// creazione di un nuovo gruppo, unificato conversazione, uso la flag 1 per indicare che si tratta di un gruppo
func (db *appdbimpl) CreateNewGroup(user string, name string, members []string) (Group, error) {
	group := Group{}
	// creo il gruppo e lo inserisco nel database
	conversationId, err := createId()
	if err != nil {
		return group, err
	}
	// inizializzo una conversazione segnando che si tratta di un gruppo
	_, err = db.c.Exec("INSERT INTO conversations (id, is_group, name) VALUES (?, 1, ?)", conversationId, name)
	if err != nil {
		return group, err
	}
	// segno la converazione per ogni membro del gruppo
	for _, userId := range members {
		_, err = db.c.Exec("INSERT INTO participants (conversation_id, user_id) VALUES (?, ?)", conversationId, userId)
		if err != nil {
			return group, err
		}
	}
	// inserisco l'utente stesso che crea il gruppo
	_, err = db.c.Exec("INSERT INTO participants (conversation_id, user_id) VALUES (?, ?)", conversationId, user)
	if err != nil {
		return group, err
	}
	// aggiungo l'utente stesso alla lista dei membri
	members = append(members, user)
	group = Group{
		Id:      conversationId,
		Name:    name,
		Members: members,
	}
	return group, nil
}

// aggiorno nome del gruppo
func (db *appdbimpl) UpdateGroupName(name string, groupId string, userId string) (string, error) {
	// controllo se l'utente fa parte del gruppo
	var id string
	err := db.c.QueryRow(`SELECT user_id FROM participants WHERE conversation_id = ? and user_id = ?`, groupId, userId).Scan(&id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", errors.New(RowNotFound)
		}
		return "", err
	}
	r, err := db.c.Exec("UPDATE conversations SET name = ? WHERE id = ? and is_group = 1", name, groupId)
	if err != nil {
		return "", err
	}
	// controllo per vedere se l'id è esistene, se non modifica nessuna riga allora non esite
	affected, err := r.RowsAffected()
	if err != nil {
		if affected == 0 {
			return "", errors.New(RowNotFound)
		}
		return "", err
	}
	return name, err
}

// inserire il path della foto
func (db *appdbimpl) SetNewGroupPhoto(groupId string, photo string, userId string) error {
	// controllo che l'utente è nel gruppo e che il gruppo esista
	var exists int
	err := db.c.QueryRow(`SELECT 1 FROM participants WHERE conversation_id = ? and user_id = ?`, groupId, userId).Scan(&exists)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New(RowNotFound)
		}
		return err
	}
	_, err = db.c.Exec("UPDATE conversations SET photo = ? WHERE id = ?", photo, groupId)
	return err
}

// aggiungere un utente al gruppo
// bisogna essere nel gruppo
func (db *appdbimpl) AddUserToGroup(memberId string, userId string, groupId string) (string, error) {
	// controllo se la conversazione esiste, se l'utente partecipa ad essa e se è un gruppo
	var isGroup bool
	err := db.c.QueryRow(`SELECT is_group FROM conversations WHERE id = ?`, groupId).Scan(&isGroup)
	if err != nil || !isGroup {
		if errors.Is(err, sql.ErrNoRows) {
			return "", errors.New(RowNotFound)
		}
		return "", err
	}
	var exists int
	err = db.c.QueryRow(`SELECT 1 from participants WHERE user_id = ? and conversation_id = ?`, userId, groupId).Scan(&exists)
	if err != nil || !isGroup {
		if errors.Is(err, sql.ErrNoRows) {
			return "", errors.New(RowNotFound)
		}
		return "", err
	}
	// controllo se l'utente fa parte del gruppo
	err = db.c.QueryRow("SELECT 1 FROM users WHERE id = ?", memberId).Scan(&exists)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", errors.New(RowNotFound)
		}
		return "", err
	}
	var numMembers int
	// controllo che aggiungendo il membro non supero il limite sui partecipanti
	err = db.c.QueryRow("SELECT COUNT (*) FROM participants WHERE conversation_id = ?", groupId).Scan(&numMembers)
	if err != nil {
		return "", err
	}
	if numMembers >= 100 {
		return "", errors.New("limit exceeded")
	}
	// se presente, aggiungo il membro al gruppo
	_, err = db.c.Exec("INSERT INTO participants (conversation_id, user_id) VALUES (?, ?)", groupId, memberId)
	if err != nil {
		return "", err
	}
	return memberId, err
}

// abbandonare un gruppo, solo se si è membro
func (db *appdbimpl) LeaveGroup(userId string, groupId string) error {
	// controllo se la conversazione esiste, se l'utente partecipa ad essa e se è un gruppo
	var isGroup bool
	err := db.c.QueryRow(`SELECT is_group FROM conversations WHERE id = ?`, groupId).Scan(&isGroup)
	if err != nil || !isGroup {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New(RowNotFound)
		}
		return err
	}
	var exists int
	err = db.c.QueryRow(`SELECT 1 from participants WHERE user_id = ? and conversation_id = ?`, userId, groupId).Scan(&exists)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New(RowNotFound)
		}
		return err
	}
	// se tutto ok elimino l'utente da participants
	_, err = db.c.Exec("DELETE FROM participants WHERE user_id = ? and conversation_id = ?", userId, groupId)
	if err != nil {
		return err
	}
	// controllo quanti utenti rimangono nel gruppo
	var numOfUsers int
	err = db.c.QueryRow(`SELECT COUNT(*) FROM participants WHERE conversation_id = ?`, groupId).Scan(&numOfUsers)
	if err != nil {
		return err
	}
	// se non ci sono più membri del gruppo lo elimino
	if numOfUsers == 0 {
		_, err = db.c.Exec("DELETE FROM conversations WHERE id = ?", groupId)
		if err != nil {
			return err
		}
	}
	return nil
}
