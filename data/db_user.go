package data

import (
	"log"
	"time"

	"github.com/rs/xid"
)

type User struct {
	UserID       int
	UserName     string
	UserCode     string
	EmailAddress string
	PasswordHash string
	UserImg      string
	EnabledOtp   bool
	OtpSecret    string
	CreatedAt    time.Time
}

func (user *User) ChangeUserImg(newUserImg string) (err error) {
	stmt, err := Db.Prepare("UPDATE users SET user_img = $1 WHERE user_id = $2")
	if err != nil {
		log.Print(err.Error())
		return
	}
	defer stmt.Close()
	_, err = stmt.Exec(newUserImg, user.UserID)
	if err != nil {
		log.Print(err.Error())
	}
	return
}

func (user *User) ChangeUserName(newUserName string) (err error) {
	stmt, err := Db.Prepare("UPDATE users SET user_name = $1 WHERE user_id = $2")
	if err != nil {
		log.Print(err.Error())
		return
	}
	defer stmt.Close()
	_, err = stmt.Exec(newUserName, user.UserID)
	if err != nil {
		log.Print(err.Error())
	}
	return
}

func (user *User) New() (err error) {
	tx, err := Db.Begin()
	if err != nil {
		log.Print(err.Error())
		return
	}
	stmt, err := tx.Prepare("INSERT INTO users (user_name, user_code, email_address, password_hash) VALUES ($1, $2, $3, $4) RETURNING user_id")
	if err != nil {
		log.Print(err.Error())
		return
	}
	defer stmt.Close()
	guid := xid.New()
	user.UserCode = guid.String()
	err = stmt.QueryRow(user.UserName, user.UserCode, user.EmailAddress, user.PasswordHash).Scan(&user.UserID)
	if err != nil {
		log.Print(err.Error())
		tx.Rollback()
		return
	}
	stmt, err = tx.Prepare("INSERT INTO group_members (user_id, group_id, status_id) VALUES ($1, 'kluck-public', (SELECT status_id FROM status WHERE status_name = 'online'))")
	if err != nil {
		log.Print(err.Error())
		return
	}
	defer stmt.Close()
	_, err = stmt.Exec(user.UserID)
	if err != nil {
		log.Print(err.Error())
		tx.Rollback()
		return
	}
	stmt, err = tx.Prepare("UPDATE groups SET members_number = (SELECT COUNT (*) FROM group_members WHERE group_id = 'kluck-public') WHERE group_id = 'kluck-public'")
	if err != nil {
		log.Print(err.Error())
		return
	}
	defer stmt.Close()
	_, err = stmt.Exec()
	if err != nil {
		log.Print(err.Error())
		tx.Rollback()
		return
	}
	err = tx.Commit()
	if err != nil {
		log.Print(err.Error())
	}
	return
}

func UserByUserID(userID int) (user User, err error) {
	user = User{}
	stmt, err := Db.Prepare("SELECT * FROM users WHERE user_id = $1")
	if err != nil {
		log.Print(err.Error())
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(userID).Scan(&user.UserID, &user.UserName, &user.UserCode, &user.EmailAddress, &user.PasswordHash, &user.UserImg, &user.EnabledOtp, &user.OtpSecret, &user.CreatedAt)
	if err != nil {
		log.Print(err.Error())
	}
	return
}

func UserByUserCode(userCode string) (user User, err error) {
	user = User{}
	stmt, err := Db.Prepare("SELECT * FROM users WHERE user_code = $1")
	if err != nil {
		log.Print(err.Error())
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(userCode).Scan(&user.UserID, &user.UserName, &user.UserCode, &user.EmailAddress, &user.PasswordHash, &user.UserImg, &user.EnabledOtp, &user.OtpSecret, &user.CreatedAt)
	if err != nil {
		log.Print(err.Error())
	}
	return
}

func UserByEmail(email string) (user User, err error) {
	user = User{}
	stmt, err := Db.Prepare("SELECT * FROM users WHERE email_address = $1")
	if err != nil {
		log.Print(err.Error())
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(email).Scan(&user.UserID, &user.UserName, &user.UserCode, &user.EmailAddress, &user.PasswordHash, &user.UserImg, &user.EnabledOtp, &user.OtpSecret, &user.CreatedAt)
	if err != nil {
		log.Print(err.Error())
	}
	return
}
