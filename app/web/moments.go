package web

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/dawnim/dawn-server/app/admin"
	"github.com/dawnim/dawn-server/app/admin/controllers"
	"github.com/dawnim/dawn-server/app/admin/response"
	sdk "github.com/dawnim/dawn-server/app/models"
	"github.com/dawnim/dawn-server/support/customer"
	"github.com/dawnim/dawn-server/support/geoip"
	"github.com/dawnim/dawn-simple/app/models"
	"github.com/dawnim/dawn-simple/app/momentcontent"
	"github.com/gin-gonic/gin"
	"github.com/prismgo/framework/route"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type momentView struct {
	UserID      uint            `json:"-"`
	Format      string          `json:"format"`
	Title       string          `json:"title"`
	Board       string          `json:"board"`
	ActivityAt  int64           `json:"activity_at"`
	ID          uint            `json:"id"`
	Content     string          `json:"content"`
	Images      string          `json:"images"`
	Location    string          `json:"location"`
	Region      string          `json:"region"`
	CreatedAt   int64           `json:"created_at"`
	UniqueID    string          `json:"unique_id"`
	Nickname    string          `json:"nickname"`
	Avatar      string          `json:"avatar"`
	UserType    string          `json:"user_type"`
	Flowers     int64           `json:"flowers"`
	Eggs        int64           `json:"eggs"`
	Comments    int64           `json:"comments"`
	IsAnonymous bool            `json:"is_anonymous"`
	CanManage   bool            `json:"can_manage"`
	ImageNames  map[uint]string `json:"image_names,omitempty" gorm:"-"`
}

type commentView struct {
	ID         uint            `json:"id"`
	MomentID   uint            `json:"moment_id"`
	ReplyID    uint            `json:"reply_id"`
	Content    string          `json:"content"`
	Region     string          `json:"region"`
	CreatedAt  int64           `json:"created_at"`
	UniqueID   string          `json:"unique_id"`
	Nickname   string          `json:"nickname"`
	Avatar     string          `json:"avatar"`
	UserType   string          `json:"user_type"`
	ReplyName  string          `json:"reply_name"`
	ImageNames map[uint]string `json:"image_names,omitempty" gorm:"-"`
}

type momentInput struct {
	Format      string `json:"format" binding:"omitempty,oneof=moment post"`
	Title       string `json:"title"`
	Board       string `json:"board" binding:"omitempty,oneof=general help showcase ai"`
	Content     string `json:"content"`
	Location    string `json:"location"`
	Images      []uint `json:"images"`
	IsAnonymous bool   `json:"is_anonymous"`
}

func (in *momentInput) normalize(media momentcontent.Media) bool {
	in.Title, in.Location = strings.TrimSpace(in.Title), strings.TrimSpace(in.Location)
	if in.Format == "" {
		in.Format = "moment"
	}
	if in.Board == "" {
		in.Board = "general"
	}
	if in.Format == "moment" && in.Title == "" {
		first, _, _ := strings.Cut(strings.TrimSpace(in.Content), "\n")
		in.Title = string([]rune(first)[:min(60, utf8.RuneCountInString(first))])
	}
	maxContent := 1024
	if in.Format == "post" {
		maxContent = 10000
	}
	// 动态正文只放文字与 Markdown 格式，图片/视频统一走编辑器下方附件区。
	if in.Format == "moment" && media.Inline {
		return false
	}
	if (in.Format == "post" && in.Title == "") || strings.TrimSpace(in.Content) == "" || len(in.Images) > 9 ||
		utf8.RuneCountInString(in.Title) > 100 || utf8.RuneCountInString(in.Content) > maxContent ||
		utf8.RuneCountInString(in.Location) > 200 {
		return false
	}
	if in.IsAnonymous && in.Format != "post" {
		in.IsAnonymous = false
	}
	return true
}

func imagesOwned(a *admin.App, c *gin.Context, userID uint, ids []uint) bool {
	if len(ids) == 0 {
		return true
	}
	unique := make(map[uint]struct{}, len(ids))
	for _, id := range ids {
		unique[id] = struct{}{}
	}
	var count int64
	err := momentcontent.Images(a.DB.WithContext(c.Request.Context()), ids).
		Where("scope = ?", fmt.Sprintf("user:%d", userID)).Count(&count).Error
	return err == nil && count == int64(len(unique))
}

