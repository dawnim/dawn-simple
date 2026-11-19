package web

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/alicebob/miniredis/v2"
	dawn "github.com/dawnim/dawn-server"
	"github.com/dawnim/dawn-server/app/admin"
	sdk "github.com/dawnim/dawn-server/app/models"
	"github.com/dawnim/dawn-server/database/schema"
	"github.com/dawnim/dawn-server/database/seeders"
	"github.com/dawnim/dawn-server/support/customer"
	appadmin "github.com/dawnim/dawn-simple/app/admin"
	appconst "github.com/dawnim/dawn-simple/app/constants"
	"github.com/dawnim/dawn-simple/app/models"
	"github.com/gin-gonic/gin"
	"github.com/prismgo/framework/config"
	"github.com/prismgo/framework/encryption"
	"github.com/prismgo/framework/foundation"
)

func newCommunityTestApp(t *testing.T) *admin.App {
	t.Helper()
	store := miniredis.RunT(t)
	root := t.TempDir()
	configs := map[string]map[string]any{
		"app":        {"name": "community-test", "env": "testing", "locale": "zh", "key": "base64:MDEyMzQ1Njc4OWFiY2RlZjAxMjM0NTY3ODlhYmNkZWY=", "server": map[string]any{"host": "127.0.0.1", "port": 5129}},
		"dawn":       {"module": "example.com/community", "geoip": map[string]any{"enabled": false}},
		"database":   {"default": "test", "connections": map[string]any{"test": map[string]any{"driver": "sqlite", "dsn": ":memory:", "max_open_conns": 1}}, "redis": map[string]any{"default": map[string]any{"url": "redis://" + store.Addr()}}},
		"session":    {"driver": "redis", "connection": "default", "prefix": "community-test:session", "encoding": "json"},
		"cache":      {"default": "redis", "prefix": "community-test:", "stores": map[string]any{"redis": map[string]any{"driver": "redis", "connection": "default"}}},
		"horizon":    {"path": "/admin/horizon"},
		"filesystem": {"default": "local", "disks": map[string]any{"local": map[string]any{"driver": "local", "root": root + "/uploads", "visibility": "private"}}},
		"logging":    {"default": "stderr", "channels": map[string]any{"stderr": map[string]any{"driver": "stderr", "level": "error"}}},
	}
	for key, value := range configs {
		config.Add(key, func() map[string]any { return value })
	}
	app := foundation.Configure(root).WithExtensionProviders(encryption.ServiceProvider{}, dawn.ServiceProvider{Configure: func(o *admin.Options) { appadmin.Configure(o); o.Storage = root }, Schema: func(d *schema.Definition) {
		d.Tables["moments"] = schema.Table{Model: &models.Moment{}}
		d.Tables["moment_comments"] = schema.Table{Model: &models.MomentComment{}}
		d.Tables["moment_reactions"] = schema.Table{Model: &models.MomentReaction{}}
	}}).WithRouting(func(r *foundation.Routing) {
		r.Routes(func(app *foundation.Application, engine *gin.Engine) error {
			if err := dawn.RegisterRoutes(app, engine); err != nil {
				return err
			}
			a, err := admin.From(app.Container())
			if err != nil {
				return err
			}
			Register(a)
			return nil
		})
	}).Create()
	t.Cleanup(func() { app.Close() })
	if err := app.Boot(); err != nil {
		t.Fatal(err)
	}
	a, err := admin.From(app.Container())
	if err != nil {
		t.Fatal(err)
	}
	if err = schema.Resolve().Create(a.DB); err != nil {
		t.Fatal(err)
	}
	if err = seeders.Seed(a.DB); err != nil {
		t.Fatal(err)
	}
	support := sdk.SupportApp{PublicKey: "dawn", Name: "Dawn", Config: customer.Raw(customer.Config{AllowLocalUsers: true})}
	if err = a.Customer.CreateApp(context.Background(), &support, customer.Secrets{}); err != nil {
		t.Fatal(err)
	}
	return a
}

