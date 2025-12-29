package http_request_instant

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// RequestOptions menyimpan konfigurasi request HTTP.
type RequestOptions struct {
	Method         string            // HTTP method (GET, POST, PUT, DELETE, dll.)
	URL            string            // Target URL
	Headers        map[string]string // Custom headers
	RequestBody    interface{}       // Body request (bisa map, struct, string, []byte)
	ContentType    string            // Content-Type request (application/json, application/xml, dll.)
	ResponseTarget interface{}       // Optional: jika diisi, response akan di-unmarshal ke struct
	*BasicAuth
}

// BasicAuth menyimpan informasi autentikasi Basic.
type BasicAuth struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// ApiResponse merepresentasikan response dari server.
type ApiResponse struct {
	StatusCode int               // HTTP status code
	Body       []byte            // Response body dalam bentuk raw
	Headers    map[string]string // Response headers
}

// HttpRequestInf mendefinisikan interface untuk request HTTP.
type HttpRequestInf interface {
	Request(ctx context.Context, options RequestOptions) (*ApiResponse, error)
}

// HttpRequest adalah implementasi HttpRequestInf
// yang menggunakan http.Client bawaan Go.
type HttpRequest struct {
	Client *http.Client

	// debug request and response
	Debug bool
}

// NewHttpRequest membuat instance baru HttpRequest dengan default timeout 30 detik.
func NewHttpRequest() *HttpRequest {
	return &HttpRequest{
		Client: &http.Client{
			Timeout: 30 * time.Second,
		},
		Debug: false,
	}
}

// SetDebug mengaktifkan atau menonaktifkan mode debug.
func (h *HttpRequest) SetDebug(debug bool) {
	h.Debug = debug
}

// Request mengeksekusi HTTP request berdasarkan RequestOptions.
func (c *HttpRequest) Request(ctx context.Context, options RequestOptions) (*ApiResponse, error) {
	var req *http.Request
	var err error

	var body []byte
	if options.RequestBody != nil {
		switch v := options.RequestBody.(type) {
		case string:
			body = []byte(v)
		case []byte:
			body = v
		default:
			switch options.ContentType {
			case "application/json", "":
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
		req, err = http.NewRequestWithContext(ctx, options.Method, options.URL, bytes.NewBuffer(body))
	} else {
		req, err = http.NewRequestWithContext(ctx, options.Method, options.URL, nil)
	}

	if err != nil {
		return nil, fmt.Errorf("error create request: %w", err)
	}

	// Set Content-Type untuk request jika ada
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

	if c.Debug {
		fmt.Println("=== [HTTP REQUEST] ===")
		fmt.Printf("URL: %s\n", req.URL.String())
		fmt.Printf("Method: %s\n", req.Method)
		fmt.Println("Headers:")
		for k, v := range req.Header {
			fmt.Printf("  %s: %s\n", k, strings.Join(v, ", "))
		}
		if body != nil {
			fmt.Printf("Body: %s\n", string(body))
		}
		fmt.Println("======================")
	}

	// Eksekusi request
	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Baca response body
	respByte, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Simpan response headers ke map
	headers := make(map[string]string)
	for k, v := range resp.Header {
		if len(v) > 0 {
			headers[k] = v[0]
		}
	}

	// Debug: print response details
	if c.Debug {
		fmt.Println("=== [HTTP RESPONSE] ===")
		fmt.Printf("Status Code: %d\n", resp.StatusCode)
		fmt.Println("Headers:")
		for k, v := range resp.Header {
			fmt.Printf("  %s: %s\n", k, strings.Join(v, ", "))
		}
		fmt.Printf("Body: %s\n", string(respByte))
		fmt.Println("=======================")
	}

	// Jika ada ResponseTarget, unmarshal otomatis
	if options.ResponseTarget != nil {
		contentType := options.ContentType
		if contentType == "" {
			contentType = resp.Header.Get("Content-Type")
		}

		switch {
		case strings.Contains(contentType, "application/json"), contentType == "":
			if err := json.Unmarshal(respByte, options.ResponseTarget); err != nil {
				return nil, fmt.Errorf("failed to unmarshal JSON response: %w", err)
			}
		case strings.Contains(contentType, "application/xml"):
			if err := xml.Unmarshal(respByte, options.ResponseTarget); err != nil {
				return nil, fmt.Errorf("failed to unmarshal XML response: %w", err)
			}
		default:
			// fallback JSON
			if err := json.Unmarshal(respByte, options.ResponseTarget); err != nil {
				return nil, fmt.Errorf("unsupported Content-Type (%s) and failed JSON fallback: %w", contentType, err)
			}
		}
	}

	return &ApiResponse{
		StatusCode: resp.StatusCode,
		Body:       respByte,
		Headers:    headers,
	}, nil
}