func imageNames(db *gorm.DB, ids []uint) (map[uint]string, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	var rows []sdk.Attachment
	err := momentcontent.Images(db, ids).Select("id", "name").Find(&rows).Error
	names := make(map[uint]string, len(rows))
	for _, row := range rows {
		names[row.ID] = row.Name
	}
	return names, err
}

// siteSession authenticates the website identity and selects its support application.
func siteSession(a *admin.App, c *gin.Context, memberOnly bool) (customer.Session, bool) {
	s, err := a.Customer.Authenticate(c.Request.Context(), strings.TrimPrefix(c.GetHeader("Authorization"), "Bearer "), 0)
	if err == nil && memberOnly && s.User.Type != sdk.UserMember {
		err = customer.ErrForbidden
	}
	if err != nil {
		response.Report(c, err)
		return s, false
	}
	var app sdk.SupportApp
	if err := a.DB.WithContext(c.Request.Context()).Where("public_key = ? AND status = ?", "dawn", sdk.Enabled).First(&app).Error; err != nil {
		response.Report(c, err)
		return s, false
	}
	s.AppID = app.ID
	return s, true
}

func ipRegion(c *gin.Context) string {
	result, err := geoip.Lookup(c.Request.Context(), c.ClientIP())
	if err == nil {
		for _, value := range []string{result.City, result.Province, result.Country} {
			if value != "" && value != "0" {
				return value
			}
		}
	}
	return "未知"
}

func maskMoments(a *admin.App, c *gin.Context, rows ...*momentView) {
	var userID uint
	if token := strings.TrimPrefix(c.GetHeader("Authorization"), "Bearer "); token != "" {
		if session, err := a.Customer.Authenticate(c.Request.Context(), token, 0); err == nil {
			userID = session.User.ID
		}
	}
	for _, row := range rows {
		row.CanManage = userID != 0 && row.UserID == userID
		if row.IsAnonymous {
			row.UniqueID, row.Nickname, row.Avatar = "", "匿名用户", ""
		}
	}
}

