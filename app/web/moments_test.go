package web

import (
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"testing"

	"github.com/dawnim/dawn-server/app/admin"
	sdk "github.com/dawnim/dawn-server/app/models"
	"github.com/dawnim/dawn-simple/app/models"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/prismgo/framework/container"
	"github.com/prismgo/framework/route"
	"gorm.io/gorm"
)

func TestPublicMomentsExcludeDeleted(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&sdk.User{}, &models.Moment{}, &models.MomentComment{}, &models.MomentReaction{}); err != nil {
		t.Fatal(err)
	}
	user := sdk.User{UniqueID: "7654321", Nickname: "Tester", Type: sdk.UserMember}
	if err := db.Session(&gorm.Session{SkipHooks: true}).Create(&user).Error; err != nil {
		t.Fatal(err)
	}
	live := models.Moment{UserID: user.ID, Content: "visible", Images: "[]"}
	deleted := models.Moment{UserID: user.ID, Content: "deleted", Images: "[]"}
	if err := db.Create(&live).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&deleted).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Delete(&deleted).Error; err != nil {
		t.Fatal(err)
	}
	comment := models.MomentComment{MomentID: live.ID, UserID: user.ID, Content: "removed"}
	if err := db.Create(&comment).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Delete(&comment).Error; err != nil {
		t.Fatal(err)
	}
	reply := models.MomentComment{MomentID: live.ID, UserID: user.ID, ReplyID: comment.ID, Content: "reply"}
	if err := db.Create(&reply).Error; err != nil {
		t.Fatal(err)
	}
	registry := container.NewContainer()
	container.SetProvider(func() *container.Container { return registry })
	t.Cleanup(func() { container.SetProvider(nil) })
	if err := registry.Instance("route.router", route.New()); err != nil {
		t.Fatal(err)
	}
	registerMoments(&admin.App{DB: db})
	engine := gin.New()
	engine.Use(func(c *gin.Context) { c.Set("admin.locale", "zh"); c.Next() })
	if err := route.Mount(engine); err != nil {
		t.Fatal(err)
	}
	response := httptest.NewRecorder()
	engine.ServeHTTP(response, httptest.NewRequest("GET", "/api/moments", nil))
	var result struct {
		Data []momentView `json:"data"`
		Meta struct {
			Total int `json:"total"`
		} `json:"meta"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if response.Code != 200 || result.Meta.Total != 1 || len(result.Data) != 1 || result.Data[0].Content != "visible" || result.Data[0].Comments != 1 {
		t.Fatalf("unexpected response %d %s", response.Code, response.Body.String())
	}
	response = httptest.NewRecorder()
	engine.ServeHTTP(response, httptest.NewRequest("GET", "/api/moments/"+fmt.Sprint(deleted.ID), nil))
	if response.Code != 404 {
		t.Fatalf("deleted post: %d %s", response.Code, response.Body.String())
	}
	response = httptest.NewRecorder()
	engine.ServeHTTP(response, httptest.NewRequest("GET", fmt.Sprintf("/api/moments/%d/comments", live.ID), nil))
	var comments struct {
		Data []commentView `json:"data"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &comments); err != nil {
		t.Fatal(err)
	}
	if response.Code != 200 || len(comments.Data) != 1 || comments.Data[0].ReplyName != "Tester" {
		t.Fatalf("reply to deleted comment: %d %s", response.Code, response.Body.String())
	}
}
