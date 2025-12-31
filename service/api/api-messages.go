package api

import (
	"encoding/json"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/oksa-sy/WASAText/service/api/reqcontext"
)

// inoltrare un messaggio in una conversazione: id messaggio nek path, id utente nel requestbody
func (rt *_router) forwardMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	var userId struct {
		UserId string `json:"id"`
	}
	// controllo se il json è corrotto e metto il contenuto in userId
	err := json.NewDecoder(r.Body).Decode(&userId)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	messageId := ps.ByName("messageId") // leggo l'id del messaggio dal path
	message, err := rt.db.ForwardMessage(ctx.UserId, userId.UserId, messageId)
	if err != nil {
		// controllo se l'errore è "not found", in quel caso ritorno un 404
		if err.Error() == RowNotFound {
			errorJson(w, http.StatusNotFound, ErrCodeNotFound, "no result with the specified criteria")
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	resultMessage := Message{
		Id:        message.Id,
		Content:   message.Content,
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

// eliminare il messaggio, id nel path
func (rt *_router) deleteMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	messageId := ps.ByName("messageId") // leggo l'id del messaggio dal path
	err := rt.db.DeleteMessage(ctx.UserId, messageId)
	if err != nil {
		if err.Error() == RowNotFound {
			errorJson(w, http.StatusNotFound, ErrCodeNotFound, "no result with the specified criteria")
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNoContent)
}

// aggiungere una reazione a un messaggio
func (rt *_router) commentMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	emoji := CreateReaction{}
	// controllo se il json è corrotto e metto il contenuto in emoji
	err := json.NewDecoder(r.Body).Decode(&emoji)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	messageId := ps.ByName("messageId") // leggo l'id del messaggio dal path
	reaction, err := rt.db.CreateReaction(emoji.Emoji, messageId, ctx.UserId)
	if err != nil {
		if err.Error() == "limit exceeded" {
			errorJson(w, http.StatusBadRequest, ErrCodeBadRequest, "limit of reactions exeeded")
			return
		}
		// controllo se l'errore è "not found", in caso ritorno 400
		if err.Error() == RowNotFound {
			errorJson(w, http.StatusBadRequest, ErrCodeBadRequest, "no result with the specified criteria")
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	resultReaction := Reaction{
		Emoji:    reaction.Emoji,
		Username: reaction.Username,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(resultReaction)
	if err != nil {
		rt.baseLogger.Errorf("Error encoding response: %v", err)
	}
}

func (rt *_router) uncommentMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	emoji := r.URL.Query().Get("emoji")
	if emoji == "" {
		errorJson(w, http.StatusNotFound, ErrCodeNotFound, "missing emoji")
		return
	}
	messageId := ps.ByName("messageId") // leggo l'id del messaggio dal path
	err := rt.db.DeleteReaction(emoji, ctx.UserId, messageId)
	if err != nil {
		if err.Error() == RowNotFound {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNoContent)
}
