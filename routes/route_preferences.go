package routes

import (
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"log"
	"net/http"
	"os"

	"github.com/boombuler/barcode"
	"github.com/boombuler/barcode/qr"
	"github.com/imagi-wa/kluck/data"
	"github.com/nfnt/resize"
	"golang.org/x/crypto/bcrypt"
)

func ChangePasswordHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("sid")
	if err == http.ErrNoCookie {
		http.Redirect(w, r, "/signin", http.StatusFound)
		return
	}
	session, err := data.SessionBySessionID(cookie.Value)
	if err != nil {
		http.Redirect(w, r, "/signin", http.StatusFound)
		return
	}
	ok := session.Verify()
	if !ok {
		http.Redirect(w, r, "/signin", http.StatusFound)
		return
	}
	user, err := data.UserByUserID(session.UserID)
	if err != nil {
		http.Redirect(w, r, "/signin", http.StatusFound)
		return
	}
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(r.PostFormValue("old_password")))
	if err != nil {
		d := data.Data{}
		d.UserInfo = user
		d.Alert = "Password is incorrect."
		generateHTML(w, d, "layout.html", "preferences_top_nav.html", "preferences_side_nav.html", "preferences.html")
		return
	}
	err = user.ChangePassword(r.PostFormValue("password"))
	if err != nil {
		http.Redirect(w, r, "/error", http.StatusFound)
		return
	}
	http.Redirect(w, r, "/preferences", http.StatusFound)
	return
}

func RemoveUserImageHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("sid")
	if err == http.ErrNoCookie {
		http.Redirect(w, r, "/signin", http.StatusFound)
		return
	}
	session, err := data.SessionBySessionID(cookie.Value)
	if err != nil {
		http.Redirect(w, r, "/signin", http.StatusFound)
		return
	}
	ok := session.Verify()
	if !ok {
		http.Redirect(w, r, "/signin", http.StatusFound)
		return
	}
	user, err := data.UserByUserID(session.UserID)
	if err != nil {
		http.Redirect(w, r, "/signin", http.StatusFound)
		return
	}
	err = user.ChangeUserImg("default.png")
	if err != nil {
		http.Redirect(w, r, "/error", http.StatusFound)
		return
	}
	http.Redirect(w, r, "/preferences", http.StatusFound)
	return
}

func UploadUserImageHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("sid")
	if err == http.ErrNoCookie {
		http.Redirect(w, r, "/signin", http.StatusFound)
		return
	}
	session, err := data.SessionBySessionID(cookie.Value)
	if err != nil {
		http.Redirect(w, r, "/signin", http.StatusFound)
		return
	}
	ok := session.Verify()
	if !ok {
		http.Redirect(w, r, "/signin", http.StatusFound)
		return
	}
	user, err := data.UserByUserID(session.UserID)
	if err != nil {
		http.Redirect(w, r, "/signin", http.StatusFound)
		return
	}
	err = r.ParseMultipartForm(10485760)
	if err != nil {
		log.Print(err.Error())
		http.Redirect(w, r, "/error", http.StatusFound)
		return
	}
	file, _, err := r.FormFile("upload_img")
	defer file.Close()
	if err != nil {
		log.Print(err.Error())
		http.Redirect(w, r, "error", http.StatusFound)
		return
	}
	img, t, err := image.Decode(file)
	if err != nil {
		log.Print(err.Error())
		http.Redirect(w, r, "/error", http.StatusFound)
		return
	}
	resizedImg := resize.Thumbnail(200, 200, img, resize.Lanczos3)
	filename := user.UserCode + "." + t
	path := "./web/images/users/" + filename
	out, err := os.Create(path)
	if err != nil {
		log.Print(err.Error())
		http.Redirect(w, r, "/error", http.StatusFound)
		return
	}
	switch t {
	case "jpeg":
		err = jpeg.Encode(out, resizedImg, nil)
		if err != nil {
			log.Print(err.Error())
			http.Redirect(w, r, "/error", http.StatusFound)
			return
		}
	case "png":
		err = png.Encode(out, resizedImg)
		if err != nil {
			log.Print(err.Error())
			http.Redirect(w, r, "/error", http.StatusFound)
			return
		}
	default:
		http.Redirect(w, r, "/error", http.StatusFound)
		return
	}
	err = user.ChangeUserImg(filename)
	if err != nil {
		http.Redirect(w, r, "/error", http.StatusFound)
		return
	}
	http.Redirect(w, r, "/preferences", http.StatusFound)
	return
}

func ChangeUserNameHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("sid")
	if err == http.ErrNoCookie {
		http.Redirect(w, r, "/signin", http.StatusFound)
		return
	}
	session, err := data.SessionBySessionID(cookie.Value)
	if err != nil {
		http.Redirect(w, r, "/signin", http.StatusFound)
		return
	}
	ok := session.Verify()
	if !ok {
		http.Redirect(w, r, "/signin", http.StatusFound)
		return
	}
	user, err := data.UserByUserID(session.UserID)
	if err != nil {
		http.Redirect(w, r, "/signin", http.StatusFound)
		return
	}
	err = user.ChangeUserName(r.PostFormValue("new_username"))
	if err != nil {
		http.Redirect(w, r, "/error", http.StatusFound)
		return
	}
	http.Redirect(w, r, "/preferences", http.StatusFound)
	return
}

func ShowOtpQrCodeHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("sid")
	if err == http.ErrNoCookie {
		http.Redirect(w, r, "/signin", http.StatusFound)
		return
	}
	session, err := data.SessionBySessionID(cookie.Value)
	if err != nil {
		http.Redirect(w, r, "/signin", http.StatusFound)
		return
	}
	ok := session.Verify()
	if !ok {
		http.Redirect(w, r, "/signin", http.StatusFound)
		return
	}
	user, err := data.UserByUserID(session.UserID)
	if err != nil {
		http.Redirect(w, r, "/signin", http.StatusFound)
		return
	}
	uri := fmt.Sprintf("otpauth://totp/Kluck:%s?secret=%s&issuer=Kluck", user.EmailAddress, user.OtpSecret)
	qrCode, _ := qr.Encode(uri, qr.M, qr.Auto)
	qrCode, _ = barcode.Scale(qrCode, 200, 200)
	png.Encode(w, qrCode)
	return
}

func EnableOtpHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("sid")
	if err == http.ErrNoCookie {
		http.Redirect(w, r, "/signin", http.StatusFound)
		return
	}
	session, err := data.SessionBySessionID(cookie.Value)
	if err != nil {
		http.Redirect(w, r, "/signin", http.StatusFound)
		return
	}
	ok := session.Verify()
	if !ok {
		http.Redirect(w, r, "/signin", http.StatusFound)
		return
	}
	user, err := data.UserByUserID(session.UserID)
	if err != nil {
		http.Redirect(w, r, "/signin", http.StatusFound)
		return
	}
	ok, err = user.VerifyOtp(r.PostFormValue("otp"))
	if err != nil {
		http.Redirect(w, r, "/error", http.StatusFound)
		return
	}
	if !ok {
		d := data.Data{}
		d.Alert = "Incorrect one-time password"
		generateHTML(w, d, "auth_layout.html", "otp_form.html")
		return
	}
	err = user.EnableOtp()
	if err != nil {
		http.Redirect(w, r, "/error", http.StatusFound)
		return
	}
	http.Redirect(w, r, "/preferences", http.StatusFound)
	return
}

func DisableOtpHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("sid")
	if err == http.ErrNoCookie {
		http.Redirect(w, r, "/signin", http.StatusFound)
		return
	}
	session, err := data.SessionBySessionID(cookie.Value)
	if err != nil {
		http.Redirect(w, r, "/signin", http.StatusFound)
		return
	}
	ok := session.Verify()
	if !ok {
		http.Redirect(w, r, "/signin", http.StatusFound)
		return
	}
	user, err := data.UserByUserID(session.UserID)
	if err != nil {
		http.Redirect(w, r, "/signin", http.StatusFound)
		return
	}
	err = user.DisableOtp()
	if err != nil {
		http.Redirect(w, r, "/error", http.StatusFound)
		return
	}
	http.Redirect(w, r, "/preferences", http.StatusFound)
	return
}

func SettingOtpFormHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("sid")
	if err == http.ErrNoCookie {
		http.Redirect(w, r, "/signin", http.StatusFound)
		return
	}
	session, err := data.SessionBySessionID(cookie.Value)
	if err != nil {
		http.Redirect(w, r, "/signin", http.StatusFound)
		return
	}
	ok := session.Verify()
	if !ok {
		http.Redirect(w, r, "/signin", http.StatusFound)
		return
	}
	user, err := data.UserByUserID(session.UserID)
	if err != nil {
		http.Redirect(w, r, "/signin", http.StatusFound)
		return
	}
	_, err = user.NewOtpSecret()
	if err != nil {
		http.Redirect(w, r, "/error", http.StatusFound)
		return
	}
	generateHTML(w, nil, "layout.html", "top_nav.html", "preferences_side_nav.html", "setting_otp.html")
	return
}

func PreferencesFormHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("sid")
	if err == http.ErrNoCookie {
		http.Redirect(w, r, "/signin", http.StatusFound)
		return
	}
	session, err := data.SessionBySessionID(cookie.Value)
	if err != nil {
		http.Redirect(w, r, "/signin", http.StatusFound)
		return
	}
	ok := session.Verify()
	if !ok {
		http.Redirect(w, r, "/signin", http.StatusFound)
		return
	}
	d := data.Data{}
	d.UserInfo, err = data.UserByUserID(session.UserID)
	if err != nil {
		http.Redirect(w, r, "/signin", http.StatusFound)
		return
	}
	generateHTML(w, d, "layout.html", "preferences_top_nav.html", "preferences_side_nav.html", "preferences.html")
	return
}
