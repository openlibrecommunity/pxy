package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	defaultAddr = "127.0.0.1:8080"
	defaultPow  = 20
	maxBody     = 1 << 20
	cfBaseURL   = "https://api.cloudflare.com/client/v4"
)

var (
	ErrBadRequest = errors.New("bad request")
	ErrNotFound   = errors.New("not found")
)

type app struct {
	client     *http.Client
	log        *slog.Logger
	challenges *challengeStore
	domains    []domainConfig
}

type domainConfig struct {
	Name  string
	Token string
}

type pageData struct {
	Domains []domainConfig
	Bits    int
}

type challenge struct {
	IP     string
	FQDN   string
	Nonce  string
	Expiry time.Time
}

type challengeStore struct {
	mu    sync.Mutex
	items map[string]challenge
}

type challengeResp struct {
	Challenge string `json:"challenge"`
	Bits      int    `json:"bits"`
}

type cfListResp struct {
	Success bool `json:"success"`
	Result  []struct {
		ID string `json:"id"`
	} `json:"result"`
	Errors []cfError `json:"errors"`
}

type cfCreateResp struct {
	Success bool      `json:"success"`
	Errors  []cfError `json:"errors"`
}

type cfError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func main() {
	if err := run(); err != nil {
		slog.Error("fatal", "err", err)
		os.Exit(1)
	}
}

func run() error {
	domains, err := loadDomains(".env", ".domains")
	if err != nil {
		return err
	}
	a := &app{
		client:     &http.Client{Timeout: 20 * time.Second},
		log:        slog.Default(),
		challenges: newChallengeStore(),
		domains:    domains,
	}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /", a.handleIndex)
	mux.HandleFunc("GET /challenge", a.handleChallenge)
	mux.HandleFunc("POST /create", a.handleCreate)
	addr := envOr("ADDR", defaultAddr)
	srv := &http.Server{Addr: addr, Handler: mux, ReadHeaderTimeout: 5 * time.Second}
	slog.Info("listen", "addr", addr)
	return srv.ListenAndServe()
}

func newChallengeStore() *challengeStore {
	return &challengeStore{items: make(map[string]challenge)}
}

func (s *challengeStore) put(c challenge) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items[c.Nonce] = c
}

func (s *challengeStore) take(nonce string) (challenge, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	c, ok := s.items[nonce]
	delete(s.items, nonce)
	return c, ok
}

func (a *app) handleIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "text/html; charset=utf-8")
	if err := indexTmpl.Execute(w, pageData{Domains: a.domains, Bits: powBits()}); err != nil {
		a.log.Error("render", "err", err)
	}
}

func (a *app) handleChallenge(w http.ResponseWriter, r *http.Request) {
	ip := strings.TrimSpace(r.URL.Query().Get("ip"))
	fqdn := strings.TrimSpace(r.URL.Query().Get("fqdn"))
	if err := validateCreateInput(ip, fqdn, a.domains); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	nonce, err := newNonce()
	if err != nil {
		http.Error(w, "nonce failed", http.StatusInternalServerError)
		return
	}
	a.challenges.put(challenge{IP: ip, FQDN: fqdn, Nonce: nonce, Expiry: time.Now().Add(10 * time.Minute)})
	writeJSON(w, challengeResp{Challenge: nonce, Bits: powBits()})
}

func (a *app) handleCreate(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(maxBody); err != nil {
		http.Error(w, ErrBadRequest.Error(), http.StatusBadRequest)
		return
	}
	ip, fqdn := strings.TrimSpace(r.Form.Get("ip")), strings.TrimSpace(r.Form.Get("fqdn"))
	nonce, solution := r.Form.Get("challenge"), r.Form.Get("solution")
	if err := a.verify(ip, fqdn, nonce, solution); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	domain, err := findDomain(fqdn, a.domains)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := a.createRecord(r.Context(), domain, fqdn, ip); err != nil {
		a.log.Error("create record", "err", err, "fqdn", fqdn)
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	w.Header().Set("content-type", "text/plain; charset=utf-8")
	_, _ = fmt.Fprintf(w, "created: %s -> %s\n", fqdn, ip)
}

func (a *app) verify(ip, fqdn, nonce, solution string) error {
	if err := validateCreateInput(ip, fqdn, a.domains); err != nil {
		return err
	}
	c, ok := a.challenges.take(nonce)
	if !ok || time.Now().After(c.Expiry) {
		return fmt.Errorf("%w: challenge expired", ErrBadRequest)
	}
	if c.IP != ip || c.FQDN != fqdn || !validPow(nonce, ip, fqdn, solution, powBits()) {
		return fmt.Errorf("%w: pow failed", ErrBadRequest)
	}
	return nil
}

func powBits() int {
	val, err := strconv.Atoi(os.Getenv("POW_BITS"))
	if err != nil || val < 0 || val > 32 {
		return defaultPow
	}
	return val
}

func (a *app) createRecord(ctx context.Context, domain domainConfig, fqdn, ip string) error {
	zoneID, err := a.zoneID(ctx, domain)
	if err != nil {
		return err
	}
	body := map[string]any{"type": "A", "name": fqdn, "content": ip, "ttl": 1, "proxied": false}
	var out cfCreateResp
	if err := a.cf(ctx, http.MethodPost, "/zones/"+zoneID+"/dns_records", domain.Token, body, &out); err != nil {
		return err
	}
	if !out.Success {
		return fmt.Errorf("cloudflare: %s", cfErrors(out.Errors))
	}
	return nil
}

func (a *app) zoneID(ctx context.Context, domain domainConfig) (string, error) {
	var out cfListResp
	path := "/zones?name=" + domain.Name
	if err := a.cf(ctx, http.MethodGet, path, domain.Token, nil, &out); err != nil {
		return "", err
	}
	if !out.Success {
		return "", fmt.Errorf("cloudflare: %s", cfErrors(out.Errors))
	}
	if len(out.Result) == 0 || out.Result[0].ID == "" {
		return "", fmt.Errorf("%w: zone %s", ErrNotFound, domain.Name)
	}
	return out.Result[0].ID, nil
}

func (a *app) cf(ctx context.Context, method, path, token string, body any, out any) error {
	buf, err := encodeBody(body)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, method, cfBaseURL+path, buf)
	if err != nil {
		return err
	}
	req.Header.Set("authorization", "Bearer "+token)
	req.Header.Set("content-type", "application/json")
	resp, err := a.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(io.LimitReader(resp.Body, maxBody))
	if err != nil {
		return err
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return fmt.Errorf("cloudflare http %d: %s", resp.StatusCode, string(data))
	}
	return json.Unmarshal(data, out)
}

