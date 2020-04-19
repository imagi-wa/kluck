package routes

import (
	"database/sql"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/imagi-wa/kluck/data"
)

func SearchGroupHandler(w http.ResponseWriter, r *http.Request) {
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
	d.Groups, err = data.SearchGroupByKeyword(keyword)
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
	generateHTML(w, d, "layout.html", "top_nav.html", "side_nav.html", "groups.html")
	return
}

func NewGroupHandler(w http.ResponseWriter, r *http.Request) {
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
	group := data.Group{}
	group.GroupName = r.PostFormValue("name")
	group.UserID = session.UserID
	group.TypeID, err = strconv.Atoi(r.PostFormValue("type"))
	if err != nil {
		log.Print(err.Error())
		http.Redirect(w, r, "/error", http.StatusFound)
		return
	}
	err = group.New()
	if err != nil {
		http.Redirect(w, r, "/error", http.StatusFound)
		return
	}
	http.Redirect(w, r, "/"+group.GroupID, http.StatusFound)
	return
}
