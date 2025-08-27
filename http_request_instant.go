package http_request_instant

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
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

// RequestWithParse mempermudah request + langsung parsing response ke struct target
func (h *HttpRequest) RequestWithParse(opts RequestOptions, target interface{}) (*ApiResponse, error) {
	resp, err := h.Request(opts)
	if err != nil {
		return nil, err
	}

	if opts.ContentType == "application/xml" {
		if err := xml.Unmarshal(resp.Body, target); err != nil {
			return resp, fmt.Errorf("failed to unmarshal XML: %w", err)
		}
	} else {
		// default JSON
		if err := json.Unmarshal(resp.Body, target); err != nil {
			return resp, fmt.Errorf("failed to unmarshal JSON: %w", err)
		}
	}

	return resp, nil
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
		req, err = http.NewRequest(options.Method, options.URL, bytes.NewBuffer(body))
	} else {
		req, err = http.NewRequest(options.Method, options.URL, nil)
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

	// Simpan response headers ke map
	headers := make(map[string]string)
	for k, v := range resp.Header {
		if len(v) > 0 {
			headers[k] = v[0]
		}
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

// handleDevelopmentRequest mengembalikan response mock
// jika IS_PRODUCTION=false.
func (c *HttpRequest) handleDevelopmentRequest(options RequestOptions) (*ApiResponse, error) {
	respByte := []byte(`{"mock":"success"}`)
	return &ApiResponse{
		StatusCode: 200,
		Body:       respByte,
		Headers:    map[string]string{"Content-Type": "application/json"},
	}, nil
}
