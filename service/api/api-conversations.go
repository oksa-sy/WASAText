package api

import (
	"encoding/json"
	"net/http"

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
			Id:        conversation.Id,
			Name:      conversation.Name,
			Photo:     conversation.Photo,
			TimeStamp: conversation.TimeStamp,
			Preview:   conversation.Preview,
			IsGroup:   conversation.IsGroup,
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
	newMessage := CreateMessage{}
	err := json.NewDecoder(r.Body).Decode(&newMessage)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	conversationId := ps.ByName("conversationId") // leggo l'id della conversazione dal path
	// controllo che il messaggio non sia vuoto o più lungo di 1000 caratteri
	if len(newMessage.Content) < 1 || len(newMessage.Content) > 1000 {
		errorJson(w, http.StatusBadRequest, ErrCodeBadRequest, "message content cannot be empty or greater then 1000 characters")
		return
	}
	message, err := rt.db.CreateMessage(conversationId, ctx.UserId, newMessage.Content, newMessage.ReplyTo)
	if err != nil {
		if err.Error() == RowNotFound {
			errorJson(w, http.StatusBadRequest, ErrCodeBadRequest, "no result with the specified criteria")
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	resultMessage := Message{
		Id:             message.Id,
		Content:        message.Content,
		Username:       message.Username,
		TimeStamp:      message.TimeStamp,
		Status:         Status(message.Status),
		ReplyToId:      message.ReplyToId,
		ReplyToContent: message.ReplyToContent,
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
			Id:             message.Id,
			Content:        message.Content,
			Username:       message.Username,
			TimeStamp:      message.TimeStamp,
			Status:         Status(message.Status),
			Reaction:       reactions,
			ReplyToId:      message.ReplyToId,
			ReplyToContent: message.ReplyToContent,
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
