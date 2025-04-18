package crawler

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// func TestFindLinks(t *testing.T) {
// 	htmlStr := `<html><body><a href="/about">About</a><a href="/blog">Blog</a></body></html>`
// 	doc, err := html.Parse(strings.NewReader(htmlStr))
// 	if err != nil {
// 		t.Fatalf("Failed to parse HTML: %v", err)
// 	}

// 	base, _ := url.Parse("http://127.0.0.1:56359")
// 	parent := &SiteMap{URL: base.String(), Depth: 0}
// 	visited := make(map[string]bool)
// 	visitedMu := sync.Mutex{}
// 	jobs := make(chan job, 10)

// 	newLinks := findLinks(doc, base, parent, visited, &visitedMu, 1, jobs)
// 	if newLinks != 2 {
// 		t.Errorf("Expected 2 new links, got %d", newLinks)
// 	}

// 	close(jobs)
// 	links := []string{}
// 	for j := range jobs {
// 		links = append(links, j.url)
// 	}

// 	expected := []string{
// 		"http://127.0.0.1:56359/about",
// 		"http://127.0.0.1:56359/blog",
// 	}
// 	for _, exp := range expected {
// 		found := false
// 		for _, link := range links {
// 			if link == exp {
// 				found = true
// 				break
// 			}
// 		}
// 		if !found {
// 			t.Errorf("Expected link %s not found in %v", exp, links)
// 		}
// 	}
// }

func TestCrawlWebsite(t *testing.T) {
	// Set up test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/robots.txt" {
			w.Write([]byte("User-agent: *\nAllow: /"))
			return
		}
		if r.URL.Path == "/" {
			html := `<html><body><a href="/about">About</a><a href="/blog">Blog</a></body></html>`
			t.Logf("Serving HTML for /: %s", html)
			w.Write([]byte(html))
		} else if r.URL.Path == "/about" {
			html := `<html><body><a href="/contact">Contact</a></body></html>`
			t.Logf("Serving HTML for /about: %s", html)
			w.Write([]byte(html))
		} else if r.URL.Path == "/blog" || r.URL.Path == "/contact" {
			html := `<html><body>No links here</body></html>`
			t.Logf("Serving HTML for %s: %s", r.URL.Path, html)
			w.Write([]byte(html))
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	tests := []struct {
		name         string
		baseURL      string
		maxDepth     int
		concurrency  int
		wantStatus   int
		wantErr      string
		wantChildren []string
	}{
		{
			name:         "Shallow crawl",
			baseURL:      server.URL,
			maxDepth:     1,
			concurrency:  1,
			wantStatus:   200,
			wantErr:      "",
			wantChildren: []string{server.URL + "/about", server.URL + "/blog"},
		},
		{
			name:         "Deep crawl",
			baseURL:      server.URL,
			maxDepth:     2,
			concurrency:  2,
			wantStatus:   200,
			wantErr:      "",
			wantChildren: []string{server.URL + "/about", server.URL + "/blog", server.URL + "/contact"},
		},
		{
			name:         "Zero depth",
			baseURL:      server.URL,
			maxDepth:     0,
			concurrency:  1,
			wantStatus:   200,
			wantErr:      "",
			wantChildren: []string{},
		},
		{
			name:         "Invalid URL",
			baseURL:      "http://nonexistent.invalid",
			maxDepth:     1,
			concurrency:  1,
			wantStatus:   0,
			wantErr:      "site http://nonexistent.invalid is unreachable",
			wantChildren: []string{},
		},
		{
			name:         "Negative depth",
			baseURL:      server.URL,
			maxDepth:     -1,
			concurrency:  1,
			wantStatus:   0,
			wantErr:      "invalid depth: -1",
			wantChildren: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, responseTime, siteMap, err := CrawlWebsite(tt.baseURL, tt.maxDepth, tt.concurrency)
			t.Logf("SiteMap: %v", dumpSiteMap(siteMap))

			if tt.wantErr != "" {
				if err == nil || err.Error() != tt.wantErr {
					t.Errorf("Expected error %q, got %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if status != tt.wantStatus {
				t.Errorf("Expected status %d, got %d", tt.wantStatus, status)
			}
			if responseTime < 0 {
				t.Errorf("Expected non-negative response time, got %d", responseTime)
			}
			if siteMap == nil {
				t.Fatal("SiteMap is nil")
			}
			if siteMap.URL != tt.baseURL && tt.wantErr == "" {
				t.Errorf("Expected root URL %s, got %s", tt.baseURL, siteMap.URL)
			}

			children := []string{}
			for _, child := range siteMap.Children {
				children = append(children, child.URL)
				for _, grandChild := range child.Children {
					children = append(children, grandChild.URL)
				}
			}
			for _, wantChild := range tt.wantChildren {
				found := false
				for _, child := range children {
					if child == wantChild {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected child %s not found in %v", wantChild, children)
				}
			}
		})
	}
}
