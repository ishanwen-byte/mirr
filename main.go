package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/atotto/clipboard"
)

const (
	githubURLPattern = `^https?://(?:www\.)?github\.com/([a-zA-Z0-9_.-]+)/([a-zA-Z0-9_.-]+)(?:\.git)?/?$`
	mirrorURLPrefix  = "https://ghfast.top/https://github.com/"
)

func main() {
	var (
		urlFlag  = flag.String("url", "", "GitHub URL to convert")
		helpFlag = flag.Bool("help", false, "Show help information")
	)
	flag.Parse()

	if *helpFlag {
		showHelp()
		return
	}

	var urls []string

	if *urlFlag != "" {
		urls = append(urls, *urlFlag)
	} else if flag.NArg() > 0 {
		urls = append(urls, flag.Args()...)
	} else {
		// Read from stdin
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line != "" {
				urls = append(urls, line)
			}
		}
		if err := scanner.Err(); err != nil {
			log.Fatalf("Error reading from stdin: %v", err)
		}
	}

	if len(urls) == 0 {
		fmt.Println("Error: No URLs provided. Use -help for usage information.")
		os.Exit(1)
	}

	for _, url := range urls {
		mirrorURL := convertToMirror(url)
		if mirrorURL != "" {
			fmt.Printf("Original:  %s\n", url)
			fmt.Printf("Mirror:    %s\n", mirrorURL)
			fmt.Println()

			// Copy to clipboard
			if err := clipboard.WriteAll(mirrorURL); err != nil {
				log.Printf("Warning: Failed to copy to clipboard: %v", err)
			} else {
				fmt.Println("Mirror URL copied to clipboard!")
			}
			break // Only process the first URL for clipboard operation
		}
	}
}

func convertToMirror(url string) string {
	re := regexp.MustCompile(githubURLPattern)
	matches := re.FindStringSubmatch(url)

	if len(matches) != 3 {
		fmt.Printf("Warning: '%s' is not a valid GitHub URL\n", url)
		return ""
	}

	owner := matches[1]
	repo := matches[2]

	// Remove .git suffix if present
	repo = strings.TrimSuffix(repo, ".git")

	return mirrorURLPrefix + owner + "/" + repo + ".git"
}

func showHelp() {
	fmt.Printf(`mirr - GitHub Mirror Converter

Convert GitHub URLs to gitcode mirror format and copy to clipboard.

Usage:
  mirr [options] [URL...]
  echo "https://github.com/user/repo" | mirr
  mirr -url https://github.com/user/repo

Options:
  -url string    GitHub URL to convert
  -help          Show this help message

Examples:
  mirr https://github.com/gorilla/mux
  mirr https://github.com/golang/go
  echo "https://github.com/kubernetes/kubernetes" | mirr

The converted URL will be copied to your clipboard automatically.
`)
}