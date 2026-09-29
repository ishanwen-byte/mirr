package main

import (
	"bytes"
	"errors"
	"os"
	"strings"
	"testing"
)

func TestConvertToMirror(t *testing.T) {
	const base = "https://ghfast.top"
	const wantPrefix = base + "/https://github.com/"

	tests := []struct {
		name  string
		input string
		want  string
	}{
		// Plain https form
		{"https", "https://github.com/golang/go", wantPrefix + "golang/go.git"},
		{"https with .git", "https://github.com/golang/go.git", wantPrefix + "golang/go.git"},
		{"https trailing slash", "https://github.com/golang/go/", wantPrefix + "golang/go.git"},
		{"http scheme", "http://github.com/golang/go", wantPrefix + "golang/go.git"},
		{"www prefix", "https://www.github.com/gorilla/mux", wantPrefix + "gorilla/mux.git"},
		{"mixed case host", "https://GITHUB.COM/golang/go", wantPrefix + "golang/go.git"},
		{"uppercase scheme", "HTTPS://GITHUB.COM/golang/go", wantPrefix + "golang/go.git"},
		{"mixed case no scheme", "www.GitHub.Com/golang/go", wantPrefix + "golang/go.git"},

		// Query strings and fragments are ignored
		{"query string", "https://github.com/golang/go?tab=readme", wantPrefix + "golang/go.git"},
		{"fragment", "https://github.com/gorilla/mux#readme", wantPrefix + "gorilla/mux.git"},
		{"query with path", "https://github.com/golang/go/tree/main?query=1#frag", wantPrefix + "golang/go.git"},
		{"tree path", "https://github.com/golang/go/tree/master/src", wantPrefix + "golang/go.git"},
		{"blob path", "https://github.com/gorilla/mux/blob/master/go.mod", wantPrefix + "gorilla/mux.git"},

		// Scheme-less forms
		{"scheme-less", "github.com/golang/go", wantPrefix + "golang/go.git"},
		{"scheme-less www", "www.github.com/golang/go", wantPrefix + "golang/go.git"},

		// SSH forms
		{"scp-style ssh", "git@github.com:golang/go.git", wantPrefix + "golang/go.git"},
		{"scp-style ssh no .git", "git@github.com:golang/go", wantPrefix + "golang/go.git"},
		{"ssh:// form", "ssh://git@github.com/golang/go.git", wantPrefix + "golang/go.git"},

		// Whitespace / BOM tolerance
		{"surrounding spaces", "  https://github.com/golang/go  ", wantPrefix + "golang/go.git"},
		{"PowerShell BOM", "\ufeffhttps://github.com/golang/go", wantPrefix + "golang/go.git"},

		// Unicode repo names (GitHub allows them)
		{"chinese repo name", "https://github.com/中文项目/仓库", wantPrefix + "中文项目/仓库.git"},

		// Invalid inputs
		{"not a url", "hello world", ""},
		{"gitlab url", "https://gitlab.com/gitlab-org/gitlab", ""},
		{"owner only", "https://github.com/golang", ""},
		{"empty string", "", ""},
		{"gists", "https://gist.github.com/golang/1234", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := convertToMirror(tt.input, base)
			if tt.want == "" {
				if err == nil {
					t.Fatalf("convertToMirror(%q) = %q, want error", tt.input, got)
				}
				if !errors.Is(err, errNotGitHub) {
					t.Fatalf("convertToMirror(%q) error = %v, want errNotGitHub", tt.input, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("convertToMirror(%q) unexpected error: %v", tt.input, err)
			}
			if got != tt.want {
				t.Errorf("convertToMirror(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestConvertToMirrorCustomMirror(t *testing.T) {
	got, err := convertToMirror("https://github.com/golang/go", "https://ghproxy.net/")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "https://ghproxy.net/https://github.com/golang/go.git"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestRunMultipleURLs(t *testing.T) {
	var out, errBuf bytes.Buffer
	code := run(options{
		urls:   []string{"https://github.com/golang/go", "bad input", "https://github.com/gorilla/mux"},
		mirror: "https://ghfast.top",
		noCopy: true,
	}, strings.NewReader(""), &out, &errBuf)

	if code != 0 {
		t.Fatalf("run() exit code = %d, want 0", code)
	}
	got := out.String()
	want := "https://ghfast.top/https://github.com/golang/go.git\n" +
		"https://ghfast.top/https://github.com/gorilla/mux.git\n"
	if got != want {
		t.Errorf("stdout = %q, want %q", got, want)
	}
	if !strings.Contains(errBuf.String(), "bad input") {
		t.Errorf("stderr should mention the invalid input, got %q", errBuf.String())
	}
}

func TestRunAllInvalidExitsNonZero(t *testing.T) {
	var out, errBuf bytes.Buffer
	code := run(options{
		urls:   []string{"not a url"},
		mirror: "https://ghfast.top",
		noCopy: true,
	}, strings.NewReader(""), &out, &errBuf)

	if code != 1 {
		t.Fatalf("run() exit code = %d, want 1", code)
	}
	if out.Len() != 0 {
		t.Errorf("stdout should be empty, got %q", out.String())
	}
}

func TestRunStdinLongLine(t *testing.T) {
	// A line beyond the default 64 KiB Scanner limit must not kill the run;
	// it is simply invalid input and gets a warning.
	long := strings.Repeat("a", 100*1024)
	in := "https://github.com/golang/go\n" + long + "\n"
	var out, errBuf bytes.Buffer
	code := run(options{mirror: "https://ghfast.top", noCopy: true},
		strings.NewReader(in), &out, &errBuf)
	if code != 0 {
		t.Fatalf("exit code = %d, want 0 (valid line should still convert)", code)
	}
	if !strings.Contains(out.String(), "golang/go.git") {
		t.Errorf("stdout = %q, want golang/go.git", out.String())
	}
}

func TestRunStdin(t *testing.T) {
	var out, errBuf bytes.Buffer
	code := run(options{
		mirror: "https://ghfast.top",
		noCopy: true,
	}, strings.NewReader("\n  https://github.com/golang/go  \n\nhttps://github.com/gorilla/mux\n"), &out, &errBuf)

	if code != 0 {
		t.Fatalf("run() exit code = %d, want 0", code)
	}
	want := "https://ghfast.top/https://github.com/golang/go.git\n" +
		"https://ghfast.top/https://github.com/gorilla/mux.git\n"
	if out.String() != want {
		t.Errorf("stdout = %q, want %q", out.String(), want)
	}
}

func TestRunNoInput(t *testing.T) {
	var out, errBuf bytes.Buffer
	code := run(options{mirror: "https://ghfast.top", noCopy: true},
		strings.NewReader(""), &out, &errBuf)
	if code != 1 {
		t.Fatalf("run() exit code = %d, want 1", code)
	}
	if !strings.Contains(errBuf.String(), "no URLs provided") {
		t.Errorf("stderr = %q, want usage hint", errBuf.String())
	}
}

func TestRunCloneSuccess(t *testing.T) {
	var out, errBuf bytes.Buffer
	var cloned []string
	code := run(options{
		urls:   []string{"https://github.com/golang/go"},
		mirror: "https://ghfast.top",
		clone:  true,
		cloneFn: func(url string) error {
			cloned = append(cloned, url)
			return nil
		},
	}, strings.NewReader(""), &out, &errBuf)

	if code != 0 {
		t.Fatalf("exit code = %d, want 0", code)
	}
	want := "https://ghfast.top/https://github.com/golang/go.git"
	if len(cloned) != 1 || cloned[0] != want {
		t.Errorf("cloned = %v, want [%s]", cloned, want)
	}
	// -clone must skip the clipboard path entirely.
	if strings.Contains(errBuf.String(), "clipboard") {
		t.Errorf("stderr should not mention clipboard with -clone, got %q", errBuf.String())
	}
}

func TestRunCloneAllFailed(t *testing.T) {
	var out, errBuf bytes.Buffer
	code := run(options{
		urls:   []string{"https://github.com/golang/go"},
		mirror: "https://ghfast.top",
		clone:  true,
		cloneFn: func(url string) error {
			return errors.New("connection refused")
		},
	}, strings.NewReader(""), &out, &errBuf)

	if code != 1 {
		t.Fatalf("exit code = clone failure should be 1, got %d", code)
	}
	if !strings.Contains(errBuf.String(), "all clones failed") {
		t.Errorf("stderr = %q, want 'all clones failed'", errBuf.String())
	}
	// The mirror URL must still reach stdout for scripted use.
	if !strings.Contains(out.String(), "golang/go.git") {
		t.Errorf("stdout = %q, want mirror URL", out.String())
	}
}

func TestRunClonePartialFailure(t *testing.T) {
	var out, errBuf bytes.Buffer
	var cloned []string
	code := run(options{
		urls:   []string{"https://github.com/golang/go", "https://github.com/gorilla/mux"},
		mirror: "https://ghfast.top",
		clone:  true,
		cloneFn: func(url string) error {
			if strings.Contains(url, "golang/go") {
				return errors.New("network down")
			}
			cloned = append(cloned, url)
			return nil
		},
	}, strings.NewReader(""), &out, &errBuf)

	if code != 0 {
		t.Fatalf("partial success exit code = %d, want 0", code)
	}
	if len(cloned) != 1 || !strings.Contains(cloned[0], "gorilla/mux") {
		t.Errorf("cloned = %v, want only gorilla/mux", cloned)
	}
	if !strings.Contains(errBuf.String(), "clone failed") {
		t.Errorf("stderr = %q, want clone failure warning", errBuf.String())
	}
}

func TestRunCloneFromStdin(t *testing.T) {
	var out, errBuf bytes.Buffer
	var cloned []string
	code := run(options{
		mirror: "https://ghfast.top",
		clone:  true,
		cloneFn: func(url string) error {
			cloned = append(cloned, url)
			return nil
		},
	}, strings.NewReader("https://github.com/golang/go\n"), &out, &errBuf)

	if code != 0 {
		t.Fatalf("exit code = %d, want 0", code)
	}
	if len(cloned) != 1 || !strings.Contains(cloned[0], "golang/go.git") {
		t.Errorf("cloned = %v, want golang/go.git", cloned)
	}
}

func TestMain(m *testing.M) {
	// clipboard is exercised in real usage; unit tests always pass -no-copy.
	os.Exit(m.Run())
}
