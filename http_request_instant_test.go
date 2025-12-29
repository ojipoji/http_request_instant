package http_request_instant

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
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
	client.SetDebug(true)
	resp, err := client.Request(context.TODO(), RequestOptions{
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
	client.SetDebug(false)
	resp, err := client.Request(context.TODO(), RequestOptions{
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
	resp, err := client.Request(context.TODO(), RequestOptions{
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

func TestTimeout(t *testing.T) {
	// bikin server dummy remote yang lambat
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(200)
	}))
	defer ts.Close()

	client := NewHttpRequest()
	// Set timeout sangat singkat
	client.Client.Timeout = 50 * time.Millisecond

	_, err := client.Request(context.Background(), RequestOptions{
		Method: "GET",
		URL:    ts.URL,
	})
	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}
}

func TestContextCancel(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(500 * time.Millisecond)
		w.WriteHeader(200)
	}))
	defer ts.Close()

	client := NewHttpRequest()
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_, err := client.Request(ctx, RequestOptions{
		Method: "GET",
		URL:    ts.URL,
	})
	if err == nil {
		t.Fatal("expected context deadline exceeded error, got nil")
	}
}
