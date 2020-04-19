package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/imagi-wa/kluck/data"
	"github.com/imagi-wa/kluck/routes"
)

func main() {
	r := mux.NewRouter()

	r.PathPrefix("/web/").Handler(http.StripPrefix("/web/", http.FileServer(http.Dir("web/"))))

	// Home
	r.HandleFunc("/", routes.IndexHandler).Methods("GET")
	r.HandleFunc("/{group_id:kluck-public|[0-9a-z]{20}}", routes.GroupIndexHandler).Methods("GET")

	// Auth
	r.HandleFunc("/signup", routes.SignupFormHandler).Methods("GET")
	r.HandleFunc("/signup", routes.SignupHandler).Methods("POST")
	r.HandleFunc("/signin", routes.SigninFormHandler).Methods("GET")
	r.HandleFunc("/signin", routes.PasswordSigninHandler).Methods("POST")
	r.HandleFunc("/signin/otp", routes.OtpFormHandler).Methods("GET")
	r.HandleFunc("/signin/otp", routes.OtpSigninHandler).Methods("POST")
	r.HandleFunc("/signout", routes.SignoutHandler).Methods("GET")
	r.HandleFunc("/google/oauth2", routes.GoogleOauth2Handler).Methods("GET")
	r.HandleFunc("/google/oauth2callback", routes.GoogleOauth2CallbackHandler).Methods("GET")
	r.HandleFunc("/yahoo/oauth2", routes.YahooOauth2Handler).Methods("GET")
	r.HandleFunc("/yahoo/callback", routes.YahooCallbackHandler).Methods("GET")

	// Error
	r.HandleFunc("/error", routes.ErrorHandler).Methods("GET")

	// Channel
	r.HandleFunc("/{group_id:kluck-public|[0-9a-z]{20}}/new", routes.NewChannelHandler).Methods("POST")
	r.HandleFunc("/{group_id:kluck-public|[0-9a-z]{20}}/{channel_id:[0-9A-Z]{26}}", routes.ReadChannelHandler).Methods("GET")
	r.HandleFunc("/{group_id:kluck-public|[0-9a-z]{20}}/{channel_id:[0-9A-Z]{26}}/join", routes.JoinChannelFormHandler).Methods("GET")
	r.HandleFunc("/{group_id:kluck-public|[0-9a-z]{20}}/{channel_id:[0-9A-Z]{26}}/join", routes.JoinChannelHandler).Methods("POST")
	r.HandleFunc("/{group_id:kluck-public|[0-9a-z]{20}}/{channel_id:[0-9A-Z]{26}}/new", routes.NewMessageHandler).Methods("POST")
	r.HandleFunc("/{group_id:kluck-public|[0-9a-z]{20}}/search/channel", routes.SearchChannelHandler).Methods("GET")

	// Direct Message
	r.HandleFunc("/{group_id:kluck-public|[0-9a-z]{20}}/{user_code:[0-9a-z]{20}}", routes.ReadDirectMessageHandler).Methods("GET")
	r.HandleFunc("/{group_id:kluck-public|[0-9a-z]{20}}/{user_code:[0-9a-z]{20}}/new", routes.NewDirectMessageHandler).Methods("POST")

	// Group
	r.HandleFunc("/new", routes.NewGroupHandler).Methods("POST")
	r.HandleFunc("/{group_id:kluck-public|[0-9a-z]{20}}/search/group", routes.SearchGroupHandler).Methods("GET")

	// Preferences
	r.HandleFunc("/preferences", routes.PreferencesFormHandler).Methods("GET")
	r.HandleFunc("/preferences/profile/username/change", routes.ChangeUserNameHandler).Methods("POST")
	r.HandleFunc("/preferences/profile/image/upload", routes.UploadUserImageHandler).Methods("POST")
	r.HandleFunc("/preferences/profile/image/remove", routes.RemoveUserImageHandler).Methods("GET")
	r.HandleFunc("/preferences/security/password/change", routes.ChangePasswordHandler).Methods("POST")
	r.HandleFunc("/preferences/security/otp", routes.SettingOtpFormHandler).Methods("GET")
	r.HandleFunc("/preferences/security/otp", routes.SettingOtpFormHandler).Methods("GET")
	r.HandleFunc("/preferences/security/otp/show", routes.ShowOtpQrCodeHandler).Methods("GET")
	r.HandleFunc("/preferences/security/otp/enable", routes.EnableOtpHandler).Methods("POST")
	r.HandleFunc("/preferences/security/otp/disable", routes.DisableOtpHandler).Methods("GET")

	raw, err := ioutil.ReadFile("./config/config.json")
	if err != nil {
		log.Panic(err)
		os.Exit(1)
	}
	var appConf data.AppConfig
	json.Unmarshal(raw, &appConf)
	server := &http.Server{
		Addr:         appConf.Server.Ipaddr + ":" + appConf.Server.Port,
		Handler:      r,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Fatal(server.ListenAndServe())
}
