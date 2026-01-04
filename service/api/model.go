package api

import (
	"encoding/json"
	"net/http"
	"regexp"
	"time"
)

type Status string

const RowNotFound = "not found"

var nameRegex = regexp.MustCompile(`^[a-zA-Z0-9]+$`)

var groupNameRegex = regexp.MustCompile(`^[a-zA-Z0-9 _-]+$`)

func validateName(name string) bool {
	// controllo su lunghezza e pattern per verificare le regole definite nell'api
	if len(name) < 3 || len(name) > 16 || !nameRegex.MatchString(name) {
		return false
	}
	return true
}

func validateGroupName(name string) bool {
	// controllo su lunghezza e pattern per verificare le regole definite nell'api
	if len(name) < 3 || len(name) > 16 || !groupNameRegex.MatchString(name) {
		return false
	}
	return true
}

// inviare errori nel formato JSON corretto rispettando l'api
func errorJson(w http.ResponseWriter, statusCode int, code string, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	errResponse := Error{
		Code:    code,
		Message: message,
	}
	_ = json.NewEncoder(w).Encode(errResponse)
}

const (
	ErrCodeInternal   = "INTERNAL_ERROR"
	ErrCodeBadRequest = "BAD_REQUEST"
	ErrCodeConflict   = "CONFLICT"
	ErrCodeNotFound   = "NOT_FOUND"
)

const (
	RecievedSatus Status = "received"
	ReadStatus    Status = "read"
)

type User struct {
	Id       string `json:"identifier"`
	Username string `json:"username"`
	Photo    string `json:"photo,omitempty"`
}

type StartConversation struct {
	ParticipantId string `json:"participantId"`
}

type Conversation struct {
	Id             string    `json:"id"`
	Name           string    `json:"name"`
	Photo          string    `json:"photo"`
	TimeStamp      time.Time `json:"timestamp"`
	Preview        string    `json:"preview"`
	PreviewIsPhoto bool      `json:"previewIsPhoto"`
	IsGroup        bool      `json:"isGroup,omitempty"`
}

type CreateMessage struct {
	Content string `json:"content,omitempty"`
	ReplyTo string `json:"replyTo,omitempty"`
	Photo   string `json:"photo,omitempty"`
}

type Message struct {
	Id              string     `json:"id"`
	Content         string     `json:"content,omitempty"`
	Photo           string     `json:"photo,omitempty"`
	Username        string     `json:"username"`
	TimeStamp       time.Time  `json:"timestamp"`
	Reaction        []Reaction `json:"reaction,omitempty"`
	Status          Status     `json:"status,omitempty"`
	Forwarded       bool       `json:"forwarded,omitempty"`
	ReplyToId       string     `json:"replyToId,omitempty"`
	ReplyToContent  string     `json:"replyToContent,omitempty"`
	ReplyToUsername string     `json:"replyToUsername,omitempty"`
	ReplyToPhoto    string     `json:"replyToPhoto,omitempty"`
}

type Group struct {
	Id      string   `json:"id,omitempty"`
	Name    string   `json:"name"`
	Photo   string   `json:"photo,omitempty"`
	Members []string `json:"members"`
}

type CreateReaction struct {
	Emoji string `json:"emoji"`
}

type Reaction struct {
	Emoji    string `json:"emoji"`
	Username string `json:"username"`
}

type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}
