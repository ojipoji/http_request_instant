# http_request_instant

[![Go Reference](https://pkg.go.dev/badge/github.com/ojipoji/http_request_instant.svg)](https://pkg.go.dev/github.com/ojipoji/http_request_instant)
[![Go Report Card](https://goreportcard.com/badge/github.com/ojipoji/http_request_instant)](https://goreportcard.com/report/github.com/ojipoji/http_request_instant)
[![CI](https://github.com/ojipoji/http_request_instant/actions/workflows/go.yml/badge.svg)](https://github.com/ojipoji/http_request_instant/actions)
[![Release](https://img.shields.io/github/v/release/ojipoji/http_request_instant)](https://github.com/ojipoji/http_request_instant/releases)

Library sederhana untuk melakukan HTTP request di Go (GET, POST, PUT, DELETE) dengan support:
- JSON & XML
- Custom Header
- Basic Auth
- Mock response untuk development
- Flexible response (bisa langsung ke `[]byte` atau ke `struct`)

---

## 🚀 Instalasi

```bash
go get github.com/ojipoji/http_request_instant@latest
```


# Contoh Penggunaan

## GET Request

```go
package main

import (
	"fmt"
	"github.com/ojipoji/http_request_instant"
)

func main() {
	client := http_request_instant.NewHttpRequest()

	resp, err := client.Request(http_request_instant.RequestOptions{
		Method: "GET",
		URL:    "https://jsonplaceholder.typicode.com/posts/1",
	})
	if err != nil {
		panic(err)
	}

	fmt.Println("Status:", resp.StatusCode)
	fmt.Println("Body:", string(resp.Body))
}
```

## POST Request dengan JSON

```go
package main

import (
	"github.com/ojipoji/http_request_instant"
)

type Post struct {
	Title  string `json:"title"`
	Body   string `json:"body"`
	UserID int    `json:"userId"`
}

func main() {
	client := http_request_instant.NewHttpRequest()

	resp, err := client.Request(http_request_instant.RequestOptions{
		Method:      "POST",
		URL:         "https://jsonplaceholder.typicode.com/posts",
		ContentType: "application/json",
		RequestBody: Post{Title: "foo", Body: "bar", UserID: 1},
	})
	if err != nil {
		panic(err)
	}

	println("Status:", resp.StatusCode)
	println("Body:", string(resp.Body))
}
```

## POST dengan XML

```go
package main

import (
	"encoding/xml"
	"github.com/ojipoji/http_request_instant"
)

type User struct {
	XMLName xml.Name `xml:"user"`
	Name    string   `xml:"name"`
	Email   string   `xml:"email"`
}

func main() {
	client := http_request_instant.NewHttpRequest()

	resp, err := client.Request(http_request_instant.RequestOptions{
		Method:      "POST",
		URL:         "https://example.com/api",
		ContentType: "application/xml",
		RequestBody: User{Name: "Fauzi", Email: "fauzi@example.com"},
	})
	if err != nil {
		panic(err)
	}

	println("Status:", resp.StatusCode)
	println("Body:", string(resp.Body))
}
```

## Basic Auth

```go
package main

import (
	"github.com/ojipoji/http_request_instant"
)

func main() {
	client := http_request_instant.NewHttpRequest()

	resp, err := client.Request(http_request_instant.RequestOptions{
		Method: "GET",
		URL:    "https://api.example.com/secure-data",
		BasicAuth: &http_request_instant.BasicAuth{
			Username: "admin",
			Password: "password",
		},
	})
	if err != nil {
		panic(err)
	}

	println("Status:", resp.StatusCode)
	println("Body:", string(resp.Body))
}
```


## Flexible Response → Unmarshal langsung ke struct (opsional)

```go
package main

import (
	"fmt"
	"github.com/ojipoji/http_request_instant"
)

type ApiResult struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
	Body  string `json:"body"`
}

func main() {
	client := http_request_instant.NewHttpRequest()

	var result ApiResult
	resp, err := client.RequestWithParse(http_request_instant.RequestOptions{
		Method: "GET",
		URL:    "https://jsonplaceholder.typicode.com/posts/1",
	}, &result)

	if err != nil {
		panic(err)
	}

	fmt.Println("Status:", resp.StatusCode)
	fmt.Println("Parsed Struct:", result)
}
```

