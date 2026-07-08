package installer

import (
	"strings"
	"testing"
)

func TestDefaultsObfs(t *testing.T) {
	t.Parallel()
	r := Request{}
	r.Defaults()
	if r.Hy2Obfs != "salamander" {
		t.Fatalf("obfs default = %q", r.Hy2Obfs)
	}
	r = Request{Hy2Obfs: "bad"}
	r.Defaults()
	if r.Hy2Obfs != "salamander" {
		t.Fatalf("bad obfs not reset: %q", r.Hy2Obfs)
	}
	r = Request{Hy2Obfs: "gecko"}
	r.Defaults()
	if r.Hy2Obfs != "gecko" {
		t.Fatalf("gecko dropped: %q", r.Hy2Obfs)
	}
}

func TestScriptHy2Obfs(t *testing.T) {
	t.Parallel()
	s := Script(Request{Domain: "a.b.c", Hy2Obfs: "gecko", Protocols: Protocols{Hysteria2: true}})
	if !strings.Contains(s, "HY2_OBFS='gecko'") {
		t.Fatal("obfs not passed to script")
	}
	if !strings.Contains(s, "type: $HY2_OBFS") {
		t.Fatal("obfs type not templated")
	}
}

func TestScriptUFW(t *testing.T) {
	t.Parallel()
	s := Script(Request{Domain: "a.b.c", Protocols: Protocols{VLESS: true, AmneziaWG: true}})
	for _, want := range []string{
		"apt-get install -y ufw",
		"ufw allow '22'/tcp",
		"ufw --force enable",
		"DEFAULT_FORWARD_POLICY",
		"iptables",
	} {
		if !strings.Contains(s, want) {
			t.Fatalf("ufw block missing %q", want)
		}
	}
}
