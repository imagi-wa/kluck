package data

import (
	"database/sql"
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"time"

	_ "github.com/lib/pq"
)

var (
	Db  *sql.DB
	err error
)

type CType struct {
	Type_id   int
	Type_name string
}

type Member struct {
	MemberID  int
	UserID    int
	ChannelID string
	JoinedAt  time.Time
}

type Data struct {
	Channels             []Channel
	DirectMessages       []DirectMessage
	ConversationUserCode string
	Messages             []Message
	ChannelsInfo         []Channel
	DirectMessagesInfo   []User
	UserInfo             User
	Groups               []Group
	GroupsInfo           []Group
	CurrentGroupInfo     Group
	CurrentGroupMembers  []User
	Alert                string
}

type AppConfig struct {
	Server struct {
		Ipaddr string `json:"ipaddr"`
		Port   string `json:"port"`
	}
	Database struct {
		Ipaddr   string `json:"ipaddr`
		Name     string `json:"name"`
		Port     string `json:"port"`
		User     string `json:"user"`
		Password string `json:"password"`
		SSLmode  string `json:"sslmode"`
	}
	Auth struct {
		GoogleClientID     string `json:"google_client_ID"`
		GoogleClientSecret string `json:"google_client_secret"`
		YahooClientID      string `json:"yahoo_client_ID"`
		YahooClientSecret  string `json:"yahoo_client_secret"`
	}
}

func init() {
	conf := GetAppConfig()
	connection := "host=" + conf.Database.Ipaddr + " port=" + conf.Database.Port + " user=" + conf.Database.User + " password=" + conf.Database.Password + " dbname=" + conf.Database.Name + " sslmode=" + conf.Database.SSLmode

	Db, err = sql.Open("postgres", connection)
	if err != nil {
		log.Print(err.Error())
	}
	return
}

func GetAppConfig() (conf AppConfig) {
	raw, err := ioutil.ReadFile("./config/config.json")
	if err != nil {
		log.Panic(err)
		os.Exit(1)
	}
	json.Unmarshal(raw, &conf)
	return
}
