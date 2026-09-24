# mirr - GitHub Mirror Converter

A simple Go CLI tool that converts GitHub URLs to a mirror prefix (default: `ghfast.top`) and automatically copies the converted URL(s) to clipboard.

## Features

- Convert GitHub URLs to mirror format, one result per line on stdout
- Accepts many URL forms: `https://`, `http://`, `www.`, scheme-less, SSH (`git@github.com:owner/repo`, `ssh://git@github.com/owner/repo`), and URLs with extra paths (`/tree/main/...`) or trailing slashes
- Handles multiple URLs at once (all are converted, invalid ones produce warnings on stderr)
- Configurable mirror site via `-mirror`
- Pipe-friendly: results go to stdout, warnings/status to stderr
- Meaningful exit codes (0 = at least one conversion, 1 = none) for scripts and CI
- Automatically copy converted URL(s) to clipboard (opt out with `-no-copy`)

## Installation

```bash
git clone ssh://REDACTED-USER@REDACTED-PRIVATE-HOST:8022/goliath/mirr.git
cd mirr
go build -o mirr.exe .
```

## Usage

### Command line arguments
```bash
mirr https://github.com/golang/go
```

### Multiple URLs
```bash
mirr https://github.com/golang/go git@github.com:gorilla/mux.git github.com/kubernetes/kubernetes
```

### Using the `-url` flag
```bash
mirr -url https://github.com/kubernetes/kubernetes
```

### Piped input
```bash
echo "https://github.com/gorilla/mux" | mirr
```

### Custom mirror site
```bash
mirr -mirror https://ghproxy.net/ https://github.com/golang/go
```

### Disable clipboard, print only
```bash
mirr -no-copy https://github.com/golang/go
```

### Help
```bash
mirr -help
```

## Example Output

```
$ mirr https://github.com/golang/go
https://github.com/golang/go
  -> https://ghfast.top/https://github.com/golang/go.git
Mirror URL copied to clipboard!
```

(The first two lines and the clipboard notice go to stderr; the mirror URL itself goes to stdout, so `mirr URL | xargs git clone` works.)

## Conversion Format

The tool converts GitHub URLs like:

```
https://github.com/[owner]/[repo]
git@github.com:[owner]/[repo].git
github.com/[owner]/[repo]
https://github.com/[owner]/[repo]/tree/main   (extra paths ignored)
```

To:

```
[mirror]/https://github.com/[owner]/[repo].git
```

## Exit Codes

- `0`: at least one URL was converted
- `1`: no valid GitHub URL found (or a usage/stdin error)

## Development

```bash
go test ./...    # run unit tests
go vet ./...     # static analysis
```

## Requirements

- Go 1.16 or later
- For clipboard functionality:
  - Windows: No additional requirements
  - Linux: xclip or xsel
  - macOS: pbcopy

## License

MIT License
