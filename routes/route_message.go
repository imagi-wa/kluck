package routes

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/imagi-wa/kluck/data"
)

func ReadDirectMessageHandler(w http.ResponseWriter, r *http.Request) {
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
	d.ConversationUserCode = vars["user_code"]
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
	dmUser, err := data.UserByUserCode(vars["user_code"])
	if err != nil {
		http.Redirect(w, r, "/error", http.StatusFound)
		return
	}
	d.DirectMessages, err = d.CurrentGroupInfo.DirectMessages(user.UserID, dmUser.UserID)
	if err != nil {
		http.Redirect(w, r, "/error", http.StatusFound)
		return
	}
	d.CurrentGroupMembers, err = d.CurrentGroupInfo.GroupMembers()
	if err != nil {
		http.Redirect(w, r, "/error", http.StatusFound)
		return
	}
	generateHTML(w, d, "layout.html", "top_nav.html", "side_nav.html", "direct_message.html")
	return
}

func NewDirectMessageHandler(w http.ResponseWriter, r *http.Request) {
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
	dm := data.DirectMessage{}
	dm.Sentence = r.PostFormValue("sentence")
	dm.SenderID = user.UserID
	receiverUser, err := data.UserByUserCode(vars["user_code"])
	if err != nil {
		http.Redirect(w, r, "/error", http.StatusFound)
		return
	}
	dm.ReceiverID = receiverUser.UserID
	dm.GroupID = vars["group_id"]
	err = dm.New()
	if err != nil {
		http.Redirect(w, r, "/error", http.StatusFound)
		return
	}
	http.Redirect(w, r, "/"+vars["group_id"]+"/"+vars["user_code"], http.StatusFound)
	return
}

func NewMessageHandler(w http.ResponseWriter, r *http.Request) {
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
		http.Redirect(w, r, "/"+vars["channel_id"]+"/join", http.StatusFound)
		return
	}
	message := data.Message{}
	message.Sentence = r.PostFormValue("sentence")
	message.ChannelID = vars["channel_id"]
	message.UserID = user.UserID
	err = message.New()
	if err != nil {
		http.Redirect(w, r, "/error", http.StatusFound)
		return
	}
	http.Redirect(w, r, "/"+vars["group_id"]+"/"+vars["channel_id"], http.StatusFound)
	return
}
