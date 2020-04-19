package connect

import (
	"github.com/imagi-wa/kluck/data"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	v2 "google.golang.org/api/oauth2/v2"
)

func GoogleConnect() (conf *oauth2.Config) {
	appConf := data.GetAppConfig()
	conf = &oauth2.Config{
		ClientID:     appConf.Auth.GoogleClientID,
		ClientSecret: appConf.Auth.GoogleClientSecret,
		Scopes:       []string{v2.UserinfoEmailScope},
		Endpoint:     google.Endpoint,
		RedirectURL:  "http://" + appConf.Server.Ipaddr + ":" + appConf.Server.Port + "/google/oauth2callback",
	}
	return
}
