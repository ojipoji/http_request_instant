package http_request_instant

import (
	"encoding/json"
	"encoding/xml"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

// Struct contoh untuk JSON
type Post struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
}

// Struct contoh untuk XML
type Note struct {
	To   string `xml:"to"`
	From string `xml:"from"`
}

func TestJSONResponse(t *testing.T) {
	// bikin server dummy JSON
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(Post{ID: 1, Title: "Hello JSON"})
	}))
	defer ts.Close()

	var post Post
	client := NewHttpRequest()
	resp, err := client.Request(RequestOptions{
		Method:         "GET",
		URL:            ts.URL,
		ResponseTarget: &post,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if post.Title != "Hello JSON" {
		t.Errorf("expected Title=Hello JSON, got %s", post.Title)
	}
	if resp.StatusCode != 200 {
		t.Errorf("expected status=200, got %d", resp.StatusCode)
	}
}

func TestXMLResponse(t *testing.T) {
	// bikin server dummy XML
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/xml")
		_ = xml.NewEncoder(w).Encode(Note{To: "Alice", From: "Bob"})
	}))
	defer ts.Close()

	var note Note
	client := NewHttpRequest()
	resp, err := client.Request(RequestOptions{
		Method:         "GET",
		URL:            ts.URL,
		ResponseTarget: &note,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if note.To != "Alice" || note.From != "Bob" {
		t.Errorf("unexpected parse result: %+v", note)
	}
	if resp.StatusCode != 200 {
		t.Errorf("expected status=200, got %d", resp.StatusCode)
	}
}

func TestRawResponse(t *testing.T) {
	// bikin server dummy plain text
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		_, _ = w.Write([]byte("RAW DATA"))
	}))
	defer ts.Close()

	client := NewHttpRequest()
	resp, err := client.Request(RequestOptions{
		Method: "GET",
		URL:    ts.URL,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(resp.Body) != "RAW DATA" {
		t.Errorf("expected body=RAW DATA, got %s", string(resp.Body))
	}
}

func TestDevelopmentMode(t *testing.T) {
	// set env ke development
	_ = os.Setenv("IS_PRODUCTION", "false")

	client := NewHttpRequest()
	resp, err := client.Request(RequestOptions{
		Method: "GET",
		URL:    "http://example.com",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(resp.Body) != `{"mock":"success"}` {
		t.Errorf("expected mock response, got %s", string(resp.Body))
	}

	// reset env
	_ = os.Setenv("IS_PRODUCTION", "true")
}
