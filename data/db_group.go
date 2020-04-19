package data

import (
	"log"
	"math/rand"
	"time"

	"github.com/oklog/ulid"
	"github.com/rs/xid"
)

type Group struct {
	GroupID        string
	GroupName      string
	NameSnippet    string
	MembersNumber  int
	ChannelsNumber int
	UserID         int
	TypeID         int
	CreatedAt      time.Time
}

func SearchGroupByKeyword(keyword string) (groups []Group, err error) {
	rows, err := Db.Query("SELECT * FROM groups WHERE type_id = (SELECT type_id FROM types WHERE type_name = 'public') AND group_name LIKE $1 OR group_id = $2", "%"+keyword+"%", keyword)
	if err != nil {
		log.Print(err.Error())
		return
	}
	for rows.Next() {
		conv := Group{}
		err = rows.Scan(&conv.GroupID, &conv.GroupName, &conv.NameSnippet, &conv.MembersNumber, &conv.ChannelsNumber, &conv.UserID, &conv.TypeID, &conv.CreatedAt)
		if err != nil {
			log.Print(err.Error())
			return
		}
		groups = append(groups, conv)
	}
	rows.Close()
	return
}

func (group *Group) CreatedTime() (str string) {
	str = PrintTime(group.CreatedAt)
	return
}

func (group *Group) GroupMembers() (members []User, err error) {
	rows, err := Db.Query("SELECT user_id, user_name, user_code FROM users WHERE user_id IN (SELECT user_id FROM group_members WHERE group_id = $1)", group.GroupID)
	if err != nil {
		log.Print(err.Error())
		return
	}
	for rows.Next() {
		conv := User{}
		err = rows.Scan(&conv.UserID, &conv.UserName, &conv.UserCode)
		if err != nil {
			log.Print(err.Error())
			return
		}
		members = append(members, conv)
	}
	rows.Close()
	return
}

func (user *User) IsGroupMember(groupID string) (ok bool, err error) {
	stmt, err := Db.Prepare("SELECT EXISTS (SELECT * FROM group_members WHERE group_id = $1 AND user_id = $2)")
	if err != nil {
		log.Print(err.Error())
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(groupID, user.UserID).Scan(&ok)
	if err != nil {
		log.Print(err.Error())
	}
	return
}

func (user *User) GroupsUserIsJoining() (groupsInfo []Group, err error) {
	rows, err := Db.Query("SELECT group_id, group_name, name_snippet FROM groups WHERE group_id IN (SELECT group_id FROM group_members WHERE user_id = $1)", user.UserID)
	if err != nil {
		log.Print(err.Error())
		return
	}
	for rows.Next() {
		conv := Group{}
		err = rows.Scan(&conv.GroupID, &conv.GroupName, &conv.NameSnippet)
		if err != nil {
			log.Print(err.Error())
			return
		}
		groupsInfo = append(groupsInfo, conv)
	}
	rows.Close()
	return
}

func (group *Group) New() (err error) {
	tx, err := Db.Begin()
	if err != nil {
		log.Print(err.Error())
		return
	}
	stmt, err := tx.Prepare("INSERT INTO groups (group_id, group_name, name_snippet, user_id, type_id) VALUES ($1, $2, $3, $4, $5)")
	if err != nil {
		log.Print(err.Error())
		tx.Rollback()
		return
	}
	defer stmt.Close()
	guid := xid.New()
	group.GroupID = guid.String()
	group.NameSnippet = Snippet(group.GroupName, 16)
	// ----- Execute a SQL statement.
	_, err = stmt.Exec(group.GroupID, group.GroupName, group.NameSnippet, group.UserID, group.TypeID)
	if err != nil {
		log.Print(err.Error())
		tx.Rollback()
		return
	}
	stmt, err = tx.Prepare("INSERT INTO group_members (user_id, group_id, status_id) VALUES ($1, $2, (SELECT status_id FROM status WHERE status_name = 'online'))")
	if err != nil {
		log.Print(err.Error())
		tx.Rollback()
		return
	}
	defer stmt.Close()
	// ----- Execute a SQL statement.
	_, err = stmt.Exec(group.UserID, group.GroupID)
	if err != nil {
		log.Print(err.Error())
		tx.Rollback()
		return
	}
	stmt, err = tx.Prepare("INSERT INTO channels (channel_id, channel_name, channel_topic, name_snippet, topic_snippet, user_id, group_id, type_id) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)")
	if err != nil {
		log.Print(err.Error())
		tx.Rollback()
		return
	}
	defer stmt.Close()
	general := Channel{}
	now := time.Now()
	entropy := ulid.Monotonic(rand.New(rand.NewSource(now.UnixNano())), 0)
	id, err := ulid.New(ulid.Timestamp(now), entropy)
	general.ChannelID = id.String()
	general.ChannelName = "general"
	general.ChannelTopic = "This channel is for group-wide communication. All group members are in this channel."
	general.NameSnippet = Snippet(general.ChannelName, 16)
	general.TopicSnippet = Snippet(general.ChannelTopic, 256)
	general.UserID = group.UserID
	general.GroupID = group.GroupID
	general.TypeID = 1
	general.CreatedAt = group.CreatedAt
	general.UpdatedAt = group.CreatedAt
	general.EditedAt = group.CreatedAt
	// ----- Execute a SQL statement.
	_, err = stmt.Exec(general.ChannelID, general.ChannelName, general.ChannelTopic, general.NameSnippet, general.TopicSnippet, general.UserID, general.GroupID, general.TypeID)
	if err != nil {
		log.Print(err.Error())
		tx.Rollback()
		return
	}
	stmt, err = tx.Prepare("INSERT INTO channel_members (user_id, channel_id, status_id) VALUES ($1, $2, (SELECT status_id FROM status WHERE status_name = 'joined'))")
	if err != nil {
		log.Print(err.Error())
		tx.Rollback()
		return
	}
	defer stmt.Close()
	_, err = stmt.Exec(group.UserID, general.ChannelID)
	if err != nil {
		log.Print(err.Error())
		tx.Rollback()
		return
	}
	stmt, err = tx.Prepare("UPDATE channels SET members_number = 1 WHERE group_id = $1")
	if err != nil {
		log.Print(err.Error())
		tx.Rollback()
		return
	}
	defer stmt.Close()
	_, err = stmt.Exec(group.GroupID)
	if err != nil {
		log.Print(err.Error())
		tx.Rollback()
		return
	}
	err = tx.Commit()
	if err != nil {
		log.Print(err.Error())
	}
	return
}

func GroupByGroupID(groupID string) (group Group, err error) {
	group = Group{}
	stmt, err := Db.Prepare("SELECT * FROM groups WHERE group_id = $1")
	if err != nil {
		log.Print(err.Error())
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(groupID).Scan(&group.GroupID, &group.GroupName, &group.NameSnippet, &group.MembersNumber, &group.ChannelsNumber, &group.UserID, &group.TypeID, &group.CreatedAt)
	if err != nil {
		log.Print(err.Error())
	}
	return
}
