# http_request_instant

Sebuah library sederhana untuk melakukan HTTP request di Go dengan dukungan:
- **Custom method** (GET, POST, PUT, DELETE, dll.)
- **Headers** custom
- **Basic Authentication**
- **JSON/XML request body**
- **Environment-based response mock** (`IS_PRODUCTION=false` → mock response)

---

## Instalasi

```bash
go get github.com/username/http_request_instant
