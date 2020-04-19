package data

import (
	"log"
	"time"
)

type Message struct {
	MessageID int
	Sentence  string
	UserID    int
	ChannelID string
	CreatedAt time.Time
	EditedAt  time.Time
}

type DirectMessage struct {
	DirectMessageID int
	Sentence        string
	SenderID        int
	ReceiverID      int
	GroupID         string
	CreatedAt       time.Time
	EditedAt        time.Time
}

func (directMessage *DirectMessage) UserImg() (str string) {
	user := User{}
	user, _ = UserByUserID(directMessage.SenderID)
	str = user.UserImg
	return
}

func (directMessage *DirectMessage) CreatedTime() (str string) {
	str = PrintTime(directMessage.CreatedAt)
	return
}

func (directMessage *DirectMessage) EditedTime() (str string) {
	str = PrintTime(directMessage.EditedAt)
	return
}
func (directMessage *DirectMessage) User() (str string) {
	user, _ := UserByUserID(directMessage.SenderID)
	str = PrintUser(user)
	return
}

func (user *User) DirectMessageUsers(groupID string) (directMessagesInfo []User, err error) {
	rows, err := Db.Query("SELECT user_id, user_name, user_code FROM users WHERE user_id IN (SELECT receiver_id FROM direct_messages WHERE group_id = $1 AND sender_id = $2 UNION SELECT sender_id FROM direct_messages WHERE group_id = $1 AND receiver_id = $2)", groupID, user.UserID)
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
		directMessagesInfo = append(directMessagesInfo, conv)
	}
	rows.Close()
	return
}

func (group *Group) DirectMessages(user1ID int, user2ID int) (directMessages []DirectMessage, err error) {
	rows, err := Db.Query("SELECT * FROM direct_messages WHERE group_id = $1 AND sender_id = $2 AND receiver_id = $3 UNION ALL SELECT * FROM direct_messages WHERE group_id = $1 AND sender_id = $3 AND receiver_id = $2 ORDER BY created_at ASC", group.GroupID, user1ID, user2ID)
	if err != nil {
		log.Print(err.Error())
		return
	}
	for rows.Next() {
		conv := DirectMessage{}
		err = rows.Scan(&conv.DirectMessageID, &conv.Sentence, &conv.SenderID, &conv.ReceiverID, &conv.GroupID, &conv.CreatedAt, &conv.EditedAt)
		if err != nil {
			log.Print(err.Error())
			return
		}
		directMessages = append(directMessages, conv)
	}
	rows.Close()
	return
}

func (directMessage *DirectMessage) New() (err error) {
	stmt, err := Db.Prepare("INSERT INTO direct_messages (sentence, sender_id, receiver_id, group_id) VALUES ($1, $2, $3, $4)")
	if err != nil {
		log.Print(err.Error())
		return
	}
	defer stmt.Close()
	_, err = stmt.Exec(directMessage.Sentence, directMessage.SenderID, directMessage.ReceiverID, directMessage.GroupID)
	if err != nil {
		log.Print(err.Error())
	}
	return
}

func (message *Message) UserImg() (str string) {
	user := User{}
	user, _ = UserByUserID(message.UserID)
	str = user.UserImg
	return
}

func (message *Message) IsEdited() (done bool) {
	done = !message.EditedAt.Equal(message.CreatedAt)
	return
}

func (message *Message) CreatedTime() (str string) {
	str = PrintTime(message.CreatedAt)
	return
}

func (message *Message) EditedTime() (str string) {
	str = PrintTime(message.EditedAt)
	return
}

func (message *Message) UserName() (str string) {
	user := User{}
	user, _ = UserByUserID(message.UserID)
	str = user.UserName
	return
}

func (message *Message) UserCode() (str string) {
	user := User{}
	user, _ = UserByUserID(message.UserID)
	str = user.UserCode
	return
}

func (message *Message) New() (err error) {
	tx, err := Db.Begin()
	if err != nil {
		log.Print(err.Error())
		return
	}
	stmt, err := Db.Prepare("INSERT INTO messages (sentence, user_id, channel_id) VALUES ($1, $2, $3)")
	if err != nil {
		log.Print(err.Error())
		tx.Rollback()
		return
	}
	defer stmt.Close()
	_, err = stmt.Exec(message.Sentence, message.UserID, message.ChannelID)
	if err != nil {
		log.Print(err.Error())
		tx.Rollback()
		return
	}
	stmt, err = tx.Prepare("UPDATE channels SET messages_number = (SELECT COUNT (*) FROM messages WHERE channel_id = $1) WHERE channel_id = $1")
	if err != nil {
		log.Print(err.Error())
		tx.Rollback()
		return
	}
	defer stmt.Close()
	_, err = stmt.Exec(message.ChannelID)
	if err != nil {
		log.Print(err.Error())
		tx.Rollback()
		return
	}
	stmt, err = tx.Prepare("UPDATE channels SET updated_at = $1 WHERE channel_id = $2")
	if err != nil {
		log.Print(err.Error())
		tx.Rollback()
		return
	}
	defer stmt.Close()
	_, err = stmt.Exec(time.Now(), message.ChannelID)
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

func (channel *Channel) Messages() (messages []Message, err error) {
	rows, err := Db.Query("SELECT * FROM messages WHERE channel_id = $1 ORDER BY created_at DESC", channel.ChannelID)
	if err != nil {
		log.Print(err.Error())
		return
	}
	for rows.Next() {
		conv := Message{}
		err = rows.Scan(&conv.MessageID, &conv.Sentence, &conv.UserID, &conv.ChannelID, &conv.CreatedAt, &conv.EditedAt)
		if err != nil {
			log.Print(err.Error())
			return
		}
		messages = append(messages, conv)
	}
	rows.Close()
	return
}
