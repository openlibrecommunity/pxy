package pxyapp

import (
	"bytes"
	"mime/multipart"
	"net/http/httptest"
	"testing"
)

func TestValidName(t *testing.T) {
	t.Parallel()
	if !validName("x.olc8.cc") {
		t.Fatal("valid name rejected")
	}
	if validName("bad..olc8.cc") {
		t.Fatal("invalid name accepted")
	}
}

func TestFindDomain(t *testing.T) {
	t.Parallel()
	domains := []domainConfig{{Name: "olc8.cc", Token: "x"}}
	if _, err := findDomain("a.olc8.cc", domains); err != nil {
		t.Fatalf("find domain: %v", err)
	}
	if _, err := findDomain("olc8.cc", domains); err == nil {
		t.Fatal("root domain accepted")
	}
}

func TestPow(t *testing.T) {
	t.Parallel()
	if !validPow("n", "2.2.2.2", "x.olc8.cc", "1", 0) {
		t.Fatal("zero difficulty pow rejected")
	}
	if validPow("n", "2.2.2.2", "x.olc8.cc", "bad", 0) {
		t.Fatal("bad solution accepted")
	}
}

func TestMultipartForm(t *testing.T) {
	t.Parallel()
	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	if err := w.WriteField("ip", "23.23.23.23"); err != nil {
		t.Fatal(err)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest("POST", "/create", &body)
	req.Header.Set("content-type", w.FormDataContentType())
	if err := req.ParseMultipartForm(maxBody); err != nil {
		t.Fatal(err)
	}
	if got := req.Form.Get("ip"); got != "23.23.23.23" {
		t.Fatalf("ip = %q", got)
	}
}
