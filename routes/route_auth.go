package routes

import (
	"bytes"
	"context"
	"crypto"
	"crypto/rsa"
	"database/sql"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"math/big"
	"net/http"
	"strings"

	"github.com/imagi-wa/kluck/data"
	"github.com/imagi-wa/kluck/tools/auth"
	"github.com/imagi-wa/kluck/tools/auth/connect"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/oauth2"
	v2 "google.golang.org/api/oauth2/v2"
)

func YahooOauth2Handler(w http.ResponseWriter, r *http.Request) {
	conf := connect.YahooConnect()
	state := auth.GenerateRandomString(64)
	nonce := auth.GenerateRandomString(64)
	stateCookie := &http.Cookie{
		Name:     "state",
		Value:    state,
		HttpOnly: true,
	}
	http.SetCookie(w, stateCookie)
	nonceCookie := &http.Cookie{
		Name:     "nonce",
		Value:    nonce,
		HttpOnly: true,
	}
	http.SetCookie(w, nonceCookie)
	var NonceOpt oauth2.AuthCodeOption = oauth2.SetAuthURLParam("nonce", nonce)
	url := conf.AuthCodeURL(state, NonceOpt, oauth2.ApprovalForce)
	http.Redirect(w, r, url, http.StatusFound)
	return
}

func YahooCallbackHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	conf := connect.YahooConnect()
	storedState, err := r.Cookie("state")
	if err != nil {
		log.Print(err.Error())
		http.Redirect(w, r, "/error", http.StatusBadRequest)
		return
	}
	stateCookie := &http.Cookie{
		Name:   "state",
		MaxAge: -1,
	}
	http.SetCookie(w, stateCookie)
	query := r.URL.Query()
	stateQuery, ok := query["state"]
	if !ok {
		http.Redirect(w, r, "/error", http.StatusBadRequest)
		return
	}
	state := stateQuery[0]
	if state != storedState.Value {
		log.Print("State does not match stored one")
		http.Redirect(w, r, "/error", http.StatusBadRequest)
		return
	}
	codeQuery, ok := query["code"]
	if !ok {
		log.Print("Code query not found")
		http.Redirect(w, r, "/error", http.StatusBadRequest)
		return
	}
	code := codeQuery[0]
	tok, err := conf.Exchange(ctx, code)
	if err != nil {
		log.Print(err.Error())
		http.Redirect(w, r, "/error", http.StatusFound)
		return
	}
	rawIDTok, ok := tok.Extra("id_token").(string)
	if !ok {
		log.Print("Missing token.")
		http.Redirect(w, r, "/error", http.StatusInternalServerError)
		return
	}
	idTokenParts := strings.SplitN(rawIDTok, ".", 3)
	header, err := base64.RawURLEncoding.DecodeString(idTokenParts[0])
	if err != nil {
		log.Print(err.Error())
		http.Redirect(w, r, "/error", http.StatusInternalServerError)
		return
	}
	// Store an ID Token header to the struct.
	var idTokenHeader connect.IDTokenHeader
	err = json.Unmarshal(header, &idTokenHeader)
	if err != nil {
		log.Print(err.Error())
		http.Redirect(w, r, "/error", http.StatusInternalServerError)
		return
	}
	// Validate the typ-value.
	if idTokenHeader.Type != "JWT" {
		log.Print("Invalid id token type.")
		http.Redirect(w, r, "/error", http.StatusInternalServerError)
		return
	}
	// Validate the alg-value.
	if idTokenHeader.Algorithm != "RS256" {
		log.Print("Invalid id token algorithm.")
		http.Redirect(w, r, "/error", http.StatusInternalServerError)
		return
	}
	// JWKs request.
	jwksResponse, err := http.Get("https://auth.login.yahoo.co.jp/yconnect/v2/jwks")
	if err != nil {
		log.Print(err.Error())
		http.Redirect(w, r, "/error", http.StatusInternalServerError)
		return
	}
	defer func() {
		_, err = io.Copy(ioutil.Discard, jwksResponse.Body)
		if err != nil {
			log.Panic(err)
		}
		err = jwksResponse.Body.Close()
		if err != nil {
			log.Panic(err)
		}
	}()
	// Store JWKs response to the struct.
	var jwksData connect.JWKsResponse
	err = json.NewDecoder(jwksResponse.Body).Decode(&jwksData)
	if err != nil {
		log.Print(err.Error())
		http.Redirect(w, r, "/error", http.StatusInternalServerError)
		return
	}
	// Extract modulus-value and exponent-value.
	var modulus, exponent string
	for _, keySet := range jwksData.KeySets {
		if keySet.KeyID == idTokenHeader.KeyID {
			if keySet.KeyType != "RSA" || keySet.Algorithm != idTokenHeader.Algorithm || keySet.Use != "sig" {
				log.Print("Invalid KeySet(kid, alg or use).")
				http.Redirect(w, r, "/error", http.StatusInternalServerError)
				return
			}
			modulus = keySet.Modulus
			exponent = keySet.Exponent
			break
		}
	}
	if modulus == "" || exponent == "" {
		log.Print("Failed to extract modulus-value or exponent-value.")
		http.Redirect(w, r, "/error", http.StatusInternalServerError)
		return
	}
	// Base64URLDecode modulus and exponent.
	decodedModulus, err := base64.RawURLEncoding.DecodeString(modulus)
	if err != nil {
		log.Print(err.Error())
		http.Redirect(w, r, "/error", http.StatusInternalServerError)
		return
	}
	decodedExponent, err := base64.RawURLEncoding.DecodeString(exponent)
	if err != nil {
		log.Print(err.Error())
		http.Redirect(w, r, "/error", http.StatusInternalServerError)
		return
	}
	// Convert string to uint64.
	var exponentBytes []byte
	if len(decodedExponent) < 8 {
		exponentBytes = make([]byte, 8-len(decodedExponent), 8)
		exponentBytes = append(exponentBytes, decodedExponent...)
	} else {
		exponentBytes = decodedExponent
	}
	reader := bytes.NewReader(exponentBytes)
	var e uint64
	err = binary.Read(reader, binary.BigEndian, &e)
	if err != nil {
		log.Print(err.Error())
		http.Redirect(w, r, "/error", http.StatusInternalServerError)
		return
	}
	// Generate public-key from modulus and exponent.
	publicKey := rsa.PublicKey{
		N: big.NewInt(0).SetBytes(decodedModulus),
		E: int(e),
	}
	// Base64URLDecode ID Token Signature.
	decodedSignature, err := base64.RawURLEncoding.DecodeString(idTokenParts[2])
	if err != nil {
		log.Print(err.Error())
		http.Redirect(w, r, "/error", http.StatusInternalServerError)
		return
	}
	// Generate a hash.
	hash := crypto.Hash.New(crypto.SHA256)
	_, err = hash.Write([]byte(idTokenParts[0] + "." + idTokenParts[1]))
	if err != nil {
		log.Print(err.Error())
		http.Redirect(w, r, "/error", http.StatusInternalServerError)
		return
	}
	hashed := hash.Sum(nil)
	// Validate ID Token Signature.
	err = rsa.VerifyPKCS1v15(&publicKey, crypto.SHA256, hashed, decodedSignature)
	if err != nil {
		log.Print(err.Error())
		http.Redirect(w, r, "/error", http.StatusInternalServerError)
		return
	}
	// Base64URLDecode ID Token Payload.
	decodedPayload, err := base64.RawURLEncoding.DecodeString(idTokenParts[1])
	if err != nil {
		log.Print(err.Error())
		http.Redirect(w, r, "/error", http.StatusInternalServerError)
		return
	}
	// Store ID Token Payload to the struct.
	var idTokenPayload connect.IDTokenPayload
	err = json.Unmarshal(decodedPayload, &idTokenPayload)
	if err != nil {
		log.Print(err.Error())
		http.Redirect(w, r, "/error", http.StatusInternalServerError)
		return
	}
	// Validate issuer-value.
	if idTokenPayload.Issuer != "https://auth.login.yahoo.co.jp/yconnect/v2" {
		log.Print("Mismatched issuer-value.")
		http.Redirect(w, r, "/error", http.StatusInternalServerError)
		return
	}
	// Validate audience-value.
	var isValidAudience bool
	for _, audience := range idTokenPayload.Audience {
		if audience == conf.ClientID {
			isValidAudience = true
			break
		}
	}
	if !isValidAudience {
		log.Print("Mismatched audience-value.")
		http.Redirect(w, r, "/error", http.StatusInternalServerError)
		return
	}
	// Validate Nonce-value.
	storedNonce, err := r.Cookie("nonce")
	if err != nil {
		log.Print(err.Error())
		http.Redirect(w, r, "/error", http.StatusBadRequest)
		return
	}
	nonceCookie := &http.Cookie{
		Name:   "nonce",
		MaxAge: -1,
	}
	http.SetCookie(w, nonceCookie)
	if idTokenPayload.Nonce != storedNonce.Value {
		log.Print("Nonce does not match stored one.")
		http.Redirect(w, r, "/error", http.StatusInternalServerError)
		return
	}

	userInfoRequest, err := http.NewRequest("POST", "https://userinfo.yahooapis.jp/yconnect/v2/attribute", nil)
	if err != nil {
		log.Print(err.Error())
		http.Redirect(w, r, "/error", http.StatusBadRequest)
		return
	}
	userInfoRequest.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	userInfoRequest.Header.Set("Authorization", "Bearer "+tok.AccessToken)
	userInfoResponse, err := http.DefaultClient.Do(userInfoRequest)
	if err != nil {
		log.Print(err.Error())
		http.Redirect(w, r, "/error", http.StatusInternalServerError)
		return
	}
	defer func() {
		_, err = io.Copy(ioutil.Discard, userInfoResponse.Body)
		if err != nil {
			log.Panic(err)
		}
		err = userInfoResponse.Body.Close()
		if err != nil {
			log.Panic(err)
		}
	}()
	var userInfoData connect.UserInfoResponse
	err = json.NewDecoder(userInfoResponse.Body).Decode(&userInfoData)
	if err != nil {
		log.Print(err.Error())
		http.Redirect(w, r, "/error", http.StatusInternalServerError)
		return
	}
	user, err := data.UserByEmail(userInfoData.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			generateHTML(w, userInfoData.Email, "auth_layout.html", "signup_form.html")
			return
		}
		http.Redirect(w, r, "/signup", http.StatusFound)
		return
	}
	session, err := user.NewSession()
	if err != nil {
		http.Redirect(w, r, "/signin", http.StatusFound)
		return
	}
	cookie := http.Cookie{
		Name:     "sid",
		Value:    session.SessionID,
		Expires:  session.ExpiresAt,
		HttpOnly: true,
		Path:     "/",
	}
	http.SetCookie(w, &cookie)
	if user.EnabledOtp {
		http.Redirect(w, r, "/signin/otp", http.StatusFound)
		return
	}
	err = session.Authenticate()
	if err != nil {
		http.Redirect(w, r, "/error", http.StatusFound)
		return
	}
	http.Redirect(w, r, "/", http.StatusFound)
	return
}

