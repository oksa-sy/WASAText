package api

import (
	"encoding/json"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (rt *_router) doLogin(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// decodifica dello json in ingresso
	var body struct {
		Username string `json:"name"`
	}
	// controllo se il json è corrotto e metto il contenuto in requestBody
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	// controllo su lunghezza e pattern per verificare le regole definite nell'api
	if !validateName(body.Username) {
		errorJson(w, http.StatusBadRequest, ErrCodeBadRequest, "invalid: username must be between 3 and 16 characters or it contains invalid characters (alphanumeric only)")
		return
	}
	// chiamo la funzione in database per ottenere l'id utente
	userId, err := rt.db.CreateUser(body.Username)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	// struct per la risposta
	var response struct {
		Identifier string `json:"identifier"`
	}
	response.Identifier = userId
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		rt.baseLogger.Errorf("Error encoding response: %v", err)
	}
}
