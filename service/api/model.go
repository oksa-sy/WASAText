package api

import (
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

const (
	RecievedSatus Status = "received"
	ReadStatus    Status = "read"
)

type User struct {
	Id       string `json:"identifier"`
	Username string `json:"username"`
}

type StartConversation struct {
	ParticipantId string `json:"participantId"`
}

type Conversation struct {
	Id        string    `json:"id"`
	Name      string    `json:"name"`
	Photo     string    `json:"photo"`
	TimeStamp time.Time `json:"timestamp"`
	Preview   string    `json:"preview"`
}

type CreateMessage struct {
	Content string `json:"content"`
	ReplyTo string `json:"replyTo,omitempty"`
}

type Message struct {
	Id        string     `json:"id"`
	Content   string     `json:"content"`
	Username  string     `json:"username"`
	TimeStamp time.Time  `json:"timestamp"`
	Reaction  []Reaction `json:"reaction,omitempty"`
	Status    Status     `json:"status,omitempty"`
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

// da implimentare in futuro se necessario
type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}
