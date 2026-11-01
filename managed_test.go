package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"syscall"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/dawnim/dawn-server/support/supervisor"
	"github.com/gorilla/websocket"
)

// TestManagedApplication 验证实际二进制、配置、Horizon 和 HTTP/WS 进程交接。
func TestManagedApplication(t *testing.T) {
	dir, err := os.MkdirTemp("/tmp", "dawn-managed-")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.RemoveAll(dir) })
	binary := filepath.Join(dir, "server")
	if output, err := exec.Command("go", "build", "-o", binary, ".").CombinedOutput(); err != nil {
		t.Fatalf("build: %s %v", output, err)
	}
	r := miniredis.RunT(t)
	clockCtx, stopClock := context.WithCancel(context.Background())
	defer stopClock()
	go func() {
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-clockCtx.Done():
				return
			case <-ticker.C:
				r.FastForward(100 * time.Millisecond)
			}
		}
	}()
	ports := []string{}
	listeners := []net.Listener{}
	for range 3 {
		l, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			t.Fatal(err)
		}
		listeners = append(listeners, l)
		ports = append(ports, strconv.Itoa(l.Addr().(*net.TCPAddr).Port))
	}
	for _, l := range listeners {
		l.Close()
	}
	root, _ := os.Getwd()
	if err := os.Symlink(filepath.Join(root, "admin"), filepath.Join(dir, "admin")); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(root, "database"), filepath.Join(dir, "database")); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".env"), []byte("DAWN_ADMIN_PASSWORD=configuration-password\n"), 0600); err != nil {
		t.Fatal(err)
	}
	env := append(os.Environ(), "APP_NAME=managed-test", "APP_ENV=testing", "APP_DEBUG=false", "APP_KEY=base64:MDEyMzQ1Njc4OWFiY2RlZjAxMjM0NTY3ODlhYmNkZWY=",
		"GEOIP_ENABLED=false", "SERVER_HOST=127.0.0.1", "SERVER_PORT="+ports[0], "SERVER_ACCESS_LOG=false", "SERVER_SHUTDOWN_TIMEOUT=5s", "WS_ADDR=127.0.0.1:"+ports[1],
		"DB_CONNECTION=sqlite", "DB_DATABASE="+filepath.Join(dir, "data.sqlite"), "REDIS_URL=redis://"+r.Addr(), "LOG_CHANNEL=single",
		"FILESYSTEM_LOCAL_ROOT="+filepath.Join(dir, "files"), "HORIZON_CONCURRENCY=1", "HORIZON_HEARTBEAT_TTL=2", "HORIZON_WORKER_TIMEOUT=3", "DAWN_FRONTEND_PORT="+ports[2])
	install := exec.Command(binary, "dawn:install")
	install.Dir, install.Env = dir, env
	if output, err := install.CombinedOutput(); err != nil {
		t.Fatalf("install: %s %v", output, err)
	}
	output, _ := os.Create(filepath.Join(dir, "console.log"))
	defer output.Close()
	cmd := exec.Command(binary, "dawn:start", "all")
	cmd.Dir, cmd.Env, cmd.Stdout, cmd.Stderr = dir, env, output, output
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	t.Cleanup(func() {
		_ = cmd.Process.Signal(syscall.SIGTERM)
		select {
		case <-done:
		case <-time.After(15 * time.Second):
			cmd.Process.Kill()
			<-done
			t.Error("manager did not stop")
		}
		if t.Failed() {
			data, _ := os.ReadFile(filepath.Join(dir, "console.log"))
			t.Log(string(data))
		}
	})
	socket := filepath.Join(dir, "storage/framework/dawn.sock")
	control := func(name, action string) supervisor.Response {
		t.Helper()
		response, connected, err := supervisor.Control(context.Background(), socket, name, action)
		if !connected || err != nil {
			t.Fatalf("%s %s: %v", name, action, err)
		}
		return response
	}
	await := func(check func() bool) {
		t.Helper()
		deadline := time.Now().Add(30 * time.Second)
		for time.Now().Before(deadline) {
			if check() {
				return
			}
			time.Sleep(100 * time.Millisecond)
		}
		t.Fatal("services did not converge")
	}
	await(func() bool {
		response, connected, err := supervisor.Control(context.Background(), socket, "all", "status")
		if !connected || err != nil {
			return false
		}
		count := 0
		for _, s := range response.Services {
			if s.State == "running" {
				count++
			}
		}
		return count == 4
	})
	base := "http://127.0.0.1:" + ports[0]
	client := &http.Client{Timeout: 5 * time.Second}
	call := func(method, path, token string, body any) map[string]any {
		t.Helper()
		data, _ := json.Marshal(body)
		request, _ := http.NewRequest(method, base+"/admin"+path, bytes.NewReader(data))
		request.Header.Set("Content-Type", "application/json")
		if token != "" {
			request.Header.Set("Authorization", "Bearer "+token)
		}
		response, err := client.Do(request)
		if err != nil {
			t.Fatal(err)
		}
		defer response.Body.Close()
		var result map[string]any
		if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
			t.Fatal(err)
		}
		if response.StatusCode != 200 {
			t.Fatal(result)
		}
		return result["data"].(map[string]any)
	}
	captcha := call("GET", "/auth/captcha", "", nil)
	uuid := captcha["uuid"].(string)
	answer, _ := r.Get("managed-test:testing:temp:admin:captcha:" + uuid)
	token := call("POST", "/auth/login", "", map[string]any{"username": "admin", "password": "configuration-password", "uuid": uuid, "code": answer})["access_token"].(string)
	connect := func(resume string) (*websocket.Conn, string) {
		t.Helper()
		ticket := call("POST", "/realtime/ticket", token, map[string]any{"resume": resume})["ticket"].(string)
		conn, _, err := websocket.DefaultDialer.Dial("ws://127.0.0.1:"+ports[1]+"/", nil)
		if err != nil {
			t.Fatal(err)
		}
		conn.SetReadDeadline(time.Now().Add(8 * time.Second))
		conn.WriteJSON(map[string]any{"type": "auth", "ticket": ticket})
		var ready struct {
			Type string `json:"type"`
			Data struct {
				Resume string `json:"resume"`
			} `json:"data"`
		}
		if err := conn.ReadJSON(&ready); err != nil {
			t.Fatal(err)
		}
		return conn, ready.Data.Resume
	}
	conn, resume := connect("")
	defer conn.Close()
	oldHTTP := control("http", "status").Services[0].PID
	control("http", "reload")
	if control("http", "status").Services[0].PID == oldHTTP {
		t.Fatal("HTTP did not reload")
	}
	call("GET", "/auth/me", token, nil)
	control("ws", "reload")
	if _, _, err := conn.ReadMessage(); !websocket.IsCloseError(err, websocket.CloseServiceRestart) {
		t.Fatalf("WS did not drain: %v", err)
	}
	next, saved := connect(resume)
	defer next.Close()
	if saved != resume {
		t.Fatal("WS session was not restored")
	}
	oldWS := control("ws", "status").Services[0].PID
	if err := syscall.Kill(oldWS, syscall.SIGKILL); err != nil {
		t.Fatal(err)
	}
	await(func() bool {
		s := control("ws", "status").Services[0]
		return s.State == "running" && s.PID != oldWS && s.Restarts > 0
	})
	call("GET", "/auth/me", token, nil)
	oldHorizon := control("horizon", "status").Services[0].PID
	control("horizon", "stop")
	if control("horizon", "status").Services[0].Wanted {
		t.Fatal("Horizon still wanted")
	}
	// 原生租约到期前保持退避；其他服务不受影响。
	_, _, _ = supervisor.Control(context.Background(), socket, "horizon", "start")
	await(func() bool {
		s := control("horizon", "status").Services[0]
		return s.State == "running" && s.PID != oldHorizon
	})
	horizonPID := control("horizon", "status").Services[0].PID
	if err := syscall.Kill(horizonPID, syscall.SIGKILL); err != nil {
		t.Fatal(err)
	}
	await(func() bool {
		s := control("horizon", "status").Services[0]
		return s.State == "running" && s.PID != horizonPID && s.Restarts > 0
	})
	call("GET", "/auth/me", token, nil)
	if _, err := os.Stat(filepath.Join(dir, "storage/logs/dawn.log")); err != nil {
		t.Fatal("default dawn.log was not created", err)
	}
	// local 在同一管理器内补启动前端，不重复启动后端服务。
	if _, err := os.Stat(filepath.Join(root, "admin/node_modules/.bin/vite")); err == nil {
		control("local", "start")
		response, err := client.Get("http://127.0.0.1:" + ports[2])
		if err != nil {
			t.Fatal(err)
		}
		body, _ := io.ReadAll(response.Body)
		response.Body.Close()
		if response.StatusCode != 200 || !bytes.Contains(body, []byte("/src/main.ts")) {
			t.Fatal("frontend did not serve application")
		}
		oldFrontend := control("frontend", "status").Services[0].PID
		if err := syscall.Kill(oldFrontend, syscall.SIGKILL); err != nil {
			t.Fatal(err)
		}
		await(func() bool {
			s := control("frontend", "status").Services[0]
			return s.State == "running" && s.PID != oldFrontend && s.Restarts > 0
		})
	}
	// 管理器被强杀后，子服务通过继承管道感知断开并回收监听资源和前端后代。
	managed := control("all", "status").Services
	if err := cmd.Process.Kill(); err != nil {
		t.Fatal(err)
	}
	r.Close() // Redis同时不可用时，收尾仍受进程退出期限约束。
	await(func() bool {
		for _, service := range managed {
			if service.PID > 0 && syscall.Kill(service.PID, 0) == nil {
				return false
			}
		}
		for _, port := range ports {
			conn, err := net.DialTimeout("tcp", "127.0.0.1:"+port, 100*time.Millisecond)
			if err == nil {
				conn.Close()
				return false
			}
		}
		return true
	})
}
