package database

import (
	"database/sql"
	"errors"

	"github.com/oksa-sy/WASAText/service/globaltime"
)

func (db *appdbimpl) GetAllConversations(userId string) ([]Conversation, error) {
	conversationResult := make([]Conversation, 0)
	// seleziono le informazioni necessarie: COALESCE prende il primo valore non nullo in modo da prelevare la foto/nome del gruppo
	// e in caso sia NULL dell'utente partecipante alla conversazione privata (solo se non è un gruppo)
	// faccio join per unire le varie tabelle in base alla conversazione
	// con una sottoquery recupero l'id dell'ultimo messaggio inviato nella conversazione
	// raggruppo per id della conversazione per eleminare duplicati
	// ritorno il risultato ordinato in ordine cronologico inverso
	conversations, err := db.c.Query(`SELECT c.id, COALESCE(c.name, u.username), CASE WHEN c.is_group = 1 THEN c.photo ELSE u.photo END, m.timestamp, 
											m.content, m.photo, c.is_group
										FROM conversations c
										JOIN participants p1 ON c.id = p1.conversation_id
										LEFT JOIN participants p2 ON c.id = p2.conversation_id AND p2.user_id != p1.user_id
										LEFT JOIN users u ON p2.user_id = u.id
										LEFT JOIN messages m ON m.id = (
											SELECT id FROM messages 
											WHERE conversation_id = c.id 
											ORDER BY timestamp DESC LIMIT 1)
										WHERE p1.user_id = ?
										GROUP BY c.id 
										ORDER BY m.timestamp DESC`, userId)
	if err != nil {
		return nil, err
	}
	defer conversations.Close() // chiudo la ricerca per liberare la connessione
	// costruisco la struct risultate per ogni riga risultante dalla select
	for conversations.Next() {
		var conversation Conversation
		// per considerare il caso in cui timestamp e content sono NULL
		var groupPhoto sql.NullString
		var dbTime sql.NullTime
		var dbContent sql.NullString
		var dbPhoto sql.NullString
		err := conversations.Scan(&conversation.Id, &conversation.Name, &groupPhoto, &dbTime, &dbContent, &dbPhoto, &conversation.IsGroup)
		if err != nil {
			return nil, err
		}
		// se i valori sono non null, li metto nella struct
		if groupPhoto.Valid {
			conversation.Photo = groupPhoto.String
		} else {
			conversation.Photo = ""
		}
		if dbTime.Valid {
			conversation.TimeStamp = dbTime.Time
		}
		if dbContent.String != "" && dbContent.Valid {
			conversation.Preview = dbContent.String
		} else {
			conversation.Preview = ""
		}
		if dbPhoto.String != "" && dbPhoto.Valid {
			conversation.PreviewIsPhoto = true
		} else {
			conversation.PreviewIsPhoto = false
		}
		conversationResult = append(conversationResult, conversation)
	}
	if err = conversations.Err(); err != nil {
		return nil, err
	}
	return conversationResult, nil
}

