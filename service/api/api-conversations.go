package api

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/julienschmidt/httprouter"
	"github.com/oksa-sy/WASAText/service/api/reqcontext"
)

// ritorno tutte le conversazioni dell'utente
func (rt *_router) getMyConversations(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	conversations, err := rt.db.GetAllConversations(ctx.UserId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	conversationReuslts := make([]Conversation, 0)
	for _, conversation := range conversations {
		c := Conversation{
			Id:             conversation.Id,
			Name:           conversation.Name,
			Photo:          conversation.Photo,
			TimeStamp:      conversation.TimeStamp,
			Preview:        conversation.Preview,
			PreviewIsPhoto: conversation.PreviewIsPhoto,
			IsGroup:        conversation.IsGroup,
		}
		conversationReuslts = append(conversationReuslts, c)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(conversationReuslts)
	if err != nil {
		rt.baseLogger.Errorf("Error encoding response: %v", err)
	}
}

// crea una nuova converazione
func (rt *_router) createConversation(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	newConversation := StartConversation{}
	err := json.NewDecoder(r.Body).Decode(&newConversation)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	// se l'utente cerca di iniziare una chat con se stesso
	if ctx.UserId == newConversation.ParticipantId {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	conversation, err := rt.db.CreateNewConversation(newConversation.ParticipantId, ctx.UserId)
	if err != nil {
		if err.Error() == RowNotFound {
			errorJson(w, http.StatusBadRequest, ErrCodeBadRequest, "no result with the specified criteria")
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	resultConversation := Conversation{
		Id:    conversation.Id,
		Name:  conversation.Name,
		Photo: conversation.Photo,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(resultConversation)
	if err != nil {
		rt.baseLogger.Errorf("Error encoding response: %v", err)
	}
}

// invia um messaggio in una conversazione
func (rt *_router) sendMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	// limite grandezza 10MB
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	// leggo content e se è in risposta a un messaggio
	content := r.FormValue("content")
	replyTo := r.FormValue("replyTo")
	// leggo il file
	file, fileHeader, err := r.FormFile("photo")
	var photoPath string
	// se è presente un file, controllo che esso si tratta effettivamente di un immagine
	if err == nil {
		defer file.Close()
		buff := make([]byte, 512)
		_, err = file.Read(buff)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		filetype := http.DetectContentType(buff)
		if filetype != "image/jpeg" && filetype != "image/png" {
			errorJson(w, http.StatusBadRequest, ErrCodeBadRequest, "invalid format: only jpeg and png allowed")
			return
		}
		// riporto il cursore a inizio file
		_, err = file.Seek(0, 0)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		// genero un nome univoco per il file
		bytes := make([]byte, 8)
		if _, err := rand.Read(bytes); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		fileId := hex.EncodeToString(bytes)
		extension := filepath.Ext(fileHeader.Filename)
		filename := fileId + extension
		filePath := filepath.Join("./images", filename)
		// salvo
		dst, _ := os.Create(filePath)
		defer dst.Close()
		io.Copy(dst, file)
		photoPath = "/images/" + filename
	}
	// controllo se manca anche il testo, in quel caso mando errore
	if content == "" && photoPath == "" {
		errorJson(w, http.StatusBadRequest, ErrCodeBadRequest, "message cannot be empty")
		return
	}
	newMessage := CreateMessage{}
	newMessage.Content = content
	newMessage.ReplyTo = replyTo
	newMessage.Photo = photoPath
	conversationId := ps.ByName("conversationId") // leggo l'id della conversazione dal path
	// controllo che il messaggio non sia vuoto o più lungo di 1000 caratteri
	if newMessage.Content != "" && len(newMessage.Content) < 1 || len(newMessage.Content) > 1000 {
		errorJson(w, http.StatusBadRequest, ErrCodeBadRequest, "message content cannot be empty or greater then 1000 characters")
		return
	}
	message, err := rt.db.CreateMessage(conversationId, ctx.UserId, newMessage.Content, newMessage.ReplyTo, newMessage.Photo)
	if err != nil {
		if err.Error() == RowNotFound {
			errorJson(w, http.StatusBadRequest, ErrCodeBadRequest, "no result with the specified criteria")
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	resultMessage := Message{
		Id:        message.Id,
		Content:   message.Content,
		Photo:     message.Photo,
		Username:  message.Username,
		TimeStamp: message.TimeStamp,
		Status:    Status(message.Status),
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(resultMessage)
	if err != nil {
		rt.baseLogger.Errorf("Error encoding response: %v", err)
	}
}

// ottiene la lista dei messaggi in una conversazione
func (rt *_router) getConversation(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	conversationId := ps.ByName("conversationId") // leggo l'id della conversazione dal path
	messages, err := rt.db.GetConversation(conversationId, ctx.UserId)
	if err != nil {
		if err.Error() == RowNotFound {
			errorJson(w, http.StatusBadRequest, ErrCodeBadRequest, "no result with the specified criteria")
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	messageReuslts := make([]Message, 0)
	for _, message := range messages {
		reactions := make([]Reaction, 0)
		for _, r := range message.Reactions {
			reactions = append(reactions, Reaction{
				Emoji:    r.Emoji,
				Username: r.Username,
			})
		}
		m := Message{
			Id:              message.Id,
			Content:         message.Content,
			Photo:           message.Photo,
			Username:        message.Username,
			TimeStamp:       message.TimeStamp,
			Status:          Status(message.Status),
			Reaction:        reactions,
			Forwarded:       message.Forwarded,
			ReplyToId:       message.ReplyToId,
			ReplyToContent:  message.ReplyToContent,
			ReplyToUsername: message.ReplyToUsername,
			ReplyToPhoto:    message.ReplyToPhoto,
		}
		messageReuslts = append(messageReuslts, m)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(messageReuslts)
	if err != nil {
		rt.baseLogger.Errorf("Error encoding response: %v", err)
	}
}
