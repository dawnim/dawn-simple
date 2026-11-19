package controllers

import (
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"
	"unicode/utf8"

	sdk "github.com/dawnim/dawn-server/app/admin"
	crud "github.com/dawnim/dawn-server/app/admin/controllers"
	"github.com/dawnim/dawn-server/app/admin/response"
	sysmodels "github.com/dawnim/dawn-server/app/models"
	"github.com/dawnim/dawn-simple/app/models"
	"github.com/gin-gonic/gin"
	"github.com/prismgo/framework/route"
	"gorm.io/gorm"
)

var momentFields = []string{"user_id", "format", "title", "board", "content", "images", "location", "region", "is_anonymous"}
var commentFields = []string{"moment_id", "user_id", "reply_id", "content", "region"}

// RegisterMoments 提供动态与评论的后台增删改查；动态直接复用框架资源，评论列表需按帖子精确过滤。
func RegisterMoments(a *sdk.App, g *route.Registrar) {
	crud.RegisterResource(a, g, crud.Resource[models.Moment]{
		Path: "/moments", Permission: "app.moments",
		Fields:   momentFields,
		Search:   []string{"title", "content", "region", "format", "board"},
		Required: []string{"user_id", "format", "title", "board", "content", "images"},
		Validate: func(m *models.Moment, _ *sdk.Actor, _ bool) error {
			if m.Format == "" {
				m.Format = "moment"
			}
			if m.Board == "" {
				m.Board = "general"
			}
			m.Title, m.Content, m.Location = strings.TrimSpace(m.Title), strings.TrimSpace(m.Content), strings.TrimSpace(m.Location)
			if m.Format == "moment" && m.Title == "" {
				first, _, _ := strings.Cut(m.Content, "\n")
				m.Title = string([]rune(first)[:min(60, utf8.RuneCountInString(first))])
			}
			limit := 300
			if m.Format == "post" {
				limit = 10000
			}
			if !slices.Contains([]string{"moment", "post"}, m.Format) ||
				!slices.Contains([]string{"general", "help", "showcase", "ai"}, m.Board) {
				return errors.New("内容形式或板块不合法")
			}
			if m.Content == "" || (m.Format == "post" && m.Title == "") ||
				utf8.RuneCountInString(m.Title) > 100 || utf8.RuneCountInString(m.Content) > limit ||
				utf8.RuneCountInString(m.Location) > 200 {
				return errors.New("标题或内容不合法")
			}
			if m.Format != "post" {
				m.IsAnonymous = false
			}
			if m.Images == "" {
				m.Images = "[]"
			}
			var images []uint
			if json.Unmarshal([]byte(m.Images), &images) != nil || len(images) > 9 {
				return errors.New("图片列表不合法")
			}
			if !userExists(a.DB, m.UserID) {
				return errors.New("发布用户不存在")
			}
			return nil
		},
	})

	query := func(c *gin.Context) *gorm.DB {
		return a.DB.WithContext(c.Request.Context()).Model(&models.MomentComment{})
	}
	validate := func(row *models.MomentComment) error {
		row.Content = strings.TrimSpace(row.Content)
		var count int64
		if row.Content == "" || utf8.RuneCountInString(row.Content) > 300 || utf8.RuneCountInString(row.Region) > 100 {
			return errors.New("评论内容不合法")
		}
		if row.MomentID == 0 || a.DB.Model(&models.Moment{}).Where("id = ?", row.MomentID).Count(&count).Error != nil || count == 0 {
			return errors.New("动态不存在")
		}
		if !userExists(a.DB, row.UserID) {
			return errors.New("评论用户不存在")
		}
		if row.ReplyID != 0 && (a.DB.Model(&models.MomentComment{}).Where("id = ? AND moment_id = ?", row.ReplyID, row.MomentID).Count(&count).Error != nil || count == 0) {
			return errors.New("回复的评论不存在")
		}
		return nil
	}
	h := crud.ResourceHandlers{Name: "app.moment_comments"}
	h.Index = func(c *gin.Context) {
		db := query(c)
		for _, field := range []string{"moment_id", "user_id"} {
			if value := c.Query(field); value != "" {
				db = db.Where(field+" = ?", value)
			}
		}
		for _, field := range []string{"content", "region"} {
			if value := c.Query(field); value != "" {
				db = db.Where(field+" LIKE ?", "%"+value+"%")
			}
		}
		var total int64
		if err := db.Session(&gorm.Session{}).Count(&total).Error; err != nil {
			response.Report(c, err)
			return
		}
		page, size := response.Page(c)
		rows := []models.MomentComment{}
		if err := db.Order(response.Order(c, append(slices.Clone(commentFields), "id", "created_at", "updated_at")...)).Order("id DESC").Offset((page - 1) * size).Limit(size).Find(&rows).Error; err != nil {
			response.Report(c, err)
			return
		}
		response.Paginated(c, rows, total, page, size)
	}
	h.Show = func(c *gin.Context) {
		id := response.ID(c)
		if c.IsAborted() {
			return
		}
		var row models.MomentComment
		if err := query(c).First(&row, id).Error; err != nil {
			response.Report(c, err)
			return
		}
		response.Respond(c, row)
	}
	h.Store = func(c *gin.Context) {
		var row models.MomentComment
		submitted := crud.BindResourceInput(a.DB, c, &row, commentFields)
		if submitted == nil {
			return
		}
		if err := validate(&row); err != nil {
			response.Report(c, err, 422)
			return
		}
		if err := crud.CreateResource(a.DB.WithContext(c.Request.Context()), &row, commentFields, submitted); err != nil {
			response.Report(c, err)
			return
		}
		response.Created(c, row, fmt.Sprintf("%s/%d", c.Request.URL.Path, row.ID))
	}
	h.Update = func(c *gin.Context) {
		id := response.ID(c)
		if c.IsAborted() {
			return
		}
		var row models.MomentComment
		if err := query(c).First(&row, id).Error; err != nil {
			response.Report(c, err)
			return
		}
		selected := crud.BindResourceInput(a.DB, c, &row, commentFields)
		if selected == nil {
			return
		}
		if c.Request.Method == "PUT" {
			selected = commentFields
		}
		if err := validate(&row); err != nil {
			response.Report(c, err, 422)
			return
		}
		if err := query(c).Where("id = ?", id).Select(selected).Updates(&row).Error; err != nil {
			response.Report(c, err)
			return
		}
		response.Respond(c, row)
	}
	h.Destroy = func(c *gin.Context) {
		ids := response.IDs(c)
		if c.IsAborted() {
			return
		}
		if err := a.DB.WithContext(c.Request.Context()).Transaction(func(tx *gorm.DB) error {
			db := tx.Model(&models.MomentComment{}).Where("id IN ?", ids)
			var total int64
			if err := db.Session(&gorm.Session{}).Count(&total).Error; err != nil {
				return err
			}
			if total != int64(len(ids)) {
				return gorm.ErrRecordNotFound
			}
			return db.Delete(&models.MomentComment{}).Error
		}); err != nil {
			response.Report(c, err)
			return
		}
		c.Status(204)
	}
	crud.RegisterResourceRoutes(a, g, "/moment-comments", "app.moment_comments", h)
}

func userExists(db *gorm.DB, id uint) bool {
	var count int64
	return id != 0 && db.Model(&sysmodels.User{}).Where("id = ?", id).Count(&count).Error == nil && count > 0
}