// creo una conversazione 1a1 con l'utente desiderato
func (db *appdbimpl) CreateNewConversation(participantId string, userId string) (Conversation, error) {
	conversation := Conversation{}
	// controllo se esiste gia una chat con quell'utente
	var convId string
	var convName string
	var convPhoto sql.NullString
	// controllo se la conversazione è un gruppo di cui l'utente fa parte
	err := db.c.QueryRow(`SELECT c.id, c.name, c.photo FROM participants p, conversations c WHERE p.user_id = ? AND p.conversation_id = ? AND c.id = ? AND c.is_group = 1`, userId, participantId, participantId).Scan(&convId, &convName, &convPhoto)
	// se esiste ritorno la conversazione
	if err == nil {
		finalPhoto := ""
		if convPhoto.Valid {
			finalPhoto = convPhoto.String
		}
		conversation = Conversation{
			Id:    convId,
			Name:  convName,
			Photo: finalPhoto,
		}
		return conversation, nil
	}
	// se si verifica un errore diverso da errnorows lo ritorno, senno creo la nuova chat
	if !errors.Is(err, sql.ErrNoRows) {
		return conversation, err
	}
	// controllo se è una chat 1-1
	err = db.c.QueryRow(`SELECT c.id FROM participants p1, participants p2, conversations c WHERE p1.user_id = ? AND p2.user_id = ? AND p1.conversation_id = p2.conversation_id AND c.id = p1.conversation_id AND c.is_group = 0`, participantId, userId).Scan(&convId)
	// se esiste recupero i dati del partecipante e ritorno la conversazione
	if err == nil {
		var username string
		var photo string
		err = db.c.QueryRow(`SELECT username, photo FROM users WHERE id = ?`, participantId).Scan(&username, &photo)
		if err != nil {
			return conversation, err
		}
		conversation = Conversation{
			Id:    convId,
			Name:  username,
			Photo: photo,
		}
		return conversation, nil
	}
	// se si verifica un errore diverso da errnorows lo ritorno, senno creo la nuova chat
	if !errors.Is(err, sql.ErrNoRows) {
		return conversation, err
	}
	// controllo che esiste un utente con quell'id
	var username string
	var photo string
	err = db.c.QueryRow(`SELECT username, photo FROM users WHERE id = ?`, participantId).Scan(&username, &photo)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return conversation, errors.New(RowNotFound)
		}
		return conversation, err
	}
	// se esiste genero l'id della conversazione
	conversationId, err := createId()
	if err != nil {
		return conversation, err
	}
	// creo la conversazione
	_, err = db.c.Exec("INSERT INTO conversations (id, is_group) VALUES (?, ?)", conversationId, 0)
	if err != nil {
		return conversation, err
	}
	conversation = Conversation{
		Id:    conversationId,
		Name:  username,
		Photo: photo,
	}
	// memorizzo la conversazione e i partecipanti
	// memorizzo il destinatario
	_, err = db.c.Exec("INSERT INTO participants (conversation_id, user_id) VALUES (?, ?)", conversationId, participantId)
	if err != nil {
		return conversation, err
	}
	// memorizzo il mittente
	_, err = db.c.Exec("INSERT INTO participants (conversation_id, user_id) VALUES (?, ?)", conversationId, userId)
	if err != nil {
		return conversation, err
	}
	return conversation, nil
}

// creo il messaggio e lo memorizzo nel database
func (db *appdbimpl) CreateMessage(conversationId string, userId string, content string, replyTo string, photo string) (Message, error) {
	message := Message{}
	// controllo che esiste una conversazione con quell'id in cui partecipa l'utente
	var cId string
	var uId string
	err := db.c.QueryRow(`SELECT conversation_id, user_id FROM participants WHERE conversation_id = ? and user_id = ?`, conversationId, userId).Scan(&cId, &uId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return message, errors.New(RowNotFound)
		}
		return message, err
	}
	// creo il messaggio
	messageId, err := createId()
	if err != nil {
		return message, err
	}
	currentTime := globaltime.Now()
	if replyTo == "" {
		_, err = db.c.Exec("INSERT INTO messages (id, conversation_id, sender_id, content, timestamp, status, photo) VALUES (?, ?, ?, ?, ?, 'recieved', ?)", messageId, conversationId, userId, content, currentTime, photo)
		if err != nil {
			return message, err
		}
	} else {
		// controllo che esiste un messaggio con id specificato nella conversazione specificata
		var mId string
		err := db.c.QueryRow(`SELECT id FROM messages WHERE conversation_id = ? and id = ?`, conversationId, replyTo).Scan(&mId)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return message, errors.New(RowNotFound)
			}
			return message, err
		}
		// se esiste lo memorizzo, altrimenti ritorno l'errore
		_, err = db.c.Exec("INSERT INTO messages (id, conversation_id, sender_id, content, timestamp, reply_to, status, photo) VALUES (?, ?, ?, ?, ?, ?, 'recieved', ?)", messageId, conversationId, userId, content, currentTime, replyTo, photo)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return message, errors.New(RowNotFound)
			}
			return message, err
		}
	}
	// trovo lo username corrispondente all'id del mittente
	var senderName string
	err = db.c.QueryRow(`SELECT username FROM users WHERE id = ?`, userId).Scan(&senderName)
	if err != nil {
		return message, err
	}
	message = Message{
		Id:        messageId,
		Content:   content,
		Photo:     photo,
		Username:  senderName,
		TimeStamp: currentTime,
		Status:    "recieved",
	}
	return message, nil
}

