package cmd

import (
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"strings"
	"time"

	"website-crawler/crawler"

	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:   "website-crawler [URL] [flags]",
	Short: "A CLI tool to crawl websites and generate a site map",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		url := "https://example.com"
		if len(args) > 0 {
			url = args[0]
		}

		outputFile, _ := cmd.Flags().GetString("output")
		outputFormat, _ := cmd.Flags().GetString("format")
		maxDepth, err := cmd.Flags().GetInt("depth")
		if err != nil {
			return fmt.Errorf("invalid depth: %v", err)
		}
		concurrency, _ := cmd.Flags().GetInt("concurrency")

		if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
			url = "https://" + url
		}

		if !crawler.IsAllowed(url, "WebsiteCrawler") && os.Getenv("TESTING") != "true" {
			return fmt.Errorf("crawling %s is not allowed by robots.txt", url)
		}

		if os.Getenv("TESTING") != "true" {
			go func() {
				if err := http.ListenAndServe(":6060", nil); err != nil && err != http.ErrServerClosed {
					fmt.Fprintf(cmd.ErrOrStderr(), "pprof server error: %v\n", err)
				}
			}()
		}

		startTime := time.Now()
		statusCode, responseTime, siteMap, err := crawler.CrawlWebsite(url, maxDepth, concurrency)
		fmt.Fprintf(cmd.OutOrStdout(), "Crawling took %v\n", time.Since(startTime))
		if err != nil {
			return err
		}

		fmt.Fprintf(cmd.OutOrStdout(), "[✓] Site: %s\n\n", url)
		fmt.Fprintf(cmd.OutOrStdout(), "Status: %d OK\n", statusCode)
		fmt.Fprintf(cmd.OutOrStdout(), "Response Time: %dms\n\n", responseTime)
		fmt.Fprintf(cmd.OutOrStdout(), "Site Map:\n\n")
		crawler.PrintSiteMapTo(siteMap, cmd.OutOrStdout())

		if outputFile != "" {
			if err := crawler.ExportSiteMap(siteMap, outputFile, outputFormat); err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "Error exporting site map: %v\n", err)
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Site map exported to %s\n", outputFile)
		}

		return nil
	},
}

func init() {
	RootCmd.Flags().StringP("output", "o", "", "Export site map to file")
	RootCmd.Flags().StringP("format", "f", "json", "Output format (json or xml)")
	RootCmd.Flags().IntP("depth", "d", 10, "Maximum crawl depth")
	RootCmd.Flags().IntP("concurrency", "c", 10, "Number of concurrent crawlers")
	RootCmd.Version = "1.0.0"
}

func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
