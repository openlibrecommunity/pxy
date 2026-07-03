// Package ssh provides an SSH client for running remote install commands.
package ssh

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"
)

var ErrConnect = errors.New("ssh connect failed")

// LogFunc receives streamed line output from remote commands.
type LogFunc func(line string)

// Client wraps an ssh session with streaming exec.
type Client struct {
	cfg  *ssh.ClientConfig
	addr string
}

// New connects to host:port as user with password.
func New(host, port, user, pass string) *Client {
	cfg := &ssh.ClientConfig{
		User:            user,
		Auth:            []ssh.AuthMethod{ssh.Password(pass)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), //nolint:gosec // installer tool, user controls server
		Timeout:         15 * time.Second,
	}
	return &Client{cfg: cfg, addr: net.JoinHostPort(host, port)}
}

// Test verifies the credentials work.
func (c *Client) Test(ctx context.Context) error {
	_ = ctx
	conn, err := ssh.Dial("tcp", c.addr, c.cfg)
	if err != nil {
		return fmt.Errorf("%w: %s", ErrConnect, err)
	}
	defer conn.Close()
	return nil
}

// Run executes cmd remotely; output lines are streamed to log.
// Returns combined trimmed output.
func (c *Client) Run(ctx context.Context, cmd string, log LogFunc) (string, error) {
	_ = ctx
	conn, err := ssh.Dial("tcp", c.addr, c.cfg)
	if err != nil {
		return "", fmt.Errorf("%w: %s", ErrConnect, err)
	}
	defer conn.Close()
	sess, err := conn.NewSession()
	if err != nil {
		return "", fmt.Errorf("ssh session: %w", err)
	}
	defer sess.Close()
	done := make(chan struct{})
	var buf strings.Builder
	pr, pw := io.Pipe()
	go streamLines(pr, log, &buf, done)
	sess.Stdout = pw
	sess.Stderr = pw
	if err := sess.Start(cmd); err != nil {
		_ = pw.Close()
		return "", fmt.Errorf("ssh start: %w", err)
	}
	errc := make(chan error, 1)
	go func() { errc <- sess.Wait(); _ = pw.Close() }()
	select {
	case <-ctx.Done():
		_ = sess.Signal(ssh.SIGKILL)
		return buf.String(), fmt.Errorf("ctx: %w", ctx.Err())
	case err := <-errc:
		<-done
		return buf.String(), err
	}
}

func streamLines(pr *io.PipeReader, log LogFunc, buf *strings.Builder, done chan<- struct{}) {
	defer close(done)
	const chunk = 8192
	r := make([]byte, chunk)
	var carry string
	for {
		n, err := pr.Read(r)
		if n > 0 {
			buf.Write(r[:n])
			carry = emitLines(carry+string(r[:n]), log)
		}
		if err != nil {
			if carry != "" && log != nil {
				log(carry)
			}
			return
		}
	}
}

func emitLines(s string, log LogFunc) string {
	for {
		i := strings.IndexByte(s, '\n')
		if i < 0 {
			return s
		}
		if log != nil {
			line := strings.TrimRight(s[:i], "\r")
			log(line)
		}
		s = s[i+1:]
	}
}

// Once runs a script slice serially, aborting on first error.
type Step struct {
	Name string
	Cmd  string
}

// Pipeline runs steps sequentially, logging each.
type Pipeline struct {
	c    *Client
	mu   sync.Mutex
	stop chan struct{}
}

// NewPipeline wraps client with cancel support.
func NewPipeline(c *Client) *Pipeline { return &Pipeline{c: c, stop: make(chan struct{})} }

// Cancel interrupts the running pipeline.
func (p *Pipeline) Cancel() {
	p.mu.Lock()
	defer p.mu.Unlock()
	select {
	case <-p.stop: // closed
	default:
		close(p.stop)
	}
}

// Exec runs steps with ctx, logging via log. Aborts on error.
func (p *Pipeline) Exec(ctx context.Context, steps []Step, log LogFunc) error {
	for _, st := range steps {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-p.stop:
			return fmt.Errorf("cancelled")
		default:
		}
		if log != nil {
			log("### " + st.Name)
		}
		if _, err := p.c.Run(ctx, st.Cmd, log); err != nil {
			return fmt.Errorf("step %s: %w", st.Name, err)
		}
		if log != nil {
			log("")
		}
	}
	return nil
}