func (db *appdbimpl) GetConversation(conversationId string, userId string) ([]Message, error) {
	messageResults := make([]Message, 0)
	// controllo che l'utente faccia parte della conversazione
	var user string
	err := db.c.QueryRow(`SELECT user_id FROM participants WHERE user_id = ? and conversation_id = ?`, userId, conversationId).Scan(&user)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New(RowNotFound)
		}
		return nil, err
	}
	// conto i partecipanti alla conversazione
	var participants int
	err = db.c.QueryRow(`SELECT count(user_id) FROM participants WHERE conversation_id = ?`, conversationId).Scan(&participants)
	if err != nil {
		return nil, err
	}
	// lista dei messaggi in ordine cronologico inverso
	messages, err := db.c.Query(`SELECT m.id, u.username, m.content, m.timestamp, m.status, m.reply_to, m.forwarded, m.photo FROM messages m, users u WHERE m.conversation_id = ? and u.id = m.sender_id ORDER BY timestamp desc`, conversationId)
	if err != nil {
		return nil, err
	}
	defer messages.Close() // chiudo la ricerca per liberare la connessione
	// itero sui messaggi per salvarli in un array e liberare la connessione
	for messages.Next() {
		var m Message
		var replyToId sql.NullString
		err := messages.Scan(&m.Id, &m.Username, &m.Content, &m.TimeStamp, &m.Status, &replyToId, &m.Forwarded, &m.Photo)
		if err != nil {
			messages.Close()
			return nil, err
		}
		if replyToId.Valid {
			m.ReplyToId = replyToId.String
		}
		messageResults = append(messageResults, m)
	}
	messages.Close()
	if err = messages.Err(); err != nil {
		return nil, err
	}
	// itero sui messaggi ottenuti
	for i := range messageResults {
		msg := &messageResults[i]
		// segno che il messaggio è stato letto dall'utente
		_, err = db.c.Exec("INSERT OR IGNORE INTO reads (message_id, user_id) VALUES (?, ?)", msg.Id, userId)
		if err != nil {
			return nil, err
		}
		// controllo se il messaggio riusulta letto da tutti i partecipanti della conversazione se si segno il messaggio come letto
		var nReads int
		err = db.c.QueryRow(`SELECT count(user_id) FROM reads WHERE message_id = ?`, msg.Id).Scan(&nReads)
		if err != nil {
			return nil, err
		}
		if nReads >= participants {
			_, err = db.c.Exec("UPDATE messages SET status = 'read' WHERE conversation_id = ? and sender_id != ? and status = 'recieved' AND id = ?", conversationId, userId, msg.Id)
			if err != nil {
				return nil, err
			}
		}
		// se è in risposta a un altro messaggio recupero il contenuto
		if msg.ReplyToId != "" {
			// gestione errore se il messaggio padre è stato cancellato
			var replyUsername string
			var replyContent string
			var replyPhoto string
			err = db.c.QueryRow(`SELECT m.content, u.username, m.photo FROM messages m, users u WHERE m.id = ? AND m.sender_id = u.id`, msg.ReplyToId).Scan(&replyContent, &replyUsername, &replyPhoto)
			if err == nil {
				// se tutto ok assegno il contenuto
				msg.ReplyToContent = replyContent
				msg.ReplyToUsername = replyUsername
				msg.ReplyToPhoto = replyPhoto
			} else if !errors.Is(err, sql.ErrNoRows) {
				return nil, err
			}
		}
		// recupero le reazioni su quel messaggio e il nome di chi le ha mandate
		reactionResults, err := db.c.Query(`SELECT r.reaction, u.username FROM reactions r, users u WHERE r.message_id = ? AND r.user_id = u.id`, msg.Id)
		if err != nil {
			return nil, err
		}
		msg.Reactions = make([]Reaction, 0)
		for reactionResults.Next() {
			var r Reaction
			err := reactionResults.Scan(&r.Emoji, &r.Username)
			if err != nil {
				return nil, err
			}
			msg.Reactions = append(msg.Reactions, r)
		}
		if err = reactionResults.Err(); err != nil {
			return nil, err
		}
		reactionResults.Close()
	}
	return messageResults, nil
}
