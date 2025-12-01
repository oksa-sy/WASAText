package api

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/julienschmidt/httprouter"
	"github.com/oksa-sy/WASAText/service/api/reqcontext"
)

func (rt *_router) createGroup(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	group := Group{}
	// controllo se il json è corrotto e metto il contenuto in group
	err := json.NewDecoder(r.Body).Decode(&group)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	// controllo sulla validità dell'input
	if len(group.Members) < 1 || len(group.Members) > 100 {
		errorJson(w, http.StatusBadRequest, ErrCodeBadRequest, "invalid number of members")
		return
	}
	if !validateGroupName(group.Name) {
		errorJson(w, http.StatusBadRequest, ErrCodeBadRequest, "invalid: group name must be between 3 and 16 characters or it contains invalid characters")
		return
	}
	resultGroup, err := rt.db.CreateNewGroup(ctx.UserId, group.Name, group.Members)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	group = Group{
		Id:      resultGroup.Id,
		Name:    resultGroup.Name,
		Members: resultGroup.Members,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(group)
	if err != nil {
		rt.baseLogger.Errorf("Error encoding response: %v", err)
	}
}

func (rt *_router) setGroupName(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	// leggo da path l'id del gruppo
	groupId := ps.ByName("groupId")
	var body struct {
		Name string `json:"name"`
	}
	// controllo se il json è corrotto e metto il contenuto in body
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if !validateGroupName(body.Name) {
		errorJson(w, http.StatusBadRequest, ErrCodeBadRequest, "invalid: group name must be between 3 and 16 characters or it contains invalid characters")
		return
	}
	name, err := rt.db.UpdateGroupName(body.Name, groupId, ctx.UserId)
	if err != nil {
		// controllo se l'errore è "group not found" o "user not in group", in quel caso ritorno un 404
		if err.Error() == RowNotFound {
			errorJson(w, http.StatusBadRequest, ErrCodeBadRequest, "no result with the specified criteria")
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	var nameResult struct {
		Name string `json:"name"`
	}
	nameResult.Name = name
	// se è tutto ok invio 200 ok con nome scelto
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(nameResult)
	if err != nil {
		rt.baseLogger.Errorf("Error encoding response: %v", err)
	}
}

// aggiungere un utente a un gruppo di cui si fa parte
func (rt *_router) addToGroup(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	var id struct {
		Id string `json:"userId"`
	}
	// controllo se il json è corrotto e metto il contenuto in id
	err := json.NewDecoder(r.Body).Decode(&id)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	// un utente non puo aggiungere se stesso
	if ctx.UserId == id.Id {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	groupId := ps.ByName("groupId") // leggo l'id del gruppo dal path
	idResult, err := rt.db.AddUserToGroup(id.Id, ctx.UserId, groupId)
	if err != nil {
		if err.Error() == "limit exceeded" {
			errorJson(w, http.StatusBadRequest, ErrCodeBadRequest, "limit of members exceeded")
			return
		}
		if err.Error() == RowNotFound {
			errorJson(w, http.StatusBadRequest, ErrCodeBadRequest, "no result with the specified criteria")
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(idResult)
	if err != nil {
		rt.baseLogger.Errorf("Error encoding response: %v", err)
	}
}

func (rt *_router) leaveGroup(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	groupId := ps.ByName("groupId") // leggo l'id del gruppo dal path
	err := rt.db.LeaveGroup(ctx.UserId, groupId)
	if err != nil {
		// controllo se l'errore è RowNotFound, in quel caso ritorno un 404
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

// recupero l'immagine dall'input utente, la salvo in images e restituisco il path
func (rt *_router) setGroupPhoto(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	// controllo che il file non superi 5MB
	err := r.ParseMultipartForm(5 << 20)
	if err != nil {
		errorJson(w, http.StatusBadRequest, ErrCodeBadRequest, "file is too large")
		return
	}
	// lettura del file dalla richiesta
	file, fileHeader, err := r.FormFile("photoFile")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer file.Close()
	// lettura dei primi 512 byte: contengono informazioni per verificare che si tratta effettivamente di un' immagine
	buff := make([]byte, 512)
	_, err = file.Read(buff)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	filetype := http.DetectContentType(buff)
	if filetype != "image/jpeg" && filetype != "image/png" {
		errorJson(w, http.StatusBadRequest, ErrCodeBadRequest, "invalid format, only jpeg and png allowed")
		return
	}
	// riporto il cursore a inizio file
	_, err = file.Seek(0, 0)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	groupId := ps.ByName("groupId") // leggo l'id del gruppo dal path
	// creazione file, memorizzato tramite id del gruppo per evitare duplicati
	extension := filepath.Ext(fileHeader.Filename)
	if extension == "" {
		extension = ".png"
	}
	filename := groupId + extension
	filePath := filepath.Join("./images", filename) // salva in cartella images
	// creo file
	dst, err := os.Create(filePath)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer dst.Close()
	// copio i dati dell'immagine nel file appena creato
	_, err = io.Copy(dst, file)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	// URL per il data base
	photoURL := "/images/" + filename
	err = rt.db.SetNewGroupPhoto(groupId, photoURL, ctx.UserId)
	if err != nil {
		if err.Error() == RowNotFound {
			errorJson(w, http.StatusNotFound, ErrCodeNotFound, "no result with the specified criteria")
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	var photo struct {
		Photo string `json:"photo"`
	}
	photo.Photo = photoURL
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(photo)
	if err != nil {
		rt.baseLogger.Errorf("Error encoding response: %v", err)
	}
}
