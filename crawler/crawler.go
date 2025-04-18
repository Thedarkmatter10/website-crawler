package crawler

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"golang.org/x/net/html"
)

type SiteMap struct {
	URL      string     `json:"url"`
	Depth    int        `json:"depth"`
	Children []*SiteMap `json:"children"`
}

func CrawlWebsite(baseURL string, maxDepth, concurrency int) (int, int64, *SiteMap, error) {
	if maxDepth < 0 {
		return 0, 0, nil, fmt.Errorf("invalid depth: %d", maxDepth)
	}

	statusCode, responseTime := CheckWebsite(baseURL)
	if statusCode == 0 {
		return 0, 0, nil, fmt.Errorf("site %s is unreachable", baseURL)
	}

	visited := make(map[string]bool)
	visitedMu := sync.Mutex{}
	siteMap := &SiteMap{URL: baseURL, Depth: 0}

	crawlRecursive(baseURL, siteMap, visited, &visitedMu, maxDepth)

	fmt.Printf("Final site map: %v\n", dumpSiteMap(siteMap))
	return statusCode, responseTime, siteMap, nil
}

func crawlRecursive(currentURL string, node *SiteMap, visited map[string]bool, visitedMu *sync.Mutex, maxDepth int) {

	visitedMu.Lock()
	if visited[currentURL] {

		visitedMu.Unlock()
		return
	}
	visited[currentURL] = true
	visitedMu.Unlock()

	// Fetch URL
	resp, err := http.Get(currentURL)
	if err != nil {
		fmt.Printf("Error fetching %s: %v\n", currentURL, err)
		return
	}

	// Read and log HTML
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading body %s: %v\n", currentURL, err)
		resp.Body.Close()
		return
	}

	resp.Body.Close()

	// Parse HTML
	doc, err := html.Parse(strings.NewReader(string(body)))
	if err != nil {
		fmt.Printf("Error parsing %s: %v\n", currentURL, err)
		return
	}
	fmt.Printf("Parsed HTML successfully for %s\n", currentURL)

	// Find links if depth allows
	if node.Depth < maxDepth {
		base, err := url.Parse(currentURL)
		if err != nil {
			fmt.Printf("Error parsing URL %s: %v\n", currentURL, err)
			return
		}
		links := findLinks(doc, base, node, visited, visitedMu, maxDepth)

		// Recursively crawl each link
		for _, link := range links {
			childNode := &SiteMap{URL: link, Depth: node.Depth + 1}
			fmt.Printf("Adding child %s to parent %s\n", link, node.URL)
			node.Children = append(node.Children, childNode)
			crawlRecursive(link, childNode, visited, visitedMu, maxDepth)
		}
	} else {
		fmt.Printf("Skipping findLinks for %s: depth %d >= maxDepth %d\n", currentURL, node.Depth, maxDepth)
	}
}

// Debug function to dump site map
func dumpSiteMap(s *SiteMap) string {
	if s == nil {
		return "nil"
	}
	var result strings.Builder
	result.WriteString(fmt.Sprintf("{URL: %s, Depth: %d, Children: [", s.URL, s.Depth))
	for i, child := range s.Children {
		if i > 0 {
			result.WriteString(", ")
		}
		result.WriteString(dumpSiteMap(child))
	}
	result.WriteString("]}")
	return result.String()
}

func findLinks(n *html.Node, base *url.URL, parent *SiteMap, visited map[string]bool, visitedMu *sync.Mutex, maxDepth int) []string {
	var links []string
	if n == nil {
		return links
	}

	if n.Type == html.ElementNode && strings.ToLower(n.Data) == "a" {
		for _, attr := range n.Attr {
			if strings.ToLower(attr.Key) == "href" {

				link, err := base.Parse(attr.Val)
				if err != nil {
					fmt.Printf("Error parsing link %s: %v\n", attr.Val, err)
					continue
				}
				linkStr := link.String()

				visitedMu.Lock()
				if IsSameDomain(linkStr, base.String()) && !visited[linkStr] && parent.Depth < maxDepth {
					fmt.Printf("Collecting link: %s, depth: %d\n", linkStr, parent.Depth+1)
					links = append(links, linkStr)
				}
				visitedMu.Unlock()
			}
		}
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		links = append(links, findLinks(c, base, parent, visited, visitedMu, maxDepth)...)
	}
	return links
}

func IsSameDomain(url1, url2 string) bool {
	u1, err1 := url.Parse(url1)
	u2, err2 := url.Parse(url2)
	if err1 != nil || err2 != nil {
		return false
	}
	return u1.Host == u2.Host
}