func GoogleOauth2Handler(w http.ResponseWriter, r *http.Request) {
	conf := connect.GoogleConnect()
	var url = conf.AuthCodeURL("")
	http.Redirect(w, r, url, http.StatusFound)
	return
}

func GoogleOauth2CallbackHandler(w http.ResponseWriter, r *http.Request) {
	conf := connect.GoogleConnect()
	code := r.URL.Query()["code"]
	if code == nil || len(code) == 0 {
		http.Redirect(w, r, "/error", http.StatusFound)
		return
	}
	ctx := context.Background()
	token, err := conf.Exchange(ctx, code[0])
	if err != nil {
		log.Print(err.Error())
		http.Redirect(w, r, "/error", http.StatusFound)
		return
	}
	client := conf.Client(ctx, token)
	service, err := v2.New(client)
	if err != nil {
		log.Print(err.Error())
		http.Redirect(w, r, "/error", http.StatusFound)
		return
	}
	info, err := service.Userinfo.Get().Do()
	if err != nil {
		log.Print(err.Error())
		http.Redirect(w, r, "/error", http.StatusFound)
		return
	}
	user, err := data.UserByEmail(info.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			generateHTML(w, info.Email, "auth_layout.html", "signup_form.html")
			return
		}
		http.Redirect(w, r, "/signup", http.StatusFound)
		return
	}
	session, err := user.NewSession()
	if err != nil {
		http.Redirect(w, r, "/signin", http.StatusFound)
		return
	}
	cookie := http.Cookie{
		Name:     "sid",
		Value:    session.SessionID,
		Expires:  session.ExpiresAt,
		HttpOnly: true,
		Path:     "/",
	}
	http.SetCookie(w, &cookie)
	if user.EnabledOtp {
		http.Redirect(w, r, "/signin/otp", http.StatusFound)
		return
	}
	err = session.Authenticate()
	if err != nil {
		http.Redirect(w, r, "/error", http.StatusFound)
		return
	}
	http.Redirect(w, r, "/", http.StatusFound)
	return
}

