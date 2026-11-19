package web

import (
	"encoding/json"
	"io/fs"
	"os"
	"path"
	"strings"
	"time"

	"github.com/dawnim/dawn-server/app/admin"
	"github.com/dawnim/dawn-server/app/admin/response"
	"github.com/dawnim/dawn-server/app/models"
	"github.com/dawnim/dawn-server/support/customer"
	appconst "github.com/dawnim/dawn-simple/app/constants"
	"github.com/gin-gonic/gin"
	"github.com/prismgo/framework/route"
	"github.com/prismgo/framework/support"
	"golang.org/x/net/html"
)

func registerMomentAI(a *admin.App) {
	route.Post("/api/moment-drafts", func(c *gin.Context) {
		if _, ok := siteSession(a, c, true); !ok {
			return
		}
		var input struct {
			Path string `json:"path"`
		}
		if c.ShouldBindJSON(&input) != nil || !strings.HasPrefix(input.Path, "/docs/") {
			response.Report(c, customer.ErrInvalid)
			return
		}
		docText, err := momentDocument(input.Path)
		if err != nil {
			response.Report(c, err)
			return
		}
		var app models.SupportApp
		if err = a.DB.WithContext(c.Request.Context()).Where("public_key = ? AND status = ?", "dawn", models.Enabled).First(&app).Error; err != nil {
			response.Report(c, err)
			return
		}
		raw, err := a.Setting(c.Request.Context(), "moments.ai_daily_limit")
		if err != nil {
			response.Report(c, err)
			return
		}
		var limit int64
		if err = json.Unmarshal(raw, &limit); err != nil {
			response.Report(c, err)
			return
		}
		now := time.Now()
		end := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
		day := now.Format("20060102")
		keys := []string{a.Key(appconst.TempMomentAI, day, "total"), a.Key(appconst.TempMomentAI, day, customer.TokenHash(c.ClientIP()))}
		allowed, err := a.Redis.Eval(c.Request.Context(), `if tonumber(redis.call('GET',KEYS[1]) or '0') >= tonumber(ARGV[1]) or tonumber(redis.call('GET',KEYS[2]) or '0') >= 2 then return 0 end for _,k in ipairs(KEYS) do redis.call('INCR',k) redis.call('EXPIRE',k,ARGV[2]) end return 1`, keys, limit, int64(time.Until(end).Seconds())+1).Int()
		if err != nil {
			response.Report(c, err)
			return
		}
		if allowed == 0 {
			response.Report(c, customer.ErrLimited)
			return
		}
		result, err := a.Customer.GenerateText(c.Request.Context(), app.ID, "根据文档创作一条有趣、俏皮的动态，使用用户的第一人称，不超过300字。只输出正文，不编造文档内容。", docText)
		if err != nil {
			response.Report(c, err)
			return
		}
		resultText := []rune(strings.TrimSpace(result))
		if len(resultText) > 300 {
			resultText = resultText[:300]
		}
		response.Respond(c, gin.H{"content": string(resultText)})
	})
}

func momentDocument(documentPath string) (string, error) {
	if !strings.HasPrefix(documentPath, "/docs/") {
		return "", customer.ErrInvalid
	}
	name := strings.TrimPrefix(documentPath, "/docs/")
	if strings.HasSuffix(name, "/") {
		name += "index.html"
	} else if path.Ext(name) == "" {
		name += ".html"
	}
	if !fs.ValidPath(name) || !strings.HasSuffix(name, ".html") {
		return "", customer.ErrInvalid
	}
	file, err := os.OpenRoot(support.PublicPath("docs"))
	if err != nil {
		return "", err
	}
	defer file.Close()
	f, err := file.Open(name)
	if err != nil {
		return "", err
	}
	defer f.Close()
	doc, err := html.Parse(f)
	if err != nil {
		return "", err
	}
	for node := range doc.Descendants() {
		for _, attr := range node.Attr {
			if attr.Key == "class" && strings.Contains(" "+attr.Val+" ", " vp-doc ") {
				doc = node
				break
			}
		}
		if doc == node {
			break
		}
	}
	var content strings.Builder
	var visit func(*html.Node)
	visit = func(n *html.Node) {
		if n.Type == html.ElementNode && (n.Data == "script" || n.Data == "style" || n.Data == "nav") {
			return
		}
		if n.Type == html.TextNode {
			content.WriteString(n.Data)
			content.WriteByte(' ')
		}
		for child := n.FirstChild; child != nil; child = child.NextSibling {
			visit(child)
		}
	}
	visit(doc)
	text := []rune(strings.Join(strings.Fields(content.String()), " "))
	if len(text) > 10000 {
		text = text[:10000]
	}
	return string(text), nil
}
