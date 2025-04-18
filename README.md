# website-crawler

A fast and flexible **command-line tool** written in **Go** that crawls a website, checks its status, builds a site map, and exports it as **JSON** or **XML**.

---

## 🚀 Features

- ✅ Crawl entire website up to a specified depth
- ✅ Respect `robots.txt` rules
- ✅ Concurrency support for faster crawling(ABSENT AT the moment it will be added in future.)
- ✅ Rate-limited by default to avoid server overload
- ✅ Export site map in JSON or XML format
- ✅ Human-readable tree output in terminal
- ✅ Built-in profiler support (`pprof`)
- ✅ Well-tested with full integration test suite

---
## 🛠️ Installation

> ⚠️ Make sure you have Go installed on your system. You can download it from https://golang.org/dl/

```bash
git clone https://github.com/yourusername/website-crawler.git
cd website-crawler
go build -o crawler
```

---

## 🔧 Usage

```bash
./crawler [URL] [flags]
```

If no URL is provided, it defaults to `https://example.com`.

---

## 📌 Flags

| Flag             | Description                                 | Default        |
|------------------|---------------------------------------------|----------------|
| `--output`, `-o` | Export site map to file                     | (none)         |
| `--format`, `-f` | Output format: `json` or `xml`              | `json`         |
| `--depth`, `-d`  | Max crawl depth                             | `10`           |
| `--concurrency`, `-c` | Number of concurrent crawlers         | `10`           |
| `--version`      | Show CLI version                            |                |

---

## 🧪 Examples

### ✅ Crawl and display site map
```bash
go run . https://example.com
```

### 📝 Export site map to JSON
```bash
go run . https://example.com --output sitemap.json --format json
```

### 🧾 Export site map to XML
```bash
go run . https://example.com --output sitemap.xml --format xml
```

### 🔄 Crawl with custom depth and concurrency
```bash
go run . https://example.com --depth 3 --concurrency 5
```

---

## ⚙️ Profiler

The tool includes optional `pprof` profiling on port `:6060`.
This runs **automatically** (except during testing):

```bash
go tool pprof http://localhost:6060/debug/pprof/profile
```

---

## 🧪 Testing

Run the full test suite:
```bash
go test ./...
```

The tests spin up a live local server and validate:
- Real crawling
- Exported file content
- Flag handling
- Robots.txt blocking
- Edge cases

---

## 📁 Example Output

```
[✓] Site: https://example.com

Status: 200 OK
Response Time: 123ms

Site Map:

- https://example.com
  - https://example.com/about
  - https://example.com/contact
```

---

---


