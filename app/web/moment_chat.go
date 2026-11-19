package web

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/dawnim/dawn-server/app/admin"
	"github.com/dawnim/dawn-server/app/admin/response"
	sdk "github.com/dawnim/dawn-server/app/models"
	"github.com/dawnim/dawn-server/support/customer"
	"github.com/gin-gonic/gin"
	"github.com/prismgo/framework/route"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// momentConversation reuses the framework conversation and transactional event outbox.
func momentConversation(tx *gorm.DB, appID uint, users ...sdk.User) (sdk.ChatConversation, error) {
	identify, kind := fmt.Sprintf("moments:hall:%d", appID), "room"
	if len(users) == 2 {
		identify, kind = fmt.Sprintf("moments:direct:%d:%d:%d", appID, min(users[0].ID, users[1].ID), max(users[0].ID, users[1].ID)), "direct"
	}
	conv := sdk.ChatConversation{Type: kind, AppID: appID, Identify: identify}
	if err := tx.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "identify"}}, DoNothing: true}).Create(&conv).Error; err != nil {
		return conv, err
	}
	if err := tx.Where("identify = ?", identify).First(&conv).Error; err != nil {
		return conv, err
	}
	for _, user := range users {
		member := sdk.ChatMember{ConversationID: conv.ID, Realm: customer.Realm(appID), Subject: user.UniqueID}
		if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&member).Error; err != nil {
			return conv, err
		}
	}
	return conv, nil
}