func SignupHandler(w http.ResponseWriter, r *http.Request) {
	d := data.Data{}
	user, err := data.UserByEmail(r.PostFormValue("email"))
	if err != nil {
		if err == sql.ErrNoRows { // Registeration
			// Generate a password hash
			hash, err := bcrypt.GenerateFromPassword([]byte(r.PostFormValue("password")), bcrypt.DefaultCost)
			if err != nil {
				log.Print(err.Error())
				http.Redirect(w, r, "/error", http.StatusFound)
				return
			}
			user = data.User{
				UserName:     r.PostFormValue("name"),
				EmailAddress: r.PostFormValue("email"),
				PasswordHash: string(hash),
			}
			err = user.New()
			if err != nil {
				http.Redirect(w, r, "/error", http.StatusFound)
				return
			}
			// Session start.
			session, err := user.NewSession()
			if err != nil { // Failed to start the session.
				http.Redirect(w, r, "/signin", http.StatusFound)
				return
			}
			// Set a Cookie.
			cookie := http.Cookie{
				Name:     "sid",
				Value:    session.SessionID,
				Expires:  session.ExpiresAt,
				HttpOnly: true,
				Path:     "/",
			}
			http.SetCookie(w, &cookie)
			err = session.Authenticate()
			if err != nil {
				http.Redirect(w, r, "/error", http.StatusFound)
				return
			}
			http.Redirect(w, r, "/", http.StatusFound)
			return
		} else { // Other errors
			http.Redirect(w, r, "/error", http.StatusFound)
			return
		}
	} else { // User already exists.
		d.Alert = "This E-mail Address is used."
	}
	generateHTML(w, d, "auth_layout.html", "signup_form.html")
	return
}

func PasswordSigninHandler(w http.ResponseWriter, r *http.Request) {
	d := data.Data{}
	user, err := data.UserByEmail(r.PostFormValue("email"))
	if err != nil { // No user or other errors.
		if err == sql.ErrNoRows {
			d.Alert = "E-mail Address or password is incorrect."
			generateHTML(w, d, "auth_layout.html", "signin_form.html")
		} else {
			http.Redirect(w, r, "/error", http.StatusFound)
		}
		return
	}
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(r.PostFormValue("password")))
	if err != nil { // Password is incorrect.
		d.Alert = "E-mail Address or password is incorrect."
		generateHTML(w, d, "auth_layout.html", "signin_form.html")
		return
	}
	session, err := user.NewSession()
	if err != nil { // Failed to start the session.
		http.Redirect(w, r, "/signin", http.StatusFound)
		return
	}
	// Set a Cookie.
	cookie := http.Cookie{
		Name:     "sid",
		Value:    session.SessionID,
		Expires:  session.ExpiresAt,
		HttpOnly: true,
		Path:     "/",
	}
	http.SetCookie(w, &cookie)
	if user.EnabledOtp {
		http.Redirect(w, r, "/signin/otp", http.StatusFound)
		return
	}
	err = session.Authenticate()
	if err != nil {
		http.Redirect(w, r, "/error", http.StatusFound)
		return
	}
	http.Redirect(w, r, "/", http.StatusFound)
	return
}

func OtpSigninHandler(w http.ResponseWriter, r *http.Request) {
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
	user, err := data.UserByUserID(session.UserID)
	if err != nil {
		http.Redirect(w, r, "/signin", http.StatusFound)
		return
	}
	ok, err := user.VerifyOtp(r.PostFormValue("otp"))
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
	err = session.Authenticate()
	if err != nil {
		http.Redirect(w, r, "/error", http.StatusFound)
		return
	}
	http.Redirect(w, r, "/", http.StatusFound)
	return
}

func SignoutHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("sid")
	if err != http.ErrNoCookie { // Delete the cookie and destroy the session.
		cookie.MaxAge = -1
		http.SetCookie(w, cookie)
		session := data.Session{
			SessionID: cookie.Value,
		}
		err := session.Destroy()
		if err != nil { // Failed to destroy the session.
			http.Redirect(w, r, "/error", http.StatusFound)
			return
		}
	}
	http.Redirect(w, r, "/signin", http.StatusFound)
	return
}

func OtpFormHandler(w http.ResponseWriter, r *http.Request) {
	generateHTML(w, nil, "auth_layout.html", "otp_form.html")
	return
}

func SigninFormHandler(w http.ResponseWriter, r *http.Request) {
	generateHTML(w, nil, "auth_layout.html", "signin_form.html")
	return
}

func SignupFormHandler(w http.ResponseWriter, r *http.Request) {
	generateHTML(w, nil, "auth_layout.html", "signup_form.html")
	return
}
