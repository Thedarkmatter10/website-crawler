package cmd_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"website-crawler/cmd"
)

type siteMapNode struct {
	URL      string         `json:"url"`
	Children []*siteMapNode `json:"children"`
}

func TestRootCmd_FullIntegration(t *testing.T) {
	_ = os.Setenv("TESTING", "true") // bypass robots.txt check

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/":
			w.Write([]byte(`<a href="/about">About</a><a href="/contact">Contact</a>`))
		case "/about":
			w.Write([]byte(`<a href="/">Home</a>`))
		case "/contact":
			w.Write([]byte(`Contact page`))
		case "/robots.txt":
			w.Write([]byte(`User-agent: *\nDisallow:`)) // allow all
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	t.Run("Export JSON and Validate Content", func(t *testing.T) {
		outputFile := "test_output.json"
		defer os.Remove(outputFile) // clean up

		var buf bytes.Buffer
		cmd.RootCmd.SetOut(&buf)
		cmd.RootCmd.SetErr(&buf)
		cmd.RootCmd.SetArgs([]string{
			server.URL, "--output", outputFile, "--format", "json", "--depth", "2", "--concurrency", "2",
		})

		if err := cmd.RootCmd.Execute(); err != nil {
			t.Fatalf("Execution failed: %v", err)
		}

		// Check file exists
		data, err := os.ReadFile(outputFile)
		if err != nil {
			t.Fatalf("Expected output file %s not found", outputFile)
		}

		// Parse JSON
		var root siteMapNode
		if err := json.Unmarshal(data, &root); err != nil {
			t.Fatalf("Invalid JSON structure: %v", err)
		}

		if root.URL != server.URL {
			t.Errorf("Expected root URL %q, got %q", server.URL, root.URL)
		}

		found := map[string]bool{}
		for _, child := range root.Children {
			found[child.URL] = true
		}

		if !found[server.URL+"/about"] || !found[server.URL+"/contact"] {
			t.Errorf("Expected /about and /contact in children, got: %+v", root.Children)
		}
	})

	t.Run("Invalid output format", func(t *testing.T) {
		var buf bytes.Buffer
		cmd.RootCmd.SetOut(&buf)
		cmd.RootCmd.SetErr(&buf)
		cmd.RootCmd.SetArgs([]string{
			server.URL, "--output", "badfile.out", "--format", "badformat",
		})

		err := cmd.RootCmd.Execute()
		if err == nil || !strings.Contains(err.Error(), "unsupported format") {
			t.Errorf("Expected unsupported format error, got: %v", err)
		}
	})

	t.Run("Negative depth should fail", func(t *testing.T) {
		var buf bytes.Buffer
		cmd.RootCmd.SetOut(&buf)
		cmd.RootCmd.SetErr(&buf)
		cmd.RootCmd.SetArgs([]string{
			server.URL, "--depth", "-1",
		})

		err := cmd.RootCmd.Execute()
		if err == nil || !strings.Contains(buf.String(), "invalid") {
			t.Errorf("Expected error on invalid depth, got: %v", err)
		}
	})

	t.Run("Robots.txt blocks crawl", func(t *testing.T) {
		_ = os.Unsetenv("TESTING") // Now robots.txt will be respected

		blocked := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/robots.txt" {
				w.Write([]byte("User-agent: *\nDisallow: /"))
				return
			}
			w.Write([]byte(`blocked`))
		}))
		defer blocked.Close()

		var buf bytes.Buffer
		cmd.RootCmd.SetOut(&buf)
		cmd.RootCmd.SetErr(&buf)
		cmd.RootCmd.SetArgs([]string{blocked.URL})

		err := cmd.RootCmd.Execute()
		if err == nil || !strings.Contains(err.Error(), "not allowed by robots.txt") {
			t.Errorf("Expected robots.txt error, got: %v", err)
		}
	})
}
