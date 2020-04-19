package connect

import (
	"github.com/imagi-wa/kluck/data"
	"golang.org/x/oauth2"
)

type UserInfoResponse struct {
	subject string `json:"sub"`
	Email   string `json:"email"`
}

type IDTokenHeader struct {
	Type      string `json:"typ"`
	Algorithm string `json:"alg"`
	KeyID     string `json:"kid"`
}

type JWKsResponse struct {
	KeySets []struct {
		KeyID     string `json:"kid"`
		KeyType   string `json:"kty"`
		Algorithm string `json:"alg"`
		Use       string `json:"use"`
		Modulus   string `json:"n"`
		Exponent  string `json:"e"`
	} `json:"keys"`
}

type IDTokenPayload struct {
	Issuer                        string   `json:"iss"`
	Subject                       string   `json:"sub"`
	Audience                      []string `json:"aud"`
	Expiration                    int      `json:"exp"`
	IssueAt                       int      `json:"iat"`
	AuthTime                      int      `json:"auth_time"`
	Nonce                         string   `json:"nonce"`
	AuthenticationMethodReference []string `json:"amr"`
	AccessTokenHash               string   `json:"at_hash"`
}

func YahooConnect() (conf *oauth2.Config) {
	appConf := data.GetAppConfig()
	conf = &oauth2.Config{
		ClientID:     appConf.Auth.YahooClientID,
		ClientSecret: appConf.Auth.YahooClientSecret,
		Endpoint: oauth2.Endpoint{
			AuthURL:   "https://auth.login.yahoo.co.jp/yconnect/v2/authorization",
			TokenURL:  "https://auth.login.yahoo.co.jp/yconnect/v2/token",
			AuthStyle: oauth2.AuthStyleInParams,
		},
		RedirectURL: "http://" + appConf.Server.Ipaddr + ":" + appConf.Server.Port + "/yahoo/callback",
		Scopes:      []string{"openid", "email"},
	}
	return
}
