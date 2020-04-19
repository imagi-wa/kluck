package data

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/oklog/ulid"
)

type Channel struct {
	ChannelID      string
	ChannelName    string
	ChannelTopic   string
	NameSnippet    string
	TopicSnippet   string
	MessagesNumber int
	MembersNumber  int
	UserID         int
	GroupID        string
	TypeID         int
	CreatedAt      time.Time
	UpdatedAt      time.Time
	EditedAt       time.Time
}

func (channel *Channel) UserImg() (str string) {
	user := User{}
	user, _ = UserByUserID(channel.UserID)
	str = user.UserImg
	return
}

func SearchChannelByKeyword(keyword string, groupID string) (channels []Channel, err error) {
	rows, err := Db.Query("SELECT channel_id, name_snippet, topic_snippet, messages_number, members_number, user_id, group_id, type_id, created_at, updated_at FROM channels WHERE group_id = $1 AND type_id = (SELECT type_id FROM types WHERE type_name = 'public') AND channel_name LIKE $2 OR channel_id = $3 ORDER BY updated_at DESC", groupID, "%"+keyword+"%", keyword)
	if err != nil {
		log.Print(err.Error())
		return
	}
	for rows.Next() {
		conv := Channel{}
		err = rows.Scan(&conv.ChannelID, &conv.NameSnippet, &conv.TopicSnippet, &conv.MessagesNumber, &conv.MembersNumber, &conv.UserID, &conv.GroupID, &conv.TypeID, &conv.CreatedAt, &conv.UpdatedAt)
		if err != nil {
			log.Print(err.Error())
			return
		}
		channels = append(channels, conv)
	}
	rows.Close()
	return
}

