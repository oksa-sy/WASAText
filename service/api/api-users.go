package api

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/julienschmidt/httprouter"
	"github.com/oksa-sy/WASAText/service/api/reqcontext"
)

func (rt *_router) searchUsers(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	username := r.URL.Query().Get(("username"))
	if !validateName(username) {
		errorJson(w, http.StatusBadRequest, ErrCodeBadRequest, "invalid: username must be between 3 and 16 characters or it contains invalid characters (alphanumeric only)")
		return
	}
	result, err := rt.db.GetUsersByName(username)
	if err != nil {
		if err.Error() == RowNotFound {
			errorJson(w, http.StatusBadRequest, ErrCodeBadRequest, "no result with the specified criteria")
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	// creo la struct user necessaria nel model del database per rispettare l'api
	finalResult := make([]User, 0)
	for _, user := range result {
		u := User{
			Id:       user.Id,
			Username: user.Username,
			Photo:    user.Photo,
		}
		finalResult = append(finalResult, u)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(finalResult)
	if err != nil {
		rt.baseLogger.Errorf("Error encoding response: %v", err)
	}

}

func (rt *_router) setMyUserName(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	var body struct {
		Username string `json:"name"`
	}
	// controllo se il json è corrotto e metto il contenuto in body
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if !validateName(body.Username) {
		errorJson(w, http.StatusBadRequest, ErrCodeBadRequest, "invalid: username must be between 3 and 16 characters or it contains invalid characters (alphanumeric only)")
		return
	}
	// passo l'user di input al database
	err = rt.db.SetUserName(ctx.UserId, body.Username)
	// controlo se si verificano errori
	if err != nil {
		// controllo se l'errore è causato dalla UNIQUE nel database, in quel caso l'utente sta provando a inserire un username gia esistente
		if strings.Contains(err.Error(), "constraint failed") || strings.Contains(err.Error(), "UNIQUE") {
			errorJson(w, http.StatusConflict, ErrCodeConflict, "username already taken")
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	// se è tutto ok invio 200 ok con username scelto
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(body)
	if err != nil {
		rt.baseLogger.Errorf("Error encoding response: %v", err)
	}
}

// recupero l'immagine dall'input utente, la salvo in images e restituisco il path
func (rt *_router) setMyPhoto(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
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
	if filetype != jpegImg && filetype != pngImg {
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
	filePath := filepath.Join("/tmp/images", filename) // salva in cartella images
	// se la cartella non esiste la crea
	if err := os.MkdirAll("/tmp/images", 0755); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
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
	err = rt.db.SetUserPhoto(ctx.UserId, photoURL)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	var photo struct {
		Photo string `json:"photo"`
	}
	photo.Photo = photoURL
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(photo)
}