func encodeBody(body any) (io.Reader, error) {
	if body == nil {
		return nil, nil
	}
	data, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(data), nil
}

func loadDomains(envPath, domainsPath string) ([]domainConfig, error) {
	secrets, err := loadKV(envPath)
	if err != nil {
		return nil, err
	}
	zoneKeys, err := loadKV(domainsPath)
	if err != nil {
		return nil, err
	}
	domains := make([]domainConfig, 0, len(zoneKeys))
	for key, name := range zoneKeys {
		token := secrets[key]
		if token == "" {
			return nil, fmt.Errorf("missing token for %s", key)
		}
		domains = append(domains, domainConfig{Name: name, Token: token})
	}
	sort.Slice(domains, func(i, j int) bool { return domains[i].Name < domains[j].Name })
	return domains, nil
}

func loadKV(path string) (map[string]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	items := make(map[string]string)
	for _, line := range strings.Split(string(data), "\n") {
		key, val, ok := strings.Cut(strings.TrimSpace(line), "=")
		if !ok || key == "" || strings.HasPrefix(key, "#") {
			continue
		}
		items[key] = val
	}
	return items, nil
}

func validateCreateInput(ip, fqdn string, domains []domainConfig) error {
	if net.ParseIP(ip) == nil || strings.Contains(ip, ":") {
		return fmt.Errorf("%w: invalid ipv4", ErrBadRequest)
	}
	if !validName(fqdn) {
		return fmt.Errorf("%w: invalid domain", ErrBadRequest)
	}
	_, err := findDomain(fqdn, domains)
	return err
}

func findDomain(fqdn string, domains []domainConfig) (domainConfig, error) {
	for _, domain := range domains {
		if strings.HasSuffix(fqdn, "."+domain.Name) && fqdn != domain.Name {
			return domain, nil
		}
	}
	return domainConfig{}, fmt.Errorf("%w: domain not allowed", ErrBadRequest)
}

func validName(name string) bool {
	if len(name) < 3 || len(name) > 253 || strings.Contains(name, "..") {
		return false
	}
	for _, part := range strings.Split(name, ".") {
		if !validLabel(part) {
			return false
		}
	}
	return true
}

func validLabel(label string) bool {
	if len(label) == 0 || len(label) > 63 || label[0] == '-' || label[len(label)-1] == '-' {
		return false
	}
	for _, r := range label {
		if r >= 'a' && r <= 'z' || r >= '0' && r <= '9' || r == '-' {
			continue
		}
		return false
	}
	return true
}

func newNonce() (string, error) {
	buf := make([]byte, 24)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

func validPow(nonce, ip, fqdn, solution string, bits int) bool {
	if _, err := strconv.ParseUint(solution, 10, 64); err != nil {
		return false
	}
	sum := sha256.Sum256([]byte(nonce + ":" + ip + ":" + fqdn + ":" + solution))
	return leadingZeroBits(sum[:]) >= bits
}

func leadingZeroBits(data []byte) int {
	total := 0
	for _, b := range data {
		if b == 0 {
			total += 8
			continue
		}
		for mask := byte(0x80); mask > 0 && b&mask == 0; mask >>= 1 {
			total++
		}
		return total
	}
	return total
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("content-type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}

func cfErrors(items []cfError) string {
	parts := make([]string, 0, len(items))
	for _, item := range items {
		parts = append(parts, item.Message)
	}
	return strings.Join(parts, "; ")
}

func envOr(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
