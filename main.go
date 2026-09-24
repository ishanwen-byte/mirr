package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"

	"github.com/atotto/clipboard"
)

const defaultMirror = "https://ghfast.top"

// Compiled once at package init instead of on every call.
var githubURLRe = regexp.MustCompile(
	`^https?://(?:www\.)?github\.com/([A-Za-z0-9_.-]+)/([A-Za-z0-9_.-]+?)(?:\.git)?(?:[/?#].*)?$`)

// errNotGitHub is returned when an input line is not a recognizable GitHub URL.
var errNotGitHub = errors.New("not a valid GitHub repository URL")

type options struct {
	urls   []string
	mirror string
	noCopy bool
}

func main() {
	var (
		urlFlag   = flag.String("url", "", "GitHub URL to convert")
		mirrorFlg = flag.String("mirror", defaultMirror, "mirror site base URL, e.g. "+defaultMirror)
		noCopyFlg = flag.Bool("no-copy", false, "do not copy result to clipboard")
		helpFlag  = flag.Bool("help", false, "show help information")
	)
	// Make -h and flag parse errors print the full custom help
	// instead of the bare generated flag list.
	flag.Usage = func() { showHelp(flag.CommandLine.Output()) }
	flag.Parse()

	if *helpFlag {
		showHelp(os.Stdout)
		return
	}

	var urls []string
	if *urlFlag != "" {
		urls = append(urls, *urlFlag)
	} else if flag.NArg() > 0 {
		urls = append(urls, flag.Args()...)
	}

	opts := options{urls: urls, mirror: *mirrorFlg, noCopy: *noCopyFlg}
	os.Exit(run(opts, os.Stdin, os.Stdout, os.Stderr))
}

// run converts every URL and prints one mirror URL per line to stdout.
// Human-readable progress, warnings, and clipboard status go to stderr so
// stdout stays pipe-friendly. It returns the process exit code:
// 0 if at least one URL was converted, 1 otherwise.
func run(opts options, stdin io.Reader, stdout, stderr io.Writer) int {
	urls := opts.urls
	if len(urls) == 0 {
		scanner := bufio.NewScanner(stdin)
		// Handle absurdly long lines (e.g. pasted junk) instead of failing
		// with the default 64 KiB Scanner limit.
		scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
		for scanner.Scan() {
			if line := strings.TrimSpace(scanner.Text()); line != "" {
				urls = append(urls, line)
			}
		}
		if err := scanner.Err(); err != nil {
			fmt.Fprintf(stderr, "mirr: reading stdin: %v\n", err)
			return 1
		}
	}

	if len(urls) == 0 {
		fmt.Fprintln(stderr, "mirr: no URLs provided. Use -help for usage information.")
		return 1
	}

	var converted []string
	for _, u := range urls {
		mirror, err := convertToMirror(u, opts.mirror)
		if err != nil {
			fmt.Fprintf(stderr, "mirr: warning: %v\n", err)
			continue
		}
		fmt.Fprintf(stderr, "%s\n  -> %s\n", u, mirror)
		fmt.Fprintln(stdout, mirror)
		converted = append(converted, mirror)
	}

	if len(converted) == 0 {
		fmt.Fprintln(stderr, "mirr: no valid GitHub URLs found")
		return 1
	}

	if !opts.noCopy {
		if err := clipboard.WriteAll(strings.Join(converted, "\n")); err != nil {
			fmt.Fprintf(stderr, "mirr: warning: clipboard copy failed: %v\n", err)
		} else if len(converted) == 1 {
			fmt.Fprintln(stderr, "Mirror URL copied to clipboard!")
		} else {
			fmt.Fprintf(stderr, "%d mirror URLs copied to clipboard!\n", len(converted))
		}
	}
	return 0
}

// normalize maps common GitHub URL variants onto a canonical https URL:
// SSH forms (git@github.com:owner/repo, ssh://git@github.com/owner/repo)
// and scheme-less forms (github.com/owner/repo, www.github.com/owner/repo).
func normalize(input string) string {
	s := strings.TrimSpace(input)
	lower := strings.ToLower(s)
	switch {
	case strings.HasPrefix(lower, "git@github.com:"):
		s = "https://github.com/" + s[len("git@github.com:"):]
	case strings.HasPrefix(lower, "ssh://git@github.com/"):
		s = "https://github.com/" + s[len("ssh://git@github.com/"):]
	case strings.HasPrefix(lower, "www.github.com/"):
		s = "https://" + s[len("www."):]
	case strings.HasPrefix(lower, "github.com/"):
		s = "https://" + s
	}
	// Schemes and hosts are case-insensitive: HTTPS://GITHUB.COM/... must match.
	if i := strings.Index(strings.ToLower(s), "://"); i >= 0 {
		rest := s[i+3:]
		if slash := strings.IndexByte(rest, '/'); slash >= 0 {
			return strings.ToLower(s[:i+3]) + strings.ToLower(rest[:slash]) + rest[slash:]
		}
		return strings.ToLower(s)
	}
	return s
}

// convertToMirror turns any supported GitHub URL into
// <mirrorBase>/https://github.com/<owner>/<repo>.git.
func convertToMirror(input, mirrorBase string) (string, error) {
	m := githubURLRe.FindStringSubmatch(normalize(input))
	if m == nil {
		return "", fmt.Errorf("%w: %s", errNotGitHub, input)
	}
	owner, repo := m[1], strings.TrimSuffix(m[2], ".git")
	return strings.TrimSuffix(mirrorBase, "/") + "/https://github.com/" + owner + "/" + repo + ".git", nil
}

func showHelp(w io.Writer) {
	fmt.Fprintf(w, `mirr - GitHub Mirror Converter

Convert GitHub URLs to a mirror prefix (default: %[1]s) and copy to clipboard.

Usage:
  mirr [options] [URL...]
  echo "https://github.com/user/repo" | mirr
  mirr -url https://github.com/user/repo

Options:
  -url string    GitHub URL to convert
  -mirror string mirror site base URL (default %[1]q)
  -no-copy       print only, skip clipboard
  -help          show this help message

Accepted input forms include:
  https://github.com/owner/repo          https://github.com/owner/repo.git
  http://www.github.com/owner/repo/      github.com/owner/repo
  git@github.com:owner/repo.git          ssh://git@github.com/owner/repo
  https://github.com/owner/repo/tree/main  (paths, ?queries, #fragments ignored)

Output:
  One converted mirror URL per line on stdout (pipe-friendly).
  Warnings and clipboard status are printed to stderr.

Exit codes:
  0  at least one URL was converted
  1  no valid GitHub URL found (or usage/stdin error)
`, defaultMirror)
}
