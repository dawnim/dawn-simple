// Package web 提供示例官网，注册登录与客服能力直接使用扩展公开接口；社区编辑器资源随官网一起嵌入。
package web

import (
	"crypto/sha256"
	"embed"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/dawnim/dawn-server/app/admin/response"
	"github.com/dawnim/dawn-server/constants"

	"github.com/dawnim/dawn-server/app/admin"
	"github.com/dawnim/dawn-server/app/models"
	"github.com/dawnim/dawn-server/support/customer"
	appconst "github.com/dawnim/dawn-simple/app/constants"
	"github.com/gin-gonic/gin"
	"github.com/prismgo/framework/route"
	"github.com/prismgo/framework/support"
)

//go:embed assets/*
var files embed.FS

func Register(a *admin.App) {
	registerMoments(a)
	registerMomentImages(a)
	registerMomentAI(a)
	registerMomentChat(a)
	route.Get("/favicon.ico", func(c *gin.Context) { c.Redirect(302, "/site-assets/logo.svg") })
	page := func(c *gin.Context) {
		data, _ := files.ReadFile("assets/index.html")
		html := string(data)
		view := "moments"
		if strings.HasPrefix(c.Request.URL.Path, "/docs/") {
			view = "docs"
		} else if c.Request.URL.Path == "/dawn" {
			view = "home"
		}
		html = strings.Replace(html, `data-page="moments"`, `data-page="`+view+`"`, 1)
		for _, name := range []string{"site.js", "site.css", "moment-markdown.js", "moment-markdown.css", "logo.svg"} {
			asset, _ := files.ReadFile("assets/" + name)
			html = strings.ReplaceAll(html, name+`"`, fmt.Sprintf(`%s?v=%x"`, name, sha256.Sum256(asset)))
		}
		// 官网使用公共脚本的可选能力；内容版本避免升级后仍执行旧缓存。
		widget, _ := customer.Web.ReadFile("web/widget.js")
		html = strings.ReplaceAll(html, `/support-assets/widget.js"`, fmt.Sprintf(`/support-assets/widget.js?v=%x"`, sha256.Sum256(widget)))
		c.Header("Cache-Control", "no-cache")
		c.Data(200, "text/html; charset=utf-8", []byte(html))
	}
	route.Get("/docs", func(c *gin.Context) { c.Redirect(302, "/docs/") })
	docs := os.DirFS(support.PublicPath("docs"))
	asset, _ := files.ReadFile("assets/docs-support.js")
	scripts := fmt.Sprintf(`<script src="/site-assets/docs-support.js?v=%x" defer></script>`, sha256.Sum256(asset))
	// 官网外壳常驻客服；iframe通过__content加载后移除内部参数，构建产物保持原样。
	route.Get("/docs/{filepath}", func(c *gin.Context) {
		name := strings.TrimPrefix(c.Param("filepath"), "/")
		if name == "" || strings.HasSuffix(name, "/") {
			name += "index.html"
		} else if path.Ext(name) == "" {
			name += ".html"
		}
		if !fs.ValidPath(name) {
			c.Status(404)
			return
		}
		info, err := fs.Stat(docs, name)
		if err != nil || info.IsDir() {
			c.Status(404)
			return
		}
		if !strings.HasSuffix(name, ".html") {
			c.FileFromFS(name, http.FS(docs))
			return
		}
		if c.Query("__content") != "1" && c.GetHeader("Sec-Fetch-Dest") != "iframe" {
			page(c)
			return
		}
		data, err := fs.ReadFile(docs, name)
		if err != nil {
			c.Status(404)
			return
		}
		if a != nil {
			reading := customer.Raw(map[string]string{"path": "/docs/" + name, "region": ipRegion(c)})
			key := a.Key(appconst.TempMomentReading)
			pipe := a.Redis.Pipeline()
			pipe.LPush(c.Request.Context(), key, reading)
			pipe.LTrim(c.Request.Context(), key, 0, 99)
			pipe.Expire(c.Request.Context(), key, appconst.TempMomentReading.TTL)
			_, _ = pipe.Exec(c.Request.Context())
		}
		c.Header("Cache-Control", "no-cache")
		c.Data(200, "text/html; charset=utf-8", []byte(strings.Replace(string(data), "</body>", scripts+"</body>", 1)))
	}).Where("filepath", ".*")
	route.Get("/", page)
	route.Get("/dawn", page)
	route.Get("/moments", page)
	route.Get("/moments/{id}", page)
	route.Get("/site-assets/{name}", func(c *gin.Context) {
		name := c.Param("name")
		kind := "text/javascript; charset=utf-8"
		switch name {
		case "site.js", "docs-support.js", "spark-md5.min.js", "moment-markdown.js":
		case "spark-md5.LICENSE":
			kind = "text/plain; charset=utf-8"
		case "site.css", "moment-markdown.css":
			kind = "text/css; charset=utf-8"
		case "logo.svg", "support.svg", "chat.svg":
			kind = "image/svg+xml"
		default:
			c.Status(404)
			return
		}
		data, err := files.ReadFile("assets/" + name)
		if err != nil {
			c.Status(404)
			return
		}
		if c.Query("v") == fmt.Sprintf("%x", sha256.Sum256(data)) {
			c.Header("Cache-Control", "public, max-age=31536000, immutable")
		} else {
			c.Header("Cache-Control", "no-cache")
		}
		c.Data(200, kind, data)
	})
	route.Post("/api/support-identities", func(c *gin.Context) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 4096)
		var input struct {
			VisitorTicket string `json:"visitor_ticket"`
		}
		if c.ShouldBindJSON(&input) != nil {
			response.Report(c, customer.ErrInvalid)
			return
		}
		count, err := a.Redis.Eval(c.Request.Context(), `local n=redis.call('INCR',KEYS[1]);if n==1 then redis.call('EXPIRE',KEYS[1],60) end;return n`, []string{a.Key(constants.TempSupportRate, customer.TokenHash(c.ClientIP()+":site-identity"))}).Int()
		if err != nil {
			response.Report(c, err)
			return
		}
		if count > 120 {
			response.Report(c, customer.ErrLimited)
			return
		}
		// 官网只接入自己的演示应用，不接受客户端指定其他客服应用。
		var app models.SupportApp
		if err = a.DB.WithContext(c.Request.Context()).Where("public_key = ? AND status = ?", "dawn", models.Enabled).First(&app).Error; err != nil {
			response.Report(c, err)
			return
		}
		session, ticket, err := a.Customer.LocalIdentity(c.Request.Context(), app.ID, strings.TrimPrefix(c.GetHeader("Authorization"), "Bearer "), input.VisitorTicket)
		if err != nil {
			response.Report(c, err)
			return
		}
		response.Created(c, gin.H{"user": session.User.Profile(), "token": session.Token, "support": gin.H{"ticket": ticket}}, c.Request.URL.Path)
	})
	route.Get("/api/site", func(c *gin.Context) {
		var app models.SupportApp
		var appID uint
		key := ""
		if a.DB.WithContext(c.Request.Context()).Where("public_key = ? AND status = ?", "dawn", models.Enabled).First(&app).Error == nil {
			appID = app.ID
			key = app.PublicKey
		}
		c.JSON(200, gin.H{"data": gin.H{"support_app_id": appID, "support_key": key, "brand": "Dawn"}})
	})
}
