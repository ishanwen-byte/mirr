---
name: mirr
description: Bypass GitHub connectivity problems (slow, timeout, blocked, unreachable) to download GitHub resources — repos, release files, raw files, go modules — through a mirror prefix. Use this skill whenever git clone/fetch, go get/install, curl/wget of any github.com URL fails, times out, or crawls, and when the user mentions GitHub 加速/镜像/中转/down/blocked. Use it proactively before any large clone even if GitHub currently works.
---

# mirr — download GitHub resources when GitHub is unreachable

`mirr` rewrites any `github.com` URL into a mirror-prefixed URL
(default `https://ghfast.top`). The rewritten URL points to the **same
resource**, served through the mirror — so `git clone`, `curl`, `go get`
all keep working when direct GitHub access is slow or blocked.

```
https://github.com/owner/repo
  ->  https://ghfast.top/https://github.com/owner/repo.git
```

Any of these URLs, prefixed with the mirror, downloads the same bytes as
direct GitHub access:

```
https://github.com/owner/repo/archive/refs/tags/v1.0.tar.gz   (release/archive)
https://github.com/owner/repo/raw/main/file.txt               (raw file)
https://github.com/owner/repo/releases/download/v1.0/bin.zip  (release asset)
```

## Install the tool (once per machine)

Try in order, stop at the first success:

```bash
mirr -version
go install github.com/ishanwen-byte/mirr@latest
git clone --depth 1 https://github.com/ishanwen-byte/mirr.git "%TEMP%\mirr-src" ^
  && cd /d "%TEMP%\mirr-src" && go build -o "%USERPROFILE%\.local\bin\mirr.exe" .
```

(Linux/macOS: `mirr` without `.exe`; if GitHub itself is unreachable for
the bootstrap clone, build mirr from a copy already on disk, or rewrite
URLs by hand — see Fallback.)

## Decision flow

```
Need something from github.com?
├─ binary available (mirr -version works)?
│   ├─ yes → use mirr (below)
│   └─ no  → install it; if that fails → Fallback (hand rewrite)
└─ what kind of resource?
    ├─ whole repo         → mirr -clone -no-copy <url>
    ├─ URL text only      → mirr -no-copy <url>
    └─ file inside repo   → mirr -no-copy <raw-url>
```

## Recipes (agent must always pass -no-copy)

### Whole repo — clone through mirror
```bash
mirr -clone -no-copy https://github.com/golang/go
# exit 0: clone succeeded. exit 1: all clones failed → try next mirror
```

### Batch clone / convert (stdin, one URL per line)
```bash
mirr -clone -no-copy < repos.txt
```

### Release asset / archive / raw file — download through mirror
```bash
M=$(mirr -no-copy -mirror https://ghfast.top/ https://github.com/owner/repo/releases/download/v1.0/bin.zip)
curl -L -o bin.zip "$M"
```

`mirr` only normalizes to `<mirror>/https://github.com/<owner>/<repo>.git`
for repo URLs; for deep paths like `releases/download/...` or `raw/...`,
keep the original path after the prefix:
```
<original-full-url>  ->  <mirror>/<original-full-url>
```

### go get / go install from GitHub when proxy.golang.org is also blocked
```bash
GOFLAGS=-mod=mod go get github.com/owner/repo@v1.2.3
# If module fetch fails, and GOPROXY is already tried:
export GOPROXY=https://goproxy.cn,direct
go get github.com/owner/repo@v1.2.3
```
(Prefer GOPROXY for Go modules — it is the correct mechanism; use mirr
for git clones and file downloads.)

### Raw file through mirror
```bash
mirr -no-copy https://github.com/owner/repo/raw/main/README.md
# -> https://ghfast.top/https://github.com/owner/repo.git   (repo URL)
# For the actual file content, prefix the raw URL manually:
curl -L https://ghfast.top/https://github.com/owner/repo/raw/main/README.md
```

## Agent rules

1. **Always `-no-copy`.** Clipboard is for humans; stdout is for you.
2. **Read stdout** for converted URLs; warnings are on stderr.
3. **Mirror down? Rotate:** `-mirror https://ghproxy.net/`, then other known
   mirrors (ghfast.top, ghproxy.net, gh-proxy.com), then direct GitHub as
   last resort, and note it in the final reply.
4. **Exit codes:** 0 success / 1 no valid URL or all clones failed / 2 bad flags.
5. **Only github.com.** gitlab.com/gitee.com/bitbucket.org are rejected —
   don't retry them through mirr.
6. **Nested clones:** if a cloned repo's submodules or scripts reference
   github.com URLs, convert those too (`mirr -no-copy <url>`) before running.
7. **Don't pass GitHub token/env to the mirror.** Mirror is a third party:
   never send `GITHUB_TOKEN`, auth headers, or private repo URLs through it.
   Private repos must be cloned directly from GitHub, not through any mirror.

## Fallback (binary unavailable)

Hand-rewrite the URL — it is a pure string operation:
```
https://github.com/owner/repo(.git)?(/...)?
  -> https://ghfast.top/https://github.com/owner/repo.git        (repo)
  -> https://ghfast.top/<full original URL>                       (deep path)
```
SSH form `git@github.com:owner/repo` → take owner/repo → same as above.
Keep mirrors' list handy: ghfast.top, ghproxy.net, gh-proxy.com.