func registerMomentChat(a *admin.App) {
	member := func(c *gin.Context, s customer.Session) (sdk.ChatConversation, error) {
		var conv sdk.ChatConversation
		err := a.DB.WithContext(c.Request.Context()).Joins("JOIN chat_members ON chat_members.conversation_id = chat_conversations.id").Where("chat_conversations.id = ? AND chat_conversations.app_id = ? AND chat_conversations.identify LIKE ? AND chat_members.realm = ? AND chat_members.subject = ?", c.Param("id"), s.AppID, "moments:%", customer.Realm(s.AppID), s.User.UniqueID).First(&conv).Error
		return conv, err
	}
	route.Get("/api/community-chats", func(c *gin.Context) {
		s, ok := siteSession(a, c, false)
		if !ok {
			return
		}
		err := a.DB.WithContext(c.Request.Context()).Transaction(func(tx *gorm.DB) error { _, err := momentConversation(tx, s.AppID, s.User); return err })
		if err != nil {
			response.Report(c, err)
			return
		}
		rows := []struct {
			sdk.ChatConversation
			Peer     string `json:"peer"`
			Nickname string `json:"nickname"`
			Content  string `json:"content"`
			Unread   int64  `json:"unread"`
		}{}
		err = a.DB.WithContext(c.Request.Context()).Table("chat_conversations AS conv").Select("conv.*, peer.subject AS peer, last.content, (SELECT COUNT(*) FROM chat_messages AS unread WHERE unread.conversation_id = conv.id AND unread.id > mine.last_read_message_id AND unread.sender_subject <> ?) AS unread", s.User.UniqueID).Joins("JOIN chat_members AS mine ON mine.conversation_id = conv.id AND mine.realm = ? AND mine.subject = ?", customer.Realm(s.AppID), s.User.UniqueID).Joins("LEFT JOIN chat_members AS peer ON peer.conversation_id = conv.id AND peer.realm = mine.realm AND peer.subject <> mine.subject AND conv.type = 'direct'").Joins("LEFT JOIN chat_messages AS last ON last.id = conv.last_message_id").Where("conv.app_id = ? AND conv.identify LIKE ?", s.AppID, "moments:%").Order("CASE WHEN conv.type = 'room' THEN 0 ELSE 1 END, conv.last_message_id DESC, conv.id DESC").Scan(&rows).Error
		if err != nil {
			response.Report(c, err)
			return
		}
		subjects := make([]string, 0, len(rows))
		for _, row := range rows {
			subjects = append(subjects, row.Peer)
		}
		users, err := momentChatUsers(a.DB.WithContext(c.Request.Context()), subjects)
		if err != nil {
			response.Report(c, err)
			return
		}
		for i := range rows {
			rows[i].Nickname = users[rows[i].Peer].Nickname
		}
		response.Respond(c, rows)
	})
	route.Post("/api/community-chats", func(c *gin.Context) {
		s, ok := siteSession(a, c, true)
		if !ok {
			return
		}
		var input struct {
			UniqueID string `json:"unique_id"`
		}
		if c.ShouldBindJSON(&input) != nil || input.UniqueID == s.User.UniqueID {
			response.Report(c, customer.ErrInvalid)
			return
		}
		var peer sdk.User
		if err := a.DB.WithContext(c.Request.Context()).Where("unique_id = ? AND type = ?", input.UniqueID, sdk.UserMember).First(&peer).Error; err != nil {
			response.Report(c, err)
			return
		}
		var conv sdk.ChatConversation
		err := a.DB.WithContext(c.Request.Context()).Transaction(func(tx *gorm.DB) error {
			var err error
			conv, err = momentConversation(tx, s.AppID, s.User, peer)
			return err
		})
		if err != nil {
			response.Report(c, err)
			return
		}
		response.Created(c, conv, fmt.Sprintf("%s/%d/messages", c.Request.URL.Path, conv.ID))
	})
	route.Get("/api/community-chats/{id}/messages", func(c *gin.Context) {
		s, ok := siteSession(a, c, false)
		if !ok {
			return
		}
		conv, err := member(c, s)
		if err != nil {
			response.Report(c, err)
			return
		}
		_, size := response.Page(c)
		rows := []struct {
			sdk.ChatMessage
			Nickname string `json:"nickname"`
			Avatar   string `json:"avatar"`
		}{}
		query := a.DB.WithContext(c.Request.Context()).Model(&sdk.ChatMessage{}).Where("conversation_id = ?", conv.ID)
		if before := c.Query("before_id"); before != "" {
			query = query.Where("chat_messages.id < ?", before)
		}
		if err = query.Order("chat_messages.id DESC").Limit(size).Scan(&rows).Error; err != nil {
			response.Report(c, err)
			return
		}
		subjects := make([]string, 0, len(rows))
		for _, row := range rows {
			subjects = append(subjects, row.SenderSubject)
		}
		users, err := momentChatUsers(a.DB.WithContext(c.Request.Context()), subjects)
		if err != nil {
			response.Report(c, err)
			return
		}
		for i := range rows {
			user := users[rows[i].SenderSubject]
			rows[i].Nickname, rows[i].Avatar = user.Nickname, user.Avatar
		}
		response.Respond(c, rows)
	})
	route.Put("/api/community-chats/{id}/read", func(c *gin.Context) {
		s, ok := siteSession(a, c, false)
		if !ok {
			return
		}
		conv, err := member(c, s)
		if err != nil {
			response.Report(c, err)
			return
		}
		var input struct {
			MessageID uint `json:"message_id"`
		}
		if c.ShouldBindJSON(&input) != nil {
			response.Report(c, customer.ErrInvalid)
			return
		}
		// Clamp to this conversation and advance only, including concurrent tabs.
		err = a.DB.WithContext(c.Request.Context()).Model(&sdk.ChatMember{}).Where("conversation_id = ? AND realm = ? AND subject = ? AND last_read_message_id < ?", conv.ID, customer.Realm(s.AppID), s.User.UniqueID, min(input.MessageID, conv.LastMessageID)).Update("last_read_message_id", min(input.MessageID, conv.LastMessageID)).Error
		if err != nil {
			response.Report(c, err)
			return
		}
		c.Status(204)
	})
	route.Post("/api/community-chats/{id}/messages", func(c *gin.Context) {
		s, ok := siteSession(a, c, true)
		if !ok {
			return
		}
		conv, err := member(c, s)
		if err != nil {
			response.Report(c, err)
			return
		}
		var input struct {
			Content   string `json:"content"`
			RequestID string `json:"request_id"`
		}
		if c.ShouldBindJSON(&input) != nil || strings.TrimSpace(input.Content) == "" || utf8.RuneCountInString(input.Content) > 4000 || len(input.RequestID) < 16 || len(input.RequestID) > 64 {
			response.Report(c, customer.ErrInvalid)
			return
		}
		message := sdk.ChatMessage{SenderRealm: customer.Realm(s.AppID), SenderSubject: s.User.UniqueID, RequestID: input.RequestID, Content: strings.TrimSpace(input.Content)}
		err = a.DB.WithContext(c.Request.Context()).Transaction(func(tx *gorm.DB) error {
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&conv, conv.ID).Error; err != nil {
				return err
			}
			var previous sdk.ChatMessage
			result := tx.Where("sender_realm = ? AND sender_subject = ? AND request_id = ?", message.SenderRealm, message.SenderSubject, message.RequestID).Limit(1).Find(&previous)
			if result.Error != nil {
				return result.Error
			}
			if result.RowsAffected > 0 {
				if previous.ConversationID != conv.ID || previous.Content != message.Content {
					return customer.ErrConflict
				}
				message = previous
				return nil
			}
			return a.Customer.Message(tx, &conv, &message)
		})
		if err != nil {
			response.Report(c, err)
			return
		}
		a.WakeChat(c.Request.Context())
		response.Created(c, message, fmt.Sprintf("%s/%d", c.Request.URL.Path, message.ID))
	})
}

// Read profiles in one batch without comparing columns with different SQL collations.
func momentChatUsers(db *gorm.DB, subjects []string) (map[string]sdk.User, error) {
	var users []sdk.User
	if err := db.Select("unique_id", "nickname", "avatar").Where("unique_id IN ?", subjects).Find(&users).Error; err != nil {
		return nil, err
	}
	profiles := make(map[string]sdk.User, len(users))
	for _, user := range users {
		profiles[user.UniqueID] = user
	}
	return profiles, nil
}
