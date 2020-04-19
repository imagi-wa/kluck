package data

import (
	"log"
	"time"

	"github.com/google/uuid"
)

type Session struct {
	SessionID     string
	UserID        int
	Authenticated bool
	ExpiresAt     time.Time
	CreatedAt     time.Time
}

func (user *User) NewSession() (session Session, err error) {
	session = Session{}
	err = DestroyAllSessionByUserID(user.UserID)
	if err != nil {
		return
	}
	stmt, err := Db.Prepare("INSERT INTO sessions (session_id, user_id) VALUES ($1, $2)")
	if err != nil {
		log.Print(err.Error())
		return
	}
	defer stmt.Close()
	u, err := uuid.NewRandom()
	if err != nil {
		log.Print(err.Error())
		return
	}
	session.SessionID = u.String()
	session.UserID = user.UserID
	_, err = stmt.Exec(session.SessionID, session.UserID)
	if err != nil {
		log.Print(err.Error())
	}
	return
}

func (session *Session) Destroy() (err error) {
	stmt, err := Db.Prepare("DELETE FROM sessions WHERE session_id = $1")
	if err != nil {
		log.Print(err.Error())
		return
	}
	defer stmt.Close()
	_, err = stmt.Exec(session.SessionID)
	if err != nil {
		log.Print(err.Error())
	}
	return
}

func DestroyAllSessionByUserID(userID int) (err error) {
	stmt, err := Db.Prepare("DELETE FROM sessions WHERE user_id = $1")
	if err != nil {
		log.Print(err.Error())
		return
	}
	defer stmt.Close()
	_, err = stmt.Exec(userID)
	return
}

func (session *Session) Verify() (ok bool) {
	if session.ExpiresAt.Before(time.Now()) || !session.Authenticated {
		ok = false
	} else {
		ok = true
	}
	return
}

func SessionBySessionID(sessionID string) (session Session, err error) {
	stmt, err := Db.Prepare("SELECT * FROM sessions WHERE session_id = $1")
	if err != nil {
		log.Print(err.Error())
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(sessionID).Scan(&session.SessionID, &session.UserID, &session.Authenticated, &session.ExpiresAt, &session.CreatedAt)
	if err != nil {
		log.Print(err.Error())
	}
	return
}