func (channel *Channel) UnreadMessagesNumber(userID int) (number int) {
	stmt, err := Db.Prepare("SELECT COUNT (*) FROM messages WHERE channel_id = $1 AND created_at > (SELECT last_accessed_at FROM channel_members WHERE channel_id = $1 AND user_id = $2)")
	if err != nil {
		number = 0
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(channel.ChannelID, userID).Scan(&number)
	if err != nil {
		number = 0
		log.Print(err.Error())
	}
	return
}

func (user *User) ChannelsUserIsJoining(groupID string) (channelsInfo []Channel, err error) {
	rows, err := Db.Query("SELECT channel_id, channel_name, name_snippet, group_id, type_id FROM channels WHERE group_id = $1 AND channel_id IN (SELECT channel_id FROM channel_members WHERE user_id = $2)", groupID, user.UserID)
	if err != nil {
		log.Print(err.Error())
		return
	}
	for rows.Next() {
		conv := Channel{}
		err = rows.Scan(&conv.ChannelID, &conv.ChannelName, &conv.NameSnippet, &conv.GroupID, &conv.TypeID)
		if err != nil {
			log.Print(err.Error())
			return
		}
		channelsInfo = append(channelsInfo, conv)
	}
	rows.Close()
	return
}

func (user *User) IsChannelMember(channel_id string) (ok bool, err error) {
	stmt, err := Db.Prepare("SELECT EXISTS (SELECT * FROM channel_members WHERE channel_id = $1 AND user_id = $2)")
	if err != nil {
		log.Print(err.Error())
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(channel_id, user.UserID).Scan(&ok)
	if err != nil {
		log.Print(err.Error())
	}
	return
}

func (user *User) JoinChannel(channelID string) (err error) {
	tx, err := Db.Begin()
	if err != nil {
		log.Print(err.Error())
		return
	}
	stmt, err := tx.Prepare("INSERT INTO channel_members (user_id, channel_id, status_id) VALUES ($1, $2, (SELECT status_id FROM status WHERE status_name = 'joined'))")
	if err != nil {
		log.Print(err.Error())
		tx.Rollback()
		return
	}
	defer stmt.Close()
	_, err = stmt.Exec(user.UserID, channelID)
	if err != nil {
		log.Print(err.Error())
		tx.Rollback()
		return
	}
	stmt, err = tx.Prepare("UPDATE channels SET members_number = (SELECT COUNT (*) FROM channel_members WHERE channel_id = $1) WHERE channel_id = $1")
	if err != nil {
		log.Print(err.Error())
		tx.Rollback()
		return
	}
	defer stmt.Close()
	_, err = stmt.Exec(channelID)
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

func (channel *Channel) IsEdited() (done bool) {
	done = !channel.EditedAt.Equal(channel.CreatedAt)
	return
}

func (channel *Channel) CreatedTime() (str string) {
	str = PrintTime(channel.CreatedAt)
	return
}

func (channel *Channel) UpdatedTime() (str string) {
	str = PrintTime(channel.UpdatedAt)
	return
}

func (channel *Channel) EditedTime() (str string) {
	str = PrintTime(channel.EditedAt)
	return
}

func (channel *Channel) Members() (members []User, err error) {
	rows, err := Db.Query("SELECT * FROM channel_members WHERE channel_id = $1 ORDER BY joined_at ASC")
	if err != nil {
		log.Print(err.Error())
		return
	}
	for rows.Next() {
		conv := User{}
		err = rows.Scan(&conv.UserID, &conv.UserName, &conv.UserCode, &conv.EmailAddress, &conv.PasswordHash, &conv.CreatedAt)
		if err != nil {
			log.Print(err.Error())
			return
		}
		members = append(members, conv)
	}
	rows.Close()
	return
}

func (channel *Channel) UserName() (str string) {
	user := User{}
	user, _ = UserByUserID(channel.UserID)
	str = user.UserName
	return
}

func (channel *Channel) UserCode() (str string) {
	user := User{}
	user, _ = UserByUserID(channel.UserID)
	str = user.UserCode
	return
}

func (channel *Channel) New() (err error) {
	tx, err := Db.Begin()
	if err != nil {
		log.Print(err.Error())
		return
	}
	stmt, err := tx.Prepare("INSERT INTO channels (channel_id, channel_name, channel_topic, name_snippet, topic_snippet, user_id, group_id, type_id) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)")
	if err != nil {
		log.Print(err.Error())
		tx.Rollback()
		return
	}
	defer stmt.Close()
	channel.NameSnippet = Snippet(channel.ChannelName, 64)
	channel.TopicSnippet = Snippet(channel.ChannelTopic, 256)
	now := time.Now()
	// Create a ULID for th channel's ID.
	entropy := ulid.Monotonic(rand.New(rand.NewSource(now.UnixNano())), 0)
	id, err := ulid.New(ulid.Timestamp(now), entropy)
	channel.ChannelID = id.String()
	// ----- Execute a SQL statement.
	_, err = stmt.Exec(channel.ChannelID, channel.ChannelName, channel.ChannelTopic, channel.NameSnippet, channel.TopicSnippet, channel.UserID, channel.GroupID, channel.TypeID)
	if err != nil {
		log.Print(err.Error())
		tx.Rollback()
		return
	}
	// Register the channel user as its member.
	stmt, err = tx.Prepare("INSERT INTO channel_members (user_id, channel_id, status_id, joined_at) VALUES ($1, $2, (SELECT status_id FROM status WHERE status_name = 'joined'), $3)")
	if err != nil {
		log.Print(err.Error())
		tx.Rollback()
		return
	}
	defer stmt.Close()
	// ----- Execute a SQL statement.
	_, err = stmt.Exec(channel.UserID, channel.ChannelID, now)
	if err != nil {
		log.Print(err.Error())
		tx.Rollback()
		return
	}
	stmt, err = tx.Prepare("UPDATE channels SET members_number = (SELECT COUNT (*) FROM channel_members WHERE channel_id = $1) WHERE channel_id = $1")
	if err != nil {
		log.Print(err.Error())
		tx.Rollback()
		return
	}
	defer stmt.Close()
	_, err = stmt.Exec(channel.ChannelID)
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

func IndexChannels(groupID string) (channels []Channel, err error) {
	rows, err := Db.Query("SELECT channel_id, name_snippet, topic_snippet, messages_number, members_number, user_id, group_id, type_id, created_at, updated_at FROM channels WHERE group_id = $1 AND type_id = (SELECT type_id FROM types WHERE type_name = 'public') ORDER BY updated_at DESC", groupID)
	if err != nil {
		log.Print(err.Error())
		return
	}
	for rows.Next() {
		conv := Channel{}
		err = rows.Scan(&conv.ChannelID, &conv.NameSnippet, &conv.TopicSnippet, &conv.MessagesNumber, &conv.MembersNumber, &conv.UserID, &conv.GroupID, &conv.TypeID, &conv.CreatedAt, &conv.UpdatedAt)
		if err != nil {
			log.Print(err.Error())
			return
		}
		channels = append(channels, conv)
	}
	rows.Close()
	return
}

func (user *User) ChannelByChannelID(channelID string) (channel Channel, err error) {
	tx, err := Db.Begin()
	if err != nil {
		log.Print(err.Error())
		return
	}
	stmt, err := Db.Prepare("SELECT * FROM channels WHERE channel_id = $1")
	if err != nil {
		log.Print(err.Error())
		tx.Rollback()
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(channelID).Scan(&channel.ChannelID, &channel.ChannelName, &channel.ChannelTopic, &channel.NameSnippet, &channel.TopicSnippet, &channel.MessagesNumber, &channel.MembersNumber, &channel.UserID, &channel.GroupID, &channel.TypeID, &channel.CreatedAt, &channel.UpdatedAt, &channel.EditedAt)
	if err != nil {
		log.Print(err.Error())
		tx.Rollback()
		return
	}
	stmt, err = Db.Prepare("UPDATE channel_members SET last_accessed_at = $1 WHERE user_id = $2 AND channel_id = $3")
	if err != nil {
		log.Print(err.Error())
		tx.Rollback()
		return
	}
	defer stmt.Close()
	_, err = stmt.Exec(time.Now(), user.UserID, channelID)
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

// ----- Helper functions ----- //
func Snippet(str string, length int) (snippet string) {
	r := []rune(str)
	if len(r) < length {
		snippet = string(r)
	} else {
		snippet = string(r[:length-3])
		snippet += "..."
	}
	return
}

func PrintTime(t time.Time) (str string) {
	str = fmt.Sprintf("%d/%d/%d %d:%d:%d", t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second())
	return
}

func PrintUser(user User) (str string) {
	str = fmt.Sprintf("%s\t(%s)", user.UserName, user.UserCode)
	return
}
