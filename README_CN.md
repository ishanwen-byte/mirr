# mirr - GitHub 镜像转换工具

一个简单的 Go 命令行工具，将 GitHub URL 转换为 ghfast.top 镜像格式，并自动复制转换后的 URL 到剪贴板。

## 功能特性

- 将 GitHub URL 转换为 ghfast.top 镜像格式
- 自动复制转换后的 URL 到剪贴板
- 支持多种输入方式：
  - 命令行参数
  - `-url` 标志
  - 标准输入（管道）
- 完整的帮助和错误处理

## 安装

```bash
git clone ssh://REDACTED-USER@REDACTED-PRIVATE-HOST:8022/goliath/mirr.git
cd mirr
go mod tidy
go build -o mirr.exe .
```

## 使用方法

### 命令行参数
```bash
mirr https://github.com/golang/go
```

### 使用 `-url` 标志
```bash
mirr -url https://github.com/kubernetes/kubernetes
```

### 管道输入
```bash
echo "https://github.com/gorilla/mux" | mirr
```

### 查看帮助
```bash
mirr -help
```

## 示例输出

```
Original:  https://github.com/golang/go
Mirror:    https://ghfast.top/https://github.com/golang/go.git

镜像 URL 已复制到剪贴板！
```

## 转换格式

工具将 GitHub URL 从：
```
https://github.com/[所有者]/[仓库]
```

转换为：
```
https://ghfast.top/https://github.com/[所有者]/[仓库].git
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
- 在 CI/CD 脚本中使用镜像站
- 避免手动修改 URL 的繁琐过程

## 许可证

MIT License