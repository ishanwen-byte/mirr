# mirr - GitHub Mirror Converter

A simple Go CLI tool that converts GitHub URLs to ghfast.top mirror format and automatically copies the converted URL to clipboard.

## Features

- Convert GitHub URLs to ghfast.top mirror format
- Automatically copy converted URLs to clipboard
- Support multiple input methods:
  - Command line arguments
  - `-url` flag
  - Standard input (piped)
- Comprehensive help and error handling

## Installation

```bash
git clone ssh://REDACTED-USER@REDACTED-PRIVATE-HOST:8022/goliath/mirr.git
cd mirr
go mod tidy
go build -o mirr.exe .
```

## Usage

### Command line arguments
```bash
mirr https://github.com/golang/go
```

### Using the `-url` flag
```bash
mirr -url https://github.com/kubernetes/kubernetes
```

### Piped input
```bash
echo "https://github.com/gorilla/mux" | mirr
```

### Help
```bash
mirr -help
```

## Example Output

```
Original:  https://github.com/golang/go
Mirror:    https://ghfast.top/https://github.com/golang/go.git

Mirror URL copied to clipboard!
```

## Conversion Format

The tool converts GitHub URLs from:
```
https://github.com/[owner]/[repo]
```

To:
```
https://ghfast.top/https://github.com/[owner]/[repo].git
```

## Requirements

- Go 1.16 or later
- For clipboard functionality:
  - Windows: No additional requirements
  - Linux: xclip or xsel
  - macOS: pbcopy

## License

MIT License