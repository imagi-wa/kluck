package routes

import (
	"database/sql"
	"fmt"
	"html/template"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/imagi-wa/kluck/data"
)

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/kluck-public", http.StatusFound)
	return
}

func GroupIndexHandler(w http.ResponseWriter, r *http.Request) {
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
	d.Channels, err = data.IndexChannels(vars["group_id"])
	if err != nil {
		if err == sql.ErrNoRows {
			d.Alert = "No channels' information..."
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

func ErrorHandler(w http.ResponseWriter, r *http.Request) {
	generateHTML(w, nil, "error_page.html")
	return
}

// -----> helper function

func generateHTML(w http.ResponseWriter, data interface{}, filenames ...string) {
	var files []string
	for _, file := range filenames {
		files = append(files, fmt.Sprintf("./web/templates/%s", file))
	}
	t := template.Must(template.ParseFiles(files...))
	t.ExecuteTemplate(w, "layout", data)
	return
}
