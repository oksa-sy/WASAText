/*
Package database is the middleware between the app database and the code. All data (de)serialization (save/load) from a
persistent database are handled here. Database specific logic should never escape this package.

To use this package you need to apply migrations to the database if needed/wanted, connect to it (using the database
data source string from config), and then initialize an instance of AppDatabase from the DB connection.

For example, this code adds a parameter in `webapi` executable for the database data source string (add it to the
main.WebAPIConfiguration structure):

	DB struct {
		Filestring string `conf:""`
	}

This is an example on how to migrate the DB and connect to it:

	// Start Database
	logger.Println("initializing database support")
	db, err := sql.Open("sqlite3", "./foo.db")
	if err != nil {
		logger.WithError(err).Error("error opening SQLite DB")
		return fmt.Errorf("opening SQLite: %w", err)
	}
	defer func() {
		logger.Debug("database stopping")
		_ = db.Close()
	}()

Then you can initialize the AppDatabase and pass it to the api package.
*/
package database

import (
	"database/sql"
	"errors"
	"fmt"
)

// AppDatabase is the high level interface for the DB
// definizione di tutte le operazioni sql
type AppDatabase interface {
	// login
	CreateUser(name string) (string, error)
	// users
	SearchUser(requestId string) (bool, error)
	GetUsersByName(name string) ([]User, error)
	SetUserName(id string, name string) error
	SetUserPhoto(id string, photo string) error
	// conversations
	GetAllConversations(userId string) ([]Conversation, error)
	CreateNewConversation(participantId string, userId string) (Conversation, error)
	CreateMessage(conversationId string, userId string, content string, replyTo string) (Message, error)
	GetConversation(conversationId string, userId string) ([]Message, error)
	// groups
	CreateNewGroup(user string, name string, members []string) (Group, error)
	UpdateGroupName(name string, groupId string, userId string) (string, error)
	SetNewGroupPhoto(groupId string, photo string, userId string) error
	AddUserToGroup(memberId string, userId string, groupId string) (string, error)
	LeaveGroup(userId string, groupId string) error
	// messages
	ForwardMessage(userId string, conversationId string, messageId string) (Message, error)
	DeleteMessage(userId string, messageId string) error
	CreateReaction(reaction string, messageId string, userId string) (Reaction, error)
	DeleteReaction(reaction string, userId string, messageId string) error

	Ping() error
}

type appdbimpl struct {
	c *sql.DB
}

// New returns a new instance of AppDatabase based on the SQLite connection `db`.
// `db` is required - an error will be returned if `db` is `nil`.
// creazione delle tabelle necessarie per le operazioni
func New(db *sql.DB) (AppDatabase, error) {
	// controlla se il database esiste, altrimenti da errore
	if db == nil {
		return nil, errors.New("database is required when building a AppDatabase")
	}
	// bisogna abilitare le foreign key per poterle usare in mysql
	_, err := db.Exec("PRAGMA foreign_keys = ON")
	if err != nil {
		return nil, err
	}
	// Check if table exists. If not, the database is empty, and we need to create the structure
	var tableName string
	err = db.QueryRow(`SELECT name FROM sqlite_master WHERE type='table' AND name='users';`).Scan(&tableName)

	if errors.Is(err, sql.ErrNoRows) {
		sqlStmt := `
		CREATE TABLE users (
			id TEXT NOT NULL PRIMARY KEY,
			username TEXT NOT NULL UNIQUE,
			photo TEXT
		);
		CREATE TABLE conversations (
			id TEXT NOT NULL PRIMARY KEY,
			is_group BOOLEAN DEFAULT 0 NOT NULL,
			name TEXT,
			photo TEXT
		);
		CREATE TABLE messages (
			id TEXT NOT NULL PRIMARY KEY,
			conversation_id TEXT NOT NULL,
			sender_id TEXT NOT NULL,
			content TEXT NOT NULL,
			timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,  
			reply_to TEXT,
			status TEXT CHECK(status IN ('recieved', 'read', 'error')) DEFAULT 'recieved',
			FOREIGN KEY (conversation_id) REFERENCES conversations(id) ON DELETE CASCADE,
			FOREIGN KEY (sender_id) REFERENCES users(id)
		);
		CREATE TABLE reactions (
			message_id TEXT NOT NULL,
			user_id TEXT NOT NULL,
			reaction TEXT NOT NULL,
			PRIMARY KEY (message_id, user_id),
			FOREIGN KEY (message_id) REFERENCES messages(id) ON DELETE CASCADE,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		);
		CREATE TABLE participants (
			conversation_id TEXT NOT NULL,
			user_id TEXT NOT NULL,
			PRIMARY KEY (conversation_id, user_id),
			FOREIGN KEY (conversation_id) REFERENCES conversations(id) ON DELETE CASCADE,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		);
		`
		_, err = db.Exec(sqlStmt)
		if err != nil {
			return nil, fmt.Errorf("error creating database structure: %w", err)
		}
	}

	return &appdbimpl{
		c: db,
	}, nil
}

func (db *appdbimpl) Ping() error {
	return db.c.Ping()
}
