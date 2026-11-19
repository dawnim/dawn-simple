package web

import (
	"fmt"
	"net/http"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/dawnim/dawn-server/app/admin"
	"github.com/dawnim/dawn-server/app/admin/controllers"
	"github.com/dawnim/dawn-server/app/admin/response"
	sdk "github.com/dawnim/dawn-server/app/models"
	"github.com/gin-gonic/gin"
	"github.com/prismgo/framework/route"
	"gorm.io/gorm"
)

var imageMIMEs = []string{"image/jpeg", "image/png", "image/gif", "image/webp"}

var attachmentRef = regexp.MustCompile(`attachment://(\d+)`)
var mediaInContent = regexp.MustCompile(`(?i)!\[[^\]]*\]\(|<img\b|<video\b|<iframe\b|MEDIA::`)

// 动态正文不允许图片/视频，媒体只走编辑器下方附件区；帖子与评论不限制。
func hasInlineMedia(content string) bool { return mediaInContent.MatchString(content) }

func contentAttachmentIDs(content string) []uint {
	ids := []uint{}
	for _, match := range attachmentRef.FindAllStringSubmatch(content, -1) {
		if id, err := strconv.ParseUint(match[1], 10, 64); err == nil {
			ids = append(ids, uint(id))
		}
	}
	return ids
}

// 附件 id → admin_files.id；秒传复用同一个 file，引用次数天然按文件聚合。
func referencedFileIDs(db *gorm.DB, attachmentIDs []uint) []uint {
	if len(attachmentIDs) == 0 {
		return nil
	}
	var rows []sdk.Attachment
	if err := db.Select("file_id").Where("id IN ?", attachmentIDs).Find(&rows).Error; err != nil {
		return nil
	}
	ids := make([]uint, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row.FileID)
	}
	return ids
}

// 内容引用到图片时，更新文件的使用时间与累计次数。
// 历史图库靠 admin_attachments 保留，删内容不删文件，无需单独的引用表。
func touchFiles(db *gorm.DB, attachmentIDs []uint) {
	ids := referencedFileIDs(db, attachmentIDs)
	if len(ids) > 0 {
		_ = db.Model(&sdk.File{}).Where("id IN ?", ids).Updates(map[string]any{
			"updated_at": time.Now().UnixMilli(),
			"used_count": gorm.Expr("used_count + 1"),
		}).Error
	}
}

// 「我的图片」列表与取图：列表按当前登录用户过滤，取图只放行用户上传的图片附件。
func registerMomentImages(a *admin.App) {
	uploader := &controllers.Controller{App: a}
	route.Get("/api/moment-images/mine", func(c *gin.Context) {
		s, ok := siteSession(a, c, true)
		if !ok {
			return
		}
		db := a.DB.WithContext(c.Request.Context()).Model(&sdk.Attachment{}).
			Where("scope = ? AND created_by = ?", fmt.Sprintf("user:%d", s.User.ID), s.User.ID).
			Where("file_id IN (?)", a.DB.Model(&sdk.File{}).Select("id").Where("mime IN ?", imageMIMEs))
		if keyword := strings.TrimSpace(c.Query("keyword")); keyword != "" {
			db = db.Where("name LIKE ?", "%"+keyword+"%")
		}
		var rows []sdk.Attachment
		if err := db.Preload("File").Order("id DESC").Limit(60).Find(&rows).Error; err != nil {
			response.Report(c, err)
			return
		}
		out := make([]gin.H, 0, len(rows))
		for _, row := range rows {
			out = append(out, gin.H{
				"id": row.ID, "name": row.Name, "url": fmt.Sprintf("/api/moment-images/%d/content", row.ID),
				"created_at": row.CreatedAt, "size": row.File.Size,
			})
		}
		response.Respond(c, out)
	})
	route.Get("/api/moment-images/{id}/content", func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil || id == 0 {
			c.Status(http.StatusNotFound)
			return
		}
		var row sdk.Attachment
		err = a.DB.WithContext(c.Request.Context()).Preload("File").
			Where("id = ? AND scope LIKE 'user:%'", uint(id)).First(&row).Error
		if err != nil || !slices.Contains(imageMIMEs, row.File.MIME) {
			c.Status(http.StatusNotFound)
			return
		}
		uploader.ServeAttachment(c, row)
	})
}
