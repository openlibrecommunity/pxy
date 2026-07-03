package guiservice

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/openlibrecommunity/pxy/internal/installer"
	pxyssh "github.com/openlibrecommunity/pxy/internal/ssh"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

var ErrRunning = errors.New("install already running")

// App is the Wails API bound to frontend.
type App struct {
	ctx     context.Context
	mu      sync.Mutex
	running bool
}

// New creates an App.
func New() *App { return &App{} }

// Startup stores Wails context.
func (a *App) Startup(ctx context.Context) { a.ctx = ctx }

// TestSSH checks server auth.
func (a *App) TestSSH(req installer.Request) string {
	req.Defaults()
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	if err := pxyssh.New(req.Host, req.SSHPort, req.User, req.Password).Test(ctx); err != nil {
		return "err: " + err.Error()
	}
	return "ok"
}

// Install runs selected protocol installers on the remote server.
func (a *App) Install(req installer.Request) (string, error) {
	req.Defaults()
	if err := validate(req); err != nil {
		return "", err
	}
	if !a.begin() {
		return "", ErrRunning
	}
	defer a.end()
	log := func(line string) { a.emit("install:log", line) }
	log("connect " + req.User + "@" + req.Host + ":" + req.SSHPort)
	script := installer.Script(req)
	encoded := base64.StdEncoding.EncodeToString([]byte(script))
	client := pxyssh.New(req.Host, req.SSHPort, req.User, req.Password)
	ctx := context.Background()
	cmd := "mkdir -p /root/pxy && printf %s " + shellArg(encoded) + " | base64 -d >/root/pxy/install.sh && chmod +x /root/pxy/install.sh"
	if _, err := client.Run(ctx, cmd, log); err != nil {
		return "", err
	}
	out, err := client.Run(ctx, "bash /root/pxy/install.sh", log)
	if err != nil {
		return out, err
	}
	res := resultPart(out)
	if res == "" {
		res = out
	}
	a.emit("install:done", res)
	return res, nil
}

func (a *App) begin() bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.running {
		return false
	}
	a.running = true
	return true
}

func (a *App) end() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.running = false
}

func (a *App) emit(name, msg string) {
	if a.ctx == nil {
		return
	}
	runtime.EventsEmit(a.ctx, name, msg)
}

func validate(req installer.Request) error {
	if req.Host == "" || req.Password == "" || req.Domain == "" {
		return fmt.Errorf("host, pass, domain required")
	}
	if !req.Protocols.VLESS && !req.Protocols.Hysteria2 && !req.Protocols.Mieru && !req.Protocols.AmneziaWG && !req.Protocols.Naive && !req.Protocols.OLCRTC {
		return fmt.Errorf("select at least one protocol")
	}
	return nil
}

func resultPart(out string) string {
	_, after, ok := strings.Cut(out, "PXYSTARTRESULT")
	if !ok {
		return ""
	}
	return strings.TrimSpace(after)
}

func shellArg(s string) string { return "'" + strings.ReplaceAll(s, "'", "'\\''") + "'" }