func TestMomentAIBudget(t *testing.T) {
	a := newCommunityTestApp(t)
	ctx := context.Background()
	docs := filepath.Join(a.Options.Storage, "public", "docs")
	if err := os.MkdirAll(docs, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(docs, "article.html"), []byte(`<aside>NAVIGATION_ONLY</aside><main class="vp-doc"><h1>正文资料</h1><p>队列用于异步处理</p></main>`), 0o644); err != nil {
		t.Fatal(err)
	}
	var calls atomic.Int32
	model := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		body, _ := io.ReadAll(r.Body)
		if r.URL.Path != "/v1/responses" || !strings.Contains(string(body), "正文资料") || strings.Contains(string(body), "NAVIGATION_ONLY") {
			t.Errorf("unexpected model input: %s %s", r.URL.Path, body)
		}
		w.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprint(w, "data: {\"type\":\"response.output_text.delta\",\"delta\":\"今天终于把队列安排明白了\"}\n\ndata: {\"type\":\"response.completed\",\"response\":{\"id\":\"test\",\"output\":[]}}\n\n")
	}))
	defer model.Close()
	var support sdk.SupportApp
	if err := a.DB.First(&support).Error; err != nil {
		t.Fatal(err)
	}
	cfg := customer.Config{AllowLocalUsers: true, ModelProtocol: "responses", Model: "test", BaseURL: model.URL}
	if err := a.DB.Model(&support).Update("config", customer.Raw(cfg)).Error; err != nil {
		t.Fatal(err)
	}
	if err := a.CacheSetting(ctx, sdk.Setting{Key: "moments", Value: `{"ai_daily_limit":3,"auto_daily_limit":1,"brave_key":""}`, Revision: 99}); err != nil {
		t.Fatal(err)
	}
	session, err := a.Customer.Register(ctx, customer.Session{}, "writer", "123456", "Writer")
	if err != nil {
		t.Fatal(err)
	}
	handler, err := a.Handler()
	if err != nil {
		t.Fatal(err)
	}
	for i, ip := range []string{"192.0.2.1", "192.0.2.1", "192.0.2.1", "192.0.2.2", "192.0.2.3"} {
		req := httptest.NewRequest("POST", "/api/moment-drafts", strings.NewReader(`{"path":"/docs/article.html"}`))
		req.RemoteAddr = ip + ":1234"
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+session.Token)
		res := httptest.NewRecorder()
		handler.ServeHTTP(res, req)
		want := http.StatusOK
		if i == 2 || i == 4 {
			want = http.StatusTooManyRequests
		}
		if res.Code != want {
			t.Fatalf("draft %d: %d %s", i, res.Code, res.Body.String())
		}
	}
	if calls.Load() != 3 {
		t.Fatal("quota allowed extra model requests", calls.Load())
	}
	a.Redis.LPush(ctx, a.Key(appconst.TempMomentReading), `{"path":"/docs/article.html","region":"北京"}`)
	for range 2 {
		if err := (&AutoMomentJob{}).Handle(ctx); err != nil {
			t.Fatal(err)
		}
	}
	var posts int64
	a.DB.Model(&models.Moment{}).Where("user_id = ?", support.AIUserID).Count(&posts)
	if posts != 1 || calls.Load() != 4 {
		t.Fatalf("automatic daily quota: posts=%d calls=%d", posts, calls.Load())
	}
}

