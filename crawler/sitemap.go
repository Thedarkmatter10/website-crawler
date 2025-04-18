package crawler

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"os"
)

func PrintSiteMap(siteMap *SiteMap) {
	if siteMap == nil {
		return
	}
	printNode(siteMap, "", os.Stdout)
}

func PrintSiteMapTo(siteMap *SiteMap, w io.Writer) {
	if siteMap == nil {
		return
	}
	printNode(siteMap, "", w)
}

func printNode(node *SiteMap, prefix string, w io.Writer) {
	fmt.Fprintf(w, "%s- %s\n", prefix, node.URL)
	for _, child := range node.Children {
		printNode(child, prefix+"  ", w)
	}
}

func ExportSiteMap(siteMap *SiteMap, filename, format string) error {
	if siteMap == nil {
		return fmt.Errorf("site map is nil")
	}
	if format != "json" && format != "xml" {
		return fmt.Errorf("unsupported format: %s", format)
	}
	var data []byte
	var err error
	if format == "json" {
		data, err = json.MarshalIndent(siteMap, "", "  ")
	} else {
		data, err = xml.MarshalIndent(siteMap, "", "  ")
	}
	if err != nil {
		return err
	}
	return os.WriteFile(filename, data, 0644)
}