func registerMoments(a *admin.App) {
	view := func(c *gin.Context) *gorm.DB {
		return a.DB.WithContext(c.Request.Context()).Model(&models.Moment{}).
			Select("moments.*, users.unique_id, users.nickname, users.avatar, users.type AS user_type, (SELECT COALESCE(SUM(flowers),0) FROM moment_reactions WHERE moment_id = moments.id) AS flowers, (SELECT COALESCE(SUM(eggs),0) FROM moment_reactions WHERE moment_id = moments.id) AS eggs, (SELECT COUNT(*) FROM moment_comments WHERE moment_id = moments.id AND deleted_at = 0) AS comments, COALESCE((SELECT MAX(created_at) FROM moment_comments WHERE moment_id = moments.id AND deleted_at = 0), moments.created_at) AS activity_at").
			Joins("JOIN users ON users.id = moments.user_id")
	}
	// 附件归属使用用户身份；公开读取仅限已发布动态引用的图片。
	uploader := &controllers.Controller{App: a}
	route.Get("/api/moment-location", func(c *gin.Context) {
		response.Respond(c, gin.H{"region": ipRegion(c)})
	})
	route.Post("/api/moment-image-checks", func(c *gin.Context) {
		s, ok := siteSession(a, c, true)
		if !ok {
			return
		}
		var input controllers.UploadInput
		if c.ShouldBindJSON(&input) != nil {
			response.Report(c, customer.ErrInvalid)
			return
		}
		input.Scope, input.OwnerID = fmt.Sprintf("user:%d", s.User.ID), s.User.ID
		input.CategoryID = 0
		row, exists, err := uploader.CheckUpload(c, input)
		if err != nil {
			response.Report(c, err, 422)
			return
		}
		if exists && !slices.Contains(momentcontent.ImageMIMEs, row.File.MIME) {
			response.Report(c, customer.ErrInvalid)
			return
		}
		response.Respond(c, gin.H{"exists": exists, "id": row.ID, "name": row.Name})
	})
	route.Post("/api/moment-images", func(c *gin.Context) {
		s, ok := siteSession(a, c, true)
		if !ok {
			return
		}
		raw, err := a.Setting(c.Request.Context(), "upload.max_size")
		if err != nil {
			response.Report(c, err)
			return
		}
		var limit int64
		if err = json.Unmarshal(raw, &limit); err != nil {
			response.Report(c, err)
			return
		}
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, limit+(1<<20))
		header, err := c.FormFile("file")
		if err != nil {
			response.Report(c, customer.ErrInvalid)
			return
		}
		defer c.Request.MultipartForm.RemoveAll()
		file, err := header.Open()
		if err != nil {
			response.Report(c, err)
			return
		}
		defer file.Close()
		head := make([]byte, 512)
		n, err := file.Read(head)
		if err != nil && err != io.EOF {
			response.Report(c, err)
			return
		}
		if !slices.Contains(momentcontent.ImageMIMEs, http.DetectContentType(head[:n])) {
			response.Report(c, customer.ErrInvalid)
			return
		}
		if _, err = file.Seek(0, io.SeekStart); err != nil {
			response.Report(c, err)
			return
		}
		row, err := uploader.SaveUpload(c, file, controllers.UploadInput{Name: header.Filename, Size: header.Size, MD5: c.PostForm("md5"), SHA256: c.PostForm("sha256"), Scope: fmt.Sprintf("user:%d", s.User.ID), OwnerID: s.User.ID}, limit)
		if err != nil {
			response.Report(c, err, 422)
			return
		}
		response.Created(c, gin.H{"id": row.ID, "name": row.Name}, c.Request.URL.Path)
	})
	serveImage := func(c *gin.Context, ids []uint) {
		imageID, err := strconv.ParseUint(c.Param("image_id"), 10, strconv.IntSize)
		if err != nil || !slices.Contains(ids, uint(imageID)) {
			c.Status(http.StatusNotFound)
			return
		}
		var image sdk.Attachment
		if err := momentcontent.Images(a.DB.WithContext(c.Request.Context()), []uint{uint(imageID)}).Preload("File").First(&image).Error; err != nil {
			response.Report(c, err)
			return
		}
		uploader.ServeAttachment(c, image)
	}
	route.Get("/api/moments/{id}/images/{image_id}", func(c *gin.Context) {
		id := response.ID(c)
		if c.IsAborted() {
			return
		}
		var row models.Moment
		if err := a.DB.WithContext(c.Request.Context()).First(&row, id).Error; err != nil {
			response.Report(c, err)
			return
		}
		var ids []uint
		if err := json.Unmarshal([]byte(row.Images), &ids); err != nil {
			response.Report(c, err)
			return
		}
		if row.Format == "post" {
			ids = append(ids, momentcontent.ParseMedia(row.Content).IDs...)
		}
		serveImage(c, ids)
	})
	route.Get("/api/moments/{id}/comments/{comment_id}/images/{image_id}", func(c *gin.Context) {
		id := response.ID(c)
		if c.IsAborted() {
			return
		}
		var comment models.MomentComment
		db := a.DB.WithContext(c.Request.Context())
		posts := db.Model(&models.Moment{}).Select("id").Where("id = ?", id)
		if err := db.Where("id = ? AND moment_id IN (?)", c.Param("comment_id"), posts).First(&comment).Error; err != nil {
			response.Report(c, err)
			return
		}
		serveImage(c, momentcontent.ParseMedia(comment.Content).IDs)
	})
	// This is a business query; the shared customer widget has no forum knowledge.
	route.Get("/api/moment-overview", func(c *gin.Context) {
		var posts, replies, members int64
		for _, item := range []struct {
			model any
			count *int64
		}{
			{&models.Moment{}, &posts}, {&models.MomentComment{}, &replies},
		} {
			if err := a.DB.WithContext(c.Request.Context()).Model(item.model).Count(item.count).Error; err != nil {
				response.Report(c, err)
				return
			}
		}
		if err := a.DB.WithContext(c.Request.Context()).Model(&sdk.User{}).Where("type = ?", sdk.UserMember).Count(&members).Error; err != nil {
			response.Report(c, err)
			return
		}
		hot := []momentView{}
		if err := view(c).Where("moments.created_at >= ?", time.Now().AddDate(0, 0, -7).UnixMilli()).Order("comments * 3 + flowers + eggs DESC, moments.id DESC").Limit(5).Scan(&hot).Error; err != nil {
			response.Report(c, err)
			return
		}
		pointers := make([]*momentView, len(hot))
		for i := range hot {
			pointers[i] = &hot[i]
		}
		maskMoments(a, c, pointers...)
		response.Respond(c, gin.H{"posts": posts, "replies": replies, "members": members, "hot": hot})
	})
	route.Get("/api/moments", func(c *gin.Context) {
		db := view(c)
		if board := c.Query("board"); board != "" {
			db = db.Where("moments.board = ?", board)
		}
		if author := c.Query("author"); author != "" {
			db = db.Where("users.unique_id = ?", author)
		}
		if q := strings.TrimSpace(c.Query("query")); q != "" {
			db = db.Where("moments.title LIKE ? OR moments.content LIKE ?", "%"+q+"%", "%"+q+"%")
		}
		sortBy := c.Query("sort_by")
		if sortBy == "activity_at" {
			db = db.Where("EXISTS (SELECT 1 FROM moment_comments mc WHERE mc.moment_id = moments.id AND mc.deleted_at = 0)")
		}
		var total int64
		if err := db.Session(&gorm.Session{}).Count(&total).Error; err != nil {
			response.Report(c, err)
			return
		}
		order := "moments.id DESC"
		switch sortBy {
		case "activity_at":
			order = "activity_at DESC, moments.id DESC"
		case "hot":
			order = "comments * 3 + flowers + eggs DESC, moments.id DESC"
		}
		page, size := response.Page(c)
		rows := []momentView{}
		if err := db.Order(order).Offset((page - 1) * size).Limit(size).Scan(&rows).Error; err != nil {
			response.Report(c, err)
			return
		}
		pointers := make([]*momentView, len(rows))
		for i := range rows {
			pointers[i] = &rows[i]
		}
		maskMoments(a, c, pointers...)
		response.Paginated(c, rows, total, page, size)
	})
	route.Get("/api/moments/{id}", func(c *gin.Context) {
		id := response.ID(c)
		if c.IsAborted() {
			return
		}
		var row momentView
		result := view(c).Where("moments.id = ?", id).Scan(&row)
		if result.Error != nil {
			response.Report(c, result.Error)
			return
		}
		if result.RowsAffected == 0 {
			response.Report(c, gorm.ErrRecordNotFound)
			return
		}
		if row.Format == "post" {
			names, err := imageNames(a.DB.WithContext(c.Request.Context()), momentcontent.ParseMedia(row.Content).IDs)
			if err != nil {
				response.Report(c, err)
				return
			}
			row.ImageNames = names
		}
		maskMoments(a, c, &row)
		response.Respond(c, row)
	})
	route.Get("/api/moment-config", func(c *gin.Context) {
		var hiddenText, emojiText string
		if raw, err := a.Setting(c.Request.Context(), "moments.toolbar_hidden"); err == nil {
			_ = json.Unmarshal(raw, &hiddenText)
		}
		if raw, err := a.Setting(c.Request.Context(), "moments.emoji_shortcodes"); err == nil {
			_ = json.Unmarshal(raw, &emojiText)
		}
		hidden, emojis := []string{}, map[string]string{}
		for _, key := range strings.Split(hiddenText, ",") {
			if key = strings.TrimSpace(key); key != "" {
				hidden = append(hidden, key)
			}
		}
		for _, line := range strings.Split(emojiText, "\n") {
			key, value, ok := strings.Cut(strings.TrimSpace(line), "=")
			if ok && key != "" && value != "" {
				emojis[strings.TrimSpace(key)] = strings.TrimSpace(value)
			}
		}
		response.Respond(c, gin.H{"toolbar_hidden": hidden, "emojis": emojis})
	})
	route.Get("/api/mentions/search", func(c *gin.Context) {
		rows := []gin.H{}
		if q := strings.TrimSpace(c.Query("q")); q != "" {
			_ = a.DB.WithContext(c.Request.Context()).Table("users").Select("unique_id, nickname").
				Where("nickname LIKE ? OR unique_id LIKE ?", "%"+q+"%", "%"+q+"%").Limit(8).Scan(&rows).Error
		}
		response.Respond(c, rows)
	})
	route.Put("/api/moments/{id}", func(c *gin.Context) {
		s, ok := siteSession(a, c, true)
		if !ok {
			return
		}
		id := response.ID(c)
		if c.IsAborted() {
			return
		}
		var input momentInput
		if c.ShouldBindJSON(&input) != nil {
			response.Report(c, customer.ErrInvalid)
			return
		}
		media := momentcontent.ParseMedia(input.Content)
		ids := append(slices.Clone(input.Images), media.IDs...)
		if !input.normalize(media) || !imagesOwned(a, c, s.User.ID, ids) {
			response.Report(c, customer.ErrInvalid)
			return
		}
		images, _ := json.Marshal(input.Images)
		result := a.DB.WithContext(c.Request.Context()).Model(&models.Moment{}).
			Where("id = ? AND user_id = ?", id, s.User.ID).
			Updates(map[string]any{
				"format": input.Format, "title": input.Title, "board": input.Board,
				"content": input.Content, "location": input.Location,
				"images": string(images), "is_anonymous": input.IsAnonymous,
			})
		if result.Error != nil {
			response.Report(c, result.Error)
			return
		}
		if result.RowsAffected == 0 {
			response.Report(c, gorm.ErrRecordNotFound)
			return
		}
		touchFiles(a.DB.WithContext(c.Request.Context()), ids)
		response.Respond(c, gin.H{"id": id})
	})
	route.Post("/api/moments", func(c *gin.Context) {
		s, ok := siteSession(a, c, true)
		if !ok {
			return
		}
		var input momentInput
		if c.ShouldBindJSON(&input) != nil {
			response.Report(c, customer.ErrInvalid)
			return
		}
		media := momentcontent.ParseMedia(input.Content)
		ids := append(slices.Clone(input.Images), media.IDs...)
		if !input.normalize(media) {
			response.Report(c, customer.ErrInvalid)
			return
		}
		raw, err := a.Setting(c.Request.Context(), "moments.daily_limit")
		if err != nil {
			response.Report(c, err)
			return
		}
		var dailyLimit int64
		if err := json.Unmarshal(raw, &dailyLimit); err != nil {
			response.Report(c, err)
			return
		}
		if !imagesOwned(a, c, s.User.ID, ids) {
			response.Report(c, customer.ErrInvalid)
			return
		}
		images, _ := json.Marshal(input.Images)
		if input.Location == "" {
			input.Location = ipRegion(c)
		}
		row := models.Moment{UserID: s.User.ID, Format: input.Format, Title: input.Title, Board: input.Board, Content: input.Content, Location: input.Location, Region: ipRegion(c), Images: string(images), IsAnonymous: input.IsAnonymous}
		err = a.DB.WithContext(c.Request.Context()).Transaction(func(tx *gorm.DB) error {
			var user sdk.User
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&user, s.User.ID).Error; err != nil {
				return err
			}
			now := time.Now()
			start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
			var count int64
			if err := tx.Unscoped().Model(&models.Moment{}).Where("user_id = ? AND created_at >= ? AND created_at < ?", s.User.ID, start.UnixMilli(), start.AddDate(0, 0, 1).UnixMilli()).Count(&count).Error; err != nil {
				return err
			}
			if count >= dailyLimit {
				return customer.ErrLimited
			}
			if err := tx.Create(&row).Error; err != nil {
				return err
			}
			conv, err := momentConversation(tx, s.AppID)
			if err != nil {
				return err
			}
			sender := s.User.UniqueID
			if row.IsAnonymous {
				sender = "anonymous"
			}
			return a.Customer.Message(tx, &conv, &sdk.ChatMessage{SenderRealm: customer.Realm(s.AppID), SenderSubject: sender, Type: "event", Content: row.Content, Payload: customer.Raw(map[string]any{"kind": "moment", "moment_id": row.ID})})
		})
		if err != nil {
			response.Report(c, err)
			return
		}
		a.WakeChat(c.Request.Context())
		touchFiles(a.DB.WithContext(c.Request.Context()), ids)
		response.Created(c, gin.H{"id": row.ID}, fmt.Sprintf("%s/%d", c.Request.URL.Path, row.ID))
	})
	// 软删除：models.Moment 继承 sdk.Base（soft_delete.DeletedAt），Delete 只写 deleted_at，
	// 行数据、评论与点赞都保留；列表、详情与统计由 GORM 自动排除已删除行。
	route.Delete("/api/moments/{id}", func(c *gin.Context) {
		s, ok := siteSession(a, c, true)
		if !ok {
			return
		}
		id := response.ID(c)
		if c.IsAborted() {
			return
		}
		result := a.DB.WithContext(c.Request.Context()).
			Where("id = ? AND user_id = ? AND deleted_at = 0", id, s.User.ID).
			Delete(&models.Moment{})
		if result.Error != nil {
			response.Report(c, result.Error)
			return
		}
		if result.RowsAffected == 0 {
			response.Report(c, gorm.ErrRecordNotFound)
			return
		}
		c.Status(http.StatusNoContent)
	})
	route.Get("/api/moments/{id}/comments", func(c *gin.Context) {
		id := response.ID(c)
		if c.IsAborted() {
			return
		}
		var post models.Moment
		if err := a.DB.WithContext(c.Request.Context()).First(&post, id).Error; err != nil {
			response.Report(c, err)
			return
		}
		var total int64
		db := a.DB.WithContext(c.Request.Context()).Model(&models.MomentComment{}).Where("moment_comments.moment_id = ?", id)
		if err := db.Count(&total).Error; err != nil {
			response.Report(c, err)
			return
		}
		page, size := response.Page(c)
		if commentID := c.Query("comment_id"); commentID != "" {
			var before int64
			if err := a.DB.WithContext(c.Request.Context()).Model(&models.MomentComment{}).Where("moment_id = ? AND id < ?", id, commentID).Count(&before).Error; err != nil {
				response.Report(c, err)
				return
			}
			page = int(before)/size + 1
		}
		rows := []commentView{}
		err := db.Select("moment_comments.*, users.unique_id, users.nickname, users.avatar, users.type AS user_type, COALESCE(NULLIF(reply_users.nickname, ''), reply_users.unique_id, '') AS reply_name").Joins("JOIN users ON users.id = moment_comments.user_id").Joins("LEFT JOIN moment_comments AS replies ON replies.id = moment_comments.reply_id AND replies.moment_id = moment_comments.moment_id").Joins("LEFT JOIN users AS reply_users ON reply_users.id = replies.user_id").Order("moment_comments.id ASC").Offset((page - 1) * size).Limit(size).Scan(&rows).Error
		if err != nil {
			response.Report(c, err)
			return
		}
		refs := make([][]uint, len(rows))
		var ids []uint
		for i := range rows {
			refs[i] = momentcontent.ParseMedia(rows[i].Content).IDs
			ids = append(ids, refs[i]...)
		}
		names, err := imageNames(a.DB.WithContext(c.Request.Context()), ids)
		if err != nil {
			response.Report(c, err)
			return
		}
		for i, keys := range refs {
			if len(keys) == 0 {
				continue
			}
			rows[i].ImageNames = make(map[uint]string, len(keys))
			for _, id := range keys {
				if name, ok := names[id]; ok {
					rows[i].ImageNames[id] = name
				}
			}
		}
		response.Paginated(c, rows, total, page, size)
	})
	route.Post("/api/moments/{id}/comments", func(c *gin.Context) {
		s, ok := siteSession(a, c, true)
		if !ok {
			return
		}
		id := response.ID(c)
		if c.IsAborted() {
			return
		}
		var input struct {
			Content string `json:"content"`
			ReplyID uint   `json:"reply_id"`
		}
		if c.ShouldBindJSON(&input) != nil {
			response.Report(c, customer.ErrInvalid)
			return
		}
		ids := momentcontent.ParseMedia(input.Content).IDs
		if strings.TrimSpace(input.Content) == "" || utf8.RuneCountInString(input.Content) > 1024 || !imagesOwned(a, c, s.User.ID, ids) {
			response.Report(c, customer.ErrInvalid)
			return
		}
		row := models.MomentComment{MomentID: id, UserID: s.User.ID, ReplyID: input.ReplyID, Content: input.Content, Region: ipRegion(c)}
		err := a.DB.WithContext(c.Request.Context()).Transaction(func(tx *gorm.DB) error {
			var post models.Moment
			if err := tx.First(&post, id).Error; err != nil {
				return err
			}
			targetID := uint(0)
			if !post.IsAnonymous {
				targetID = post.UserID
			}
			if input.ReplyID != 0 {
				var reply models.MomentComment
				if err := tx.Where("moment_id = ?", id).First(&reply, input.ReplyID).Error; err != nil {
					return err
				}
				targetID = reply.UserID
			}
			if err := tx.Create(&row).Error; err != nil {
				return err
			}
			if targetID == 0 || targetID == s.User.ID {
				return nil
			}
			var target sdk.User
			if err := tx.First(&target, targetID).Error; err != nil {
				return err
			}
			if target.Type != sdk.UserMember {
				return nil
			}
			conv, err := momentConversation(tx, s.AppID, s.User, target)
			if err != nil {
				return err
			}
			return a.Customer.Message(tx, &conv, &sdk.ChatMessage{SenderRealm: customer.Realm(s.AppID), SenderSubject: s.User.UniqueID, Type: "event", Content: row.Content, Payload: customer.Raw(map[string]any{"kind": "comment", "moment_id": id, "comment_id": row.ID})})
		})
		if err != nil {
			response.Report(c, err)
			return
		}
		a.WakeChat(c.Request.Context())
		touchFiles(a.DB.WithContext(c.Request.Context()), ids)
		response.Created(c, gin.H{"id": row.ID}, fmt.Sprintf("%s/%d", c.Request.URL.Path, row.ID))
	})
	route.Delete("/api/moments/{id}/comments/{comment_id}", func(c *gin.Context) {
		s, ok := siteSession(a, c, true)
		if !ok {
			return
		}
		id := response.ID(c)
		if c.IsAborted() {
			return
		}
		result := a.DB.WithContext(c.Request.Context()).Where("id = ? AND moment_id = ? AND user_id = ?", c.Param("comment_id"), id, s.User.ID).Delete(&models.MomentComment{})
		if result.Error != nil {
			response.Report(c, result.Error)
			return
		}
		if result.RowsAffected == 0 {
			response.Report(c, gorm.ErrRecordNotFound)
			return
		}
		c.Status(http.StatusNoContent)
	})
	route.Post("/api/moments/{id}/reactions", func(c *gin.Context) {
		s, ok := siteSession(a, c, true)
		if !ok {
			return
		}
		id := response.ID(c)
		if c.IsAborted() {
			return
		}
		var input struct {
			Kind string `json:"kind"`
		}
		if c.ShouldBindJSON(&input) != nil || (input.Kind != "flowers" && input.Kind != "eggs") {
			response.Report(c, customer.ErrInvalid)
			return
		}
		err := a.DB.WithContext(c.Request.Context()).Transaction(func(tx *gorm.DB) error {
			var post models.Moment
			if err := tx.First(&post, id).Error; err != nil {
				return err
			}
			row := models.MomentReaction{MomentID: id, UserID: s.User.ID}
			if input.Kind == "flowers" {
				row.Flowers = 1
			} else {
				row.Eggs = 1
			}
			return tx.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "moment_id"}, {Name: "user_id"}}, DoUpdates: clause.Assignments(map[string]any{input.Kind: gorm.Expr(input.Kind+" + ?", 1), "updated_at": time.Now().UnixMilli()})}).Create(&row).Error
		})
		if err != nil {
			response.Report(c, err)
			return
		}
		c.Status(http.StatusNoContent)
	})
}