func TestCommunityJourney(t *testing.T) {
	a := newCommunityTestApp(t)
	handler, err := a.Handler()
	if err != nil {
		t.Fatal(err)
	}
	call := func(want int, method, path, token string, input any) map[string]any {
		t.Helper()
		body, _ := json.Marshal(input)
		req := httptest.NewRequest(method, path, bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		res := httptest.NewRecorder()
		handler.ServeHTTP(res, req)
		if res.Code != want {
			t.Fatalf("%s %s: %d %s", method, path, res.Code, res.Body.String())
		}
		var result map[string]any
		if res.Body.Len() > 0 {
			if err := json.Unmarshal(res.Body.Bytes(), &result); err != nil {
				t.Fatal(err)
			}
		}
		return result
	}
	register := func(name string) string {
		return call(201, "POST", "/api/account/registrations", "", map[string]any{"username": name, "password": "123456", "nickname": name})["data"].(map[string]any)["token"].(string)
	}
	alice, bob := register("alice"), register("bob")
	post := call(201, "POST", "/api/moments", alice, map[string]any{"title": "Dawn community", "board": "showcase", "content": "Hello Dawn"})["data"].(map[string]any)
	id := int(post["id"].(float64))
	path := fmt.Sprintf("/api/moments/%d", id)
	call(401, "POST", "/api/moments", "", map[string]any{"content": "Anonymous"})
	call(404, "DELETE", path, bob, nil)
	for range 2 {
		call(204, "POST", path+"/reactions", bob, map[string]any{"kind": "flowers"})
	}
	call(204, "POST", path+"/reactions", bob, map[string]any{"kind": "eggs"})
	comment := call(201, "POST", path+"/comments", bob, map[string]any{"content": "Nice!"})["data"].(map[string]any)
	commentID := int(comment["id"].(float64))
	call(201, "POST", path+"/comments", alice, map[string]any{"content": "Thank you", "reply_id": commentID})
	var reactions int64
	a.DB.Model(&models.MomentReaction{}).Count(&reactions)
	if reactions != 1 {
		t.Fatal("duplicate reaction rows", reactions)
	}
	for _, query := range []string{"?sort_by=hot", "?board=showcase&query=Dawn", "?sort_by=created_at"} {
		result := call(200, "GET", "/api/moments"+query, "", nil)
		rows := result["data"].([]any)
		if len(rows) != 1 || rows[0].(map[string]any)["flowers"] != float64(2) {
			t.Fatalf("query: %#v", result)
		}
	}
	anonymous := call(201, "POST", "/api/moments", bob, map[string]any{"format": "post", "title": "Private notes", "content": "## Hidden\n\n```go\nfmt.Println(1)\n```", "is_anonymous": true})["data"].(map[string]any)
	anonymousView := call(200, "GET", "/api/moments/"+fmt.Sprint(int(anonymous["id"].(float64))), "", nil)["data"].(map[string]any)
	if anonymousView["is_anonymous"] != true || anonymousView["nickname"] != "匿名用户" || anonymousView["unique_id"] != "" {
		t.Fatal("anonymous author leaked", anonymousView)
	}
	ownerView := call(200, "GET", "/api/moments/"+fmt.Sprint(int(anonymous["id"].(float64))), bob, nil)["data"].(map[string]any)
	if ownerView["can_manage"] != true {
		t.Fatal("anonymous owner cannot manage post", ownerView)
	}
	call(200, "GET", "/api/moment-overview", "", nil)
	chats := call(200, "GET", "/api/community-chats", alice, nil)["data"].([]any)
	if len(chats) != 2 || chats[0].(map[string]any)["type"] != "room" {
		t.Fatal("hall must be first", chats)
	}
	if chats[1].(map[string]any)["nickname"] != "bob" {
		t.Fatal("peer profile missing", chats)
	}
	direct := int(chats[1].(map[string]any)["id"].(float64))
	chatPath := fmt.Sprintf("/api/community-chats/%d/messages", direct)
	input := map[string]any{"content": "Hello Bob", "request_id": "community-request-0001"}
	first := call(201, "POST", chatPath, alice, input)["data"].(map[string]any)
	second := call(201, "POST", chatPath, alice, input)["data"].(map[string]any)
	if first["id"] != second["id"] {
		t.Fatal("retry duplicated message")
	}
	outsider := register("outsider")
	call(404, "GET", chatPath, outsider, nil)
	messages := call(200, "GET", chatPath, bob, nil)["data"].([]any)
	if messages[0].(map[string]any)["nickname"] != "alice" || messages[0].(map[string]any)["content"] != "Hello Bob" {
		t.Fatal("sender profile or message missing", messages)
	}
	call(204, "PUT", fmt.Sprintf("/api/community-chats/%d/read", direct), bob, map[string]any{"message_id": first["id"]})
	chats = call(200, "GET", "/api/community-chats", bob, nil)["data"].([]any)
	if chats[1].(map[string]any)["unread"] != float64(0) {
		t.Fatal("read position was not applied", chats)
	}
	call(204, "DELETE", fmt.Sprintf("%s/comments/%d", path, commentID), bob, nil)
	call(201, "POST", "/api/moments", alice, map[string]any{"content": "Second"})
	call(429, "POST", "/api/moments", alice, map[string]any{"content": "Over quota"})
	call(204, "DELETE", path, alice, nil)
	call(404, "GET", path, "", nil)
	var deleted models.Moment
	if err := a.DB.Unscoped().First(&deleted, id).Error; err != nil || deleted.DeletedAt == 0 {
		t.Fatal("post was not soft deleted", err)
	}
}
