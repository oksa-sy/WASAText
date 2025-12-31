package database

import (
	"crypto/rand"
	"encoding/hex"
	"time"
)

type Status string

const (
	RecievedSatus Status = "received"
	ReadStatus    Status = "read"
)

const RowNotFound = "not found"

type User struct {
	Id       string
	Username string
	Photo    string
}

type Conversation struct {
	Id        string
	Name      string
	Photo     string
	TimeStamp time.Time
	Preview   string
	IsGroup   bool
}

type Message struct {
	Id             string
	Content        string
	Username       string
	TimeStamp      time.Time
	Status         Status
	Reactions      []Reaction
	ReplyToId      string
	ReplyToContent string
}

type Group struct {
	Id      string
	Name    string
	Members []string
}

type Reaction struct {
	Emoji    string
	Username string
}

type Error struct {
	Code    string
	Message string
}

// creazione di stringhe da 16 caratteri randomizzate per creare gli id
func createId() (string, error) {
	bytes := make([]byte, 8)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
