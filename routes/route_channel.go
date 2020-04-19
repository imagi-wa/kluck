package routes

import (
	"database/sql"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/imagi-wa/kluck/data"
)

func SearchChannelHandler(w http.ResponseWriter, r *http.Request) {
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
	vars := mux.Vars(r)
	ok, err = user.IsGroupMember(vars["group_id"])
	if err != nil {
		http.Redirect(w, r, "/error", http.StatusFound)
		return
	} else if !ok {
		http.Redirect(w, r, "/signin", http.StatusFound)
		return
	}
	d := data.Data{}
	keyword := r.URL.Query().Get("keyword")
	d.Channels, err = data.SearchChannelByKeyword(keyword, vars["group_id"])
	if err != nil {
		if err == sql.ErrNoRows {
			d.Alert = "Not found..."
		} else {
			http.Redirect(w, r, "/error", http.StatusFound)
			return
		}
	}
	d.ChannelsInfo, err = user.ChannelsUserIsJoining(vars["group_id"])
	if err != nil {
		http.Redirect(w, r, "/error", http.StatusFound)
		return
	}
	d.DirectMessagesInfo, err = user.DirectMessageUsers(vars["group_id"])
	if err != nil {
		http.Redirect(w, r, "/error", http.StatusFound)
		return
	}
	d.UserInfo = user
	d.GroupsInfo, err = user.GroupsUserIsJoining()
	if err != nil {
		http.Redirect(w, r, "/error", http.StatusFound)
		return
	}
	d.CurrentGroupInfo, err = data.GroupByGroupID(vars["group_id"])
	if err != nil {
		http.Redirect(w, r, "/error", http.StatusFound)
		return
	}
	d.CurrentGroupMembers, err = d.CurrentGroupInfo.GroupMembers()
	if err != nil {
		http.Redirect(w, r, "/error", http.StatusFound)
		return
	}
	generateHTML(w, d, "layout.html", "top_nav.html", "side_nav.html", "index.html")
	return
}

func NewChannelHandler(w http.ResponseWriter, r *http.Request) {
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
	vars := mux.Vars(r)
	ok, err = user.IsGroupMember(vars["group_id"])
	if err != nil {
		http.Redirect(w, r, "/error", http.StatusFound)
		return
	} else if !ok {
		http.Redirect(w, r, "/signin", http.StatusFound)
		return
	}
	channel := data.Channel{}
	channel.ChannelName = r.PostFormValue("name")
	channel.ChannelTopic = r.PostFormValue("topic")
	channel.UserID = user.UserID
	channel.GroupID = vars["group_id"]
	channel.TypeID, err = strconv.Atoi(r.PostFormValue("type"))
	if err != nil {
		log.Print(err.Error())
		http.Redirect(w, r, "/error", http.StatusFound)
		return
	}
	err = channel.New()
	if err != nil {
		http.Redirect(w, r, "/error", http.StatusFound)
		return
	}
	http.Redirect(w, r, "/"+vars["group_id"]+"/"+channel.ChannelID, http.StatusFound)
	return
}

func JoinChannelHandler(w http.ResponseWriter, r *http.Request) {
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
	vars := mux.Vars(r)
	ok, err = user.IsGroupMember(vars["group_id"])
	if err != nil {
		http.Redirect(w, r, "/error", http.StatusFound)
		return
	} else if !ok {
		http.Redirect(w, r, "/signin", http.StatusFound)
		return
	}
	err = user.JoinChannel(vars["channel_id"])
	if err != nil {
		http.Redirect(w, r, "/error", http.StatusFound)
		return
	}
	http.Redirect(w, r, "/"+vars["group_id"]+"/"+vars["channel_id"], http.StatusFound)
	return
}

func JoinChannelFormHandler(w http.ResponseWriter, r *http.Request) {
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
	vars := mux.Vars(r)
	ok, err = user.IsGroupMember(vars["group_id"])
	if err != nil {
		http.Redirect(w, r, "/error", http.StatusFound)
		return
	} else if !ok {
		http.Redirect(w, r, "/signin", http.StatusFound)
		return
	}
	channel, err := user.ChannelByChannelID(vars["channel_id"])
	if err != nil {
		http.Redirect(w, r, "/error", http.StatusFound)
		return
	}
	d := data.Data{}
	d.Channels = append(d.Channels, channel)
	d.ChannelsInfo, err = user.ChannelsUserIsJoining(vars["group_id"])
	if err != nil {
		http.Redirect(w, r, "/error", http.StatusFound)
		return
	}
	d.DirectMessagesInfo, err = user.DirectMessageUsers(vars["group_id"])
	if err != nil {
		http.Redirect(w, r, "/error", http.StatusFound)
		return
	}
	d.UserInfo = user
	d.GroupsInfo, err = user.GroupsUserIsJoining()
	if err != nil {
		http.Redirect(w, r, "/error", http.StatusFound)
		return
	}
	d.CurrentGroupInfo, err = data.GroupByGroupID(vars["group_id"])
	if err != nil {
		http.Redirect(w, r, "/error", http.StatusFound)
		return
	}
	d.CurrentGroupMembers, err = d.CurrentGroupInfo.GroupMembers()
	if err != nil {
		http.Redirect(w, r, "/error", http.StatusFound)
		return
	}
	generateHTML(w, d, "layout.html", "top_nav.html", "side_nav.html", "join_channel_form.html")
	return
}

func ReadChannelHandler(w http.ResponseWriter, r *http.Request) {
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
	vars := mux.Vars(r)
	ok, err = user.IsGroupMember(vars["group_id"])
	if err != nil {
		http.Redirect(w, r, "/error", http.StatusFound)
		return
	} else if !ok {
		http.Redirect(w, r, "/signin", http.StatusFound)
		return
	}
	ok, err = user.IsChannelMember(vars["channel_id"])
	if err != nil {
		http.Redirect(w, r, "/error", http.StatusFound)
		return
	} else if !ok {
		http.Redirect(w, r, "/"+vars["group_id"]+"/"+vars["channel_id"]+"/join", http.StatusFound)
		return
	}
	d := data.Data{}
	channel, err := user.ChannelByChannelID(vars["channel_id"])
	if err != nil {
		http.Redirect(w, r, "/error", http.StatusFound)
		return
	}
	d.Channels = append(d.Channels, channel)
	d.Messages, err = channel.Messages()
	if err != nil {
		if err == sql.ErrNoRows {
			d.Alert = "No messages..."
		} else {
			http.Redirect(w, r, "/error", http.StatusFound)
			return
		}
	}
	d.ChannelsInfo, err = user.ChannelsUserIsJoining(vars["group_id"])
	if err != nil {
		http.Redirect(w, r, "/error", http.StatusFound)
		return
	}
	d.DirectMessagesInfo, err = user.DirectMessageUsers(vars["group_id"])
	if err != nil {
		http.Redirect(w, r, "/error", http.StatusFound)
		return
	}
	d.UserInfo = user
	d.GroupsInfo, err = user.GroupsUserIsJoining()
	if err != nil {
		http.Redirect(w, r, "/error", http.StatusFound)
		return
	}
	d.CurrentGroupInfo, err = data.GroupByGroupID(vars["group_id"])
	if err != nil {
		http.Redirect(w, r, "/error", http.StatusFound)
		return
	}
	d.CurrentGroupMembers, err = d.CurrentGroupInfo.GroupMembers()
	if err != nil {
		http.Redirect(w, r, "/error", http.StatusFound)
		return
	}
	generateHTML(w, d, "layout.html", "top_nav.html", "side_nav.html", "channel.html")
	return
}
