package database

import (
	"database/sql"
	"errors"

	"github.com/oksa-sy/WASAText/service/globaltime"
)

// inoltrare il messaggio in una conversazione: prendo i dati nel messaggio originale e li aggiungo in una nuova riga con la conversazione specificata
func (db *appdbimpl) ForwardMessage(userId string, participantId string, messageId string) (Message, error) {
	message := Message{}
	// controllo che il messaggio esiste e che l'untente faccia parte di quella conversazione, prendo il contenuto
	var content string
	conv, err := db.CreateNewConversation(participantId, userId)
	convId := conv.Id
	if err != nil {
		return message, err
	}
	// recupero il contenuto del messaggio solo se l'utente fa parte di quella conversazione
	err = db.c.QueryRow(`SELECT m.content FROM messages m, participants p WHERE m.id = ? AND m.conversation_id = p.conversation_id AND p.user_id = ?`, messageId, userId).Scan(&content)
	if err != nil {
		return message, err
	}
	// creo il nuovo messaggio nella conversaione specificata con gli stessi dati dell'originale
	forwardedMessageId, err := createId()
	if err != nil {
		return message, err
	}
	currentTime := globaltime.Now()
	_, err = db.c.Exec("INSERT INTO messages (id, conversation_id, sender_id, content, timestamp, status) VALUES (?, ?, ?, ?, ?, 'recieved')", forwardedMessageId, convId, userId, content, currentTime)
	if err != nil {
		return message, err
	}
	// username utente che inoltra il messaggio
	var username string
	err = db.c.QueryRow(`SELECT username FROM users WHERE id = ?`, userId).Scan(&username)
	if err != nil {
		return message, err
	}
	message = Message{
		Id:        forwardedMessageId,
		Content:   content,
		Username:  username,
		TimeStamp: currentTime,
		Status:    "recieved",
	}
	return message, nil
}

// eliminare un messaggio, controllo che l'utente ne sia il proprietario
func (db *appdbimpl) DeleteMessage(userId string, messageId string) error {
	// controllo che l'utente abbia invito il messaggio
	var exists int
	err := db.c.QueryRow(`SELECT 1 FROM messages WHERE id = ? and sender_id = ?`, messageId, userId).Scan(&exists)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New(RowNotFound)
		}
		return err
	}
	// elimino il messaggio
	_, err = db.c.Exec("DELETE FROM messages WHERE id = ?", messageId)
	if err != nil {
		return err
	}
	return nil
}

// aggiugere una reazione a un messaggio
func (db *appdbimpl) CreateReaction(reaction string, messageId string, userId string) (Reaction, error) {
	r := Reaction{}
	// controllo che il messaggio esiste e fa parte di una conversazione in cui l'utente partecipa e non è colui che ha inviato il messaggio, ottengo il suo username
	var username string
	err := db.c.QueryRow(`SELECT u.username FROM messages m, participants p, users u WHERE p.user_id = ? and m.conversation_id = p.conversation_id and m.id = ? and m.sender_id != ? and u.id = ?`, userId, messageId, userId, userId).Scan(&username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return r, errors.New(RowNotFound)
		}
		return r, err
	}
	// controllo il numero di reazioni del messaggio, se più di 100 errore
	var numReactions int
	err = db.c.QueryRow(`SELECT COUNT(*) FROM reactions WHERE message_id = ?`, messageId).Scan(&numReactions)
	if err != nil {
		return r, err
	}
	if numReactions >= 100 {
		return r, errors.New("limit exceeded")
	}
	// se esiste inserisco la reazione
	_, err = db.c.Exec("INSERT INTO reactions (message_id, user_id, reaction) VALUES (?, ?, ?)", messageId, userId, reaction)
	if err != nil {
		return r, err
	}
	r = Reaction{
		Emoji:    reaction,
		Username: username,
	}
	return r, nil
}

// eliminare una reazione, controllo che l'utente ne sia il proprietario
func (db *appdbimpl) DeleteReaction(reaction string, userId string, messageId string) error {
	// controllo che l'utente abbia inviato la reazione
	var exists int
	err := db.c.QueryRow(`SELECT 1 FROM reactions WHERE message_id = ? and user_id = ? and reaction = ?`, messageId, userId, reaction).Scan(&exists)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New(RowNotFound)
		}
		return err
	}
	// elimino la reazione
	_, err = db.c.Exec("DELETE FROM reactions WHERE message_id = ? and reaction = ? and user_id = ?", messageId, reaction, userId)
	if err != nil {
		return err
	}
	return nil
}
