# mirr - GitHub 镜像转换工具

一个简单的 Go 命令行工具，将 GitHub URL 转换为镜像格式（默认 `ghfast.top`），并自动复制转换后的 URL 到剪贴板。

## 功能特性

- 将 GitHub URL 转换为镜像格式，结果逐行输出到 stdout
- 支持多种 URL 形式：`https://`、`http://`、`www.`、无协议、SSH（`git@github.com:owner/repo`、`ssh://git@github.com/owner/repo`）、带额外路径（`/tree/main/...`）或尾斜杠
- **`-clone`：转换并一步执行 `git clone`** —— 无需手动粘贴，git 在前台交互运行
- 支持一次处理多个 URL（全部转换，无效的在 stderr 上给出警告）
- 通过 `-mirror` 可配置镜像站点
- 对管道友好：转换结果走 stdout，警告和状态信息走 stderr
- 有意义的退出码（0 = 至少转换一个，1 = 一个都没有），便于脚本和 CI 判断
- 自动复制转换结果到剪贴板（可用 `-no-copy` 关闭；使用 `-clone` 时自动跳过）
- 内置 agent skill —— 见 [skills/mirr/SKILL.md](skills/mirr/SKILL.md)：安装后 agent 可在 GitHub 无法访问时照常克隆/下载 GitHub 资源

## 安装

```bash
go install github.com/ishanwen-byte/mirr@latest
```

或从源码构建：

```bash
git clone https://github.com/ishanwen-byte/mirr.git
cd mirr
go build -o mirr.exe .
```

## 使用方法

### 命令行参数
```bash
mirr https://github.com/golang/go
```

### 多个 URL
```bash
mirr https://github.com/golang/go git@github.com:gorilla/mux.git github.com/kubernetes/kubernetes
```

### 使用 `-url` 标志
```bash
mirr -url https://github.com/kubernetes/kubernetes
```

（若同时提供 `-url` 和位置参数，则只用 `-url`。）

### 管道输入
```bash
echo "https://github.com/gorilla/mux" | mirr
```

### 转换并一步克隆
```bash
mirr -clone https://github.com/golang/go
echo "https://github.com/gorilla/mux" | mirr -clone
```

`-clone` 会在前台执行 `git clone <镜像URL>`，进度显示和凭据提示与直接使用 git 相同，并自动跳过剪贴板。

### 自行接 git 管道
```bash
mirr -no-copy https://github.com/golang/go | xargs git clone
```

### 自定义镜像站点
```bash
mirr -mirror https://ghproxy.net/ https://github.com/golang/go
```

### 仅打印，不写剪贴板
```bash
mirr -no-copy https://github.com/golang/go
```

### 查看版本号
```bash
mirr -version
```

### 查看帮助
```bash
mirr -help
mirr -h
```

## 示例输出

```
$ mirr https://github.com/golang/go
https://github.com/golang/go
  -> https://ghfast.top/https://github.com/golang/go.git
镜像 URL 已复制到剪贴板！
```

（前两行和剪贴板提示走 stderr；镜像 URL 本身走 stdout，因此 `mirr URL | xargs git clone` 可以直接使用。）

## 转换格式

工具将如下形式的 GitHub URL：

```
https://github.com/[所有者]/[仓库]
git@github.com:[所有者]/[仓库].git
github.com/[所有者]/[仓库]
https://github.com/[所有者]/[仓库]/tree/main   （忽略额外路径）
```

转换为：

```
[镜像]/https://github.com/[所有者]/[仓库].git
```

## 退出码

- `0`：至少转换了一个 URL（使用 `-clone` 时至少一个克隆成功）
- `1`：没有找到有效的 GitHub URL，或（使用 `-clone` 时）全部克隆失败
- `2`：命令行标志无效（如未知标志）

## 版本化构建

构建时嵌入当前 git 版本：

```bash
go build -ldflags "-X main.version=$(git describe --tags --always)" -o mirr.exe .
```

## 开发

```bash
go test ./...    # 运行单元测试
go vet ./...     # 静态检查
```

## 系统要求

- Go 1.16 或更高版本
- 剪贴板功能需要：
  - Windows：无额外要求
  - Linux：xclip 或 xsel
  - macOS：pbcopy

## 使用场景

该工具特别适用于：
- 在国内访问 GitHub 较慢时，使用镜像站加速
- 快速转换 GitHub 仓库地址并复制到剪贴板
- 在 CI/CD 脚本中使用镜像站（可依据退出码判断成败）
- 批量转换多个仓库地址
- 避免手动修改 URL 的繁琐过程

## 许可证

MIT License
