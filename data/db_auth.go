package data

import (
	"encoding/base32"
	"fmt"
	"log"
	"time"

	"github.com/imagi-wa/kluck/tools/auth"
	"github.com/imagi-wa/totp"
	"golang.org/x/crypto/bcrypt"
)

func (user *User) ChangePassword(newPassword string) (err error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		log.Print(err.Error())
		return
	}
	stmt, err := Db.Prepare("UPDATE users SET password_hash = $1 WHERE user_id = $2")
	if err != nil {
		log.Print(err.Error())
		return
	}
	defer stmt.Close()
	_, err = stmt.Exec(string(hash), user.UserID)
	if err != nil {
		log.Print(err.Error())
	}
	return
}

func (session *Session) Authenticate() (err error) {
	stmt, err := Db.Prepare("UPDATE sessions SET authenticated = TRUE, expires_at = $1 WHERE session_id = $2")
	if err != nil {
		log.Print(err.Error())
		return
	}
	defer stmt.Close()
	now := time.Now()
	session.ExpiresAt = now.AddDate(0, 1, 0)
	_, err = stmt.Exec(session.ExpiresAt, session.SessionID)
	if err != nil {
		log.Print(err.Error())
	}
	return
}

func (user *User) NewOtpSecret() (key string, err error) {
	key = string(auth.GenerateRandomBytes(32))
	stmt, err := Db.Prepare("UPDATE users SET otp_secret = $1 WHERE user_id = $2")
	if err != nil {
		log.Print(err.Error())
		return
	}
	defer stmt.Close()
	_, err = stmt.Exec(key, user.UserID)
	if err != nil {
		log.Print(err.Error())
	}
	return
}

func (user *User) VerifyOtp(otp string) (ok bool, err error) {
	key, err := base32.StdEncoding.DecodeString(user.OtpSecret)
	if err != nil {
		log.Print(err.Error())
		return
	}
	keyBits := []byte(key)
	val := fmt.Sprintf("%06d", totp.TOTP(keyBits))
	if val == otp {
		ok = true
	} else {
		ok = false
	}
	return
}

func (user *User) EnableOtp() (err error) {
	stmt, err := Db.Prepare("UPDATE users SET enabled_otp = TRUE WHERE user_id = $1")
	if err != nil {
		log.Print(err.Error())
		return
	}
	defer stmt.Close()
	_, err = stmt.Exec(user.UserID)
	if err != nil {
		log.Print(err.Error())
	}
	return
}

func (user *User) DisableOtp() (err error) {
	stmt, err := Db.Prepare("UPDATE users SET enabled_otp = FALSE WHERE user_id = $1")
	if err != nil {
		log.Print(err.Error())
		return
	}
	defer stmt.Close()
	_, err = stmt.Exec(user.UserID)
	if err != nil {
		log.Print(err.Error())
	}
	return
}
