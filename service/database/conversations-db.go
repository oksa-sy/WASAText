package database

import "github.com/oksa-sy/WASAText/service/globaltime"

func (db *appdbimpl) GetAllConversations(userId string) ([]Conversation, error) {
	conversationResult := make([]Conversation, 0)
	// seleziono le informazioni necessarie: COALESCE prende il primo valore non nullo in modo da prelevare la foto/nome del gruppo
	// e in caso sia NULL dell'utente partecipante alla conversazione privata
	// faccio join per unire le varie tabelle in base alla conversazione
	// con una sottoquery recupero l'id dell'ultimo messaggio inviato nella conversazione
	// raggruppo per id della conversazione per eleminare duplicati
	// ritorno il risultato ordinato in ordine cronologico inverso
	conversations, err := db.c.Query(`SELECT c.id, COALESCE(c.name, u.username), COALESCE(c.photo, u.photo), m.timestamp, m.content
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
		err := conversations.Scan(&conversation.Id, &conversation.Name, &conversation.Photo, &conversation.TimeStamp, &conversation.Preview)
		if err != nil {
			return nil, err
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
	// controllo che esiste un utente con quell'id
	var username string
	var photo string
	err := db.c.QueryRow(`SELECT username, photo FROM users WHERE id = ?`, participantId).Scan(&username, &photo)
	if err != nil {
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
func (db *appdbimpl) CreateMessage(conversationId string, userId string, content string, replyTo string) (Message, error) {
	message := Message{}
	// controllo che esiste una conversazione con quell'id in cui partecipa l'utente
	var cId string
	var uId string
	err := db.c.QueryRow(`SELECT conversation_id, user_id FROM participants WHERE conversation_id = ? and user_id = ?`, conversationId, userId).Scan(&cId, &uId)
	if err != nil {
		return message, err
	}
	// creo il messaggio
	messageId, err := createId()
	if err != nil {
		return message, err
	}
	currentTime := globaltime.Now()
	if replyTo == "" {
		_, err = db.c.Exec("INSERT INTO messages (id, conversation_id, sender_id, content, timestamp, status) VALUES (?, ?, ?, ?, ?, 'recieved')", messageId, conversationId, userId, content, currentTime)
		if err != nil {
			return message, err
		}
	} else {
		// controllo che esiste un messaggio con id specificato nella conversazione specificata
		var mId string
		err := db.c.QueryRow(`SELECT id FROM messages WHERE conversation_id = ? and id = ?`, conversationId, replyTo).Scan(&mId)
		if err != nil {
			return message, err
		}
		// se esiste lo memorizzo, altrimenti ritorno l'errore
		_, err = db.c.Exec("INSERT INTO messages (id, conversation_id, sender_id, content, timestamp, reply_to, status) VALUES (?, ?, ?, ?, ?, ?, 'recieved')", messageId, conversationId, userId, content, currentTime, replyTo)
		if err != nil {
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
		Username:  senderName,
		TimeStamp: currentTime,
		Status:    "recieved",
	}
	return message, nil
}

func (db *appdbimpl) GetConversation(conversationId string, userId string) ([]Message, error) {
	messageReuslts := make([]Message, 0)
	// controllo che l'utente faccia parte della conversazione
	var user string
	err := db.c.QueryRow(`SELECT user_id FROM participants WHERE user_id = ? and conversation_id = ?`, userId, conversationId).Scan(&user)
	if err != nil {
		return nil, err
	}
	// segno i messaggi nella conversazione inviati dall'altro utente come letti
	_, err = db.c.Exec("UPDATE messages SET status = 'read' WHERE conversation_id = ? and sender_id != ? and status = 'recieved'", conversationId, userId)
	if err != nil {
		return nil, err
	}
	// lista dei messaggi in ordine cronologico inverso
	messages, err := db.c.Query(`SELECT m.id, u.username, m.content, m.timestamp FROM messages m, users u WHERE m.conversation_id = ? and u.id = m.sender_id ORDER BY timestamp desc`, conversationId)
	if err != nil {
		return nil, err
	}
	defer messages.Close() // chiudo la ricerca per liberare la connessione
	// costruisco la struct risultate per ogni riga risultante dalla select
	for messages.Next() {
		var message Message
		err := messages.Scan(&message.Id, &message.Username, &message.Content, &message.TimeStamp)
		if err != nil {
			return nil, err
		}
		// aggiungo il messaggio al risultato finale
		messageReuslts = append(messageReuslts, message)
	}
	if err = messages.Err(); err != nil {
		return nil, err
	}
	return messageReuslts, nil
}
