package http_request_instant

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"os"
)

// RequestOptions menyimpan konfigurasi request HTTP.
type RequestOptions struct {
	Method      string            // HTTP method (GET, POST, PUT, DELETE, dll.)
	URL         string            // Target URL
	Headers     map[string]string // Custom headers
	RequestBody interface{}       // Body request (bisa map, struct, string, []byte)
	ContentType string            // application/json, application/xml, dll.
	*BasicAuth                    // Optional: Basic Auth
}

// BasicAuth menyimpan informasi autentikasi Basic.
type BasicAuth struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// ApiResponse merepresentasikan response dari server.
type ApiResponse struct {
	StatusCode int    // HTTP status code
	Body       []byte // Response body
}

// HttpRequestInf mendefinisikan interface untuk request HTTP.
type HttpRequestInf interface {
	Request(options RequestOptions) (*ApiResponse, error)
}

// HttpRequest adalah implementasi HttpRequestInf
// yang menggunakan http.Client bawaan Go.
type HttpRequest struct {
	cl *http.Client
}

// NewHttpRequest membuat instance baru HttpRequest.
func NewHttpRequest() *HttpRequest {
	return &HttpRequest{
		cl: &http.Client{},
	}
}

// Request mengeksekusi HTTP request berdasarkan RequestOptions.
// Jika IS_PRODUCTION=false, maka akan menggunakan mock response.
func (c *HttpRequest) Request(options RequestOptions) (*ApiResponse, error) {

	// Mode development: return mock response
	if os.Getenv("IS_PRODUCTION") == "false" {
		return c.handleDevelopmentRequest(options)
	}

	var req *http.Request
	var err error

	var body []byte
	if options.RequestBody != nil {
		// Jika RequestBody sudah string atau []byte → langsung dipakai
		switch v := options.RequestBody.(type) {
		case string:
			body = []byte(v)
		case []byte:
			body = v
		default:
			// Jika bukan string/[]byte → marshal sesuai ContentType
			switch options.ContentType {
			case "application/json":
				body, err = json.Marshal(v)
			case "application/xml":
				body, err = xml.Marshal(v)
			default:
				return nil, fmt.Errorf("unsupported Content-Type: %s", options.ContentType)
			}
			if err != nil {
				return nil, fmt.Errorf("error marshal request body: %w", err)
			}
		}
		req, err = http.NewRequest(options.Method, options.URL, bytes.NewBuffer(body))
	} else {
		req, err = http.NewRequest(options.Method, options.URL, nil)
	}

	if err != nil {
		return nil, fmt.Errorf("error create request: %w", err)
	}

	// Set Content-Type jika ada
	if options.ContentType != "" {
		req.Header.Set("Content-Type", options.ContentType)
	}

	// Set custom headers
	for key, value := range options.Headers {
		req.Header.Set(key, value)
	}

	// Set Basic Auth jika diisi
	if options.BasicAuth != nil {
		req.SetBasicAuth(options.BasicAuth.Username, options.BasicAuth.Password)
	}

	// Eksekusi request
	resp, err := c.cl.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Baca response body
	respByte, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return &ApiResponse{
		StatusCode: resp.StatusCode,
		Body:       respByte,
	}, nil
}

// handleDevelopmentRequest mengembalikan response mock
// jika IS_PRODUCTION=false.
func (c *HttpRequest) handleDevelopmentRequest(options RequestOptions) (*ApiResponse, error) {
	respByte := []byte("success")
	return &ApiResponse{
		StatusCode: 200,
		Body:       respByte,
	}, nil
}
