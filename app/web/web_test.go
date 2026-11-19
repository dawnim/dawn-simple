package web

import (
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/prismgo/framework/container"
	"github.com/prismgo/framework/route"
)

func TestDocsSupport(t *testing.T) {
	root := t.TempDir()
	html := "<!doctype html><html><body><div id=app>docs</div></body></html>"
	fixtures := map[string]string{
		"docs/index.html": html, "docs/prismgo/session.html": html,
		"docs/assets/app.js": "const docs = true;", "private.txt": "private",
	}
	for name, data := range fixtures {
		path := filepath.Join(root, name)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(data), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	registry := container.NewContainer()
	container.SetProvider(func() *container.Container { return registry })
	t.Cleanup(func() { container.SetProvider(nil) })
	if err := registry.Instance("path.public", root); err != nil {
		t.Fatal(err)
	}
	if err := registry.Instance("route.router", route.New()); err != nil {
		t.Fatal(err)
	}
	Register(nil)
	engine := gin.New()
	if err := route.Mount(engine); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		path        string
		status      int
		contentType string
		iframe      bool
	}{
		{"/docs/", 200, "text/html", false},
		{"/docs/prismgo/session.html", 200, "text/html", false},
		{"/docs/prismgo/session", 200, "text/html", false},
		{"/docs/prismgo/session.html", 200, "text/html", true},
		{"/docs/prismgo/session.html?__content=1", 200, "text/html", false},
		{"/docs/assets/app.js", 200, "javascript", false},
		{"/site-assets/docs-support.js", 200, "text/javascript", false},
		{"/site-assets/logo.svg", 200, "image/svg+xml", false},
		{"/docs/missing.html", 404, "", false},
		{"/docs/../private.txt", 404, "", false},
		{"/docs/%2e%2e/private.txt", 404, "", false},
		{"/docs/assets", 404, "", false},
	} {
		t.Run(tc.path, func(t *testing.T) {
			response := httptest.NewRecorder()
			request := httptest.NewRequest("GET", tc.path, nil)
			if tc.iframe {
				request.Header.Set("Sec-Fetch-Dest", "iframe")
			}
			engine.ServeHTTP(response, request)
			if response.Code != tc.status || !strings.Contains(response.Header().Get("Content-Type"), tc.contentType) {
				t.Fatalf("status=%d type=%q", response.Code, response.Header().Get("Content-Type"))
			}
			body := response.Body.String()
			if tc.contentType == "text/html" {
				if tc.iframe || strings.Contains(tc.path, "__content=1") {
					if !strings.Contains(body, "<div id=app>docs</div>") || strings.Contains(body, "/support-assets/widget.js") || !strings.Contains(body, "/site-assets/docs-support.js?v=") {
						t.Fatal("文档桥接脚本不完整或重复挂载客服")
					}
				} else if !strings.Contains(body, `id="documentation"`) || !strings.Contains(body, `data-page="docs"`) || strings.Count(body, "/support-assets/widget.js") != 1 {
					t.Fatal("官网外壳或客服脚本不完整")
				}
			}
			if tc.path == "/docs/assets/app.js" && body != fixtures["docs/assets/app.js"] {
				t.Fatal("文档静态资源被修改")
			}
		})
	}
	for name, expected := range fixtures {
		data, err := os.ReadFile(filepath.Join(root, name))
		if err != nil || string(data) != expected {
			t.Fatalf("源文件发生变化: %s", name)
		}
	}
}
