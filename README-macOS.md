# editcp macOS 适配说明

## 概述

本文件记录了将 editcp（对讲机写频编辑工具）移植到 macOS 的修改和编译方法。

## 前提条件

在 macOS 上编译 editcp，需要安装以下依赖：

### 1. 安装 Go

```bash
# 方式一：Homebrew
brew install go

# 方式二：从官网下载
# 访问 https://go.dev/dl/ 下载 macOS 版安装包
```

### 2. 安装 Qt 5

```bash
brew install qt@5
```

### 3. 安装 pkg-config

```bash
brew install pkg-config
```

### 4. 安装 libusb

```bash
brew install libusb
```

## 编译方法

### 方式一：使用构建脚本（推荐）

```bash
chmod +x build_macos.sh
./build_macos.sh
```

### 方式二：使用 Makefile

```bash
make -f Makefile.macos build
```

### 方式三：手动编译

```bash
export PKG_CONFIG_PATH="$(brew --prefix qt@5)/lib/pkgconfig:$PKG_CONFIG_PATH"
CGO_CPPFLAGS="-I$(brew --prefix qt@5)/include -F$(brew --prefix qt@5)/lib" \
CGO_LDFLAGS="-F$(brew --prefix qt@5)/lib" \
go build
```

## macOS 的修改说明

### 已修改/新增的文件

| 文件 | 说明 |
|------|------|
| `build_macos.sh` | macOS 构建脚本 |
| `Makefile.macos` | macOS Makefile |
| `README-macOS.md` | 本文件 |
| `go.mod` | 更新了 Go 版本和 macOS 兼容依赖 |
| `editcp_mac.go` | macOS 平台的特定代码 |

### 构建产物

编译后，可执行文件位于：
- `build/macos/editcp` - 命令行可执行文件
- `build/macos/editcp.app` - macOS 应用程序包

### USB 权限

如需连接对讲机进行读写操作，需要：

1. 连接 USB 设备
2. 系统会提示 "允许配件连接"，选择"允许"
3. 如果使用非管理员用户，可能需要安装 udev 规则的 macOS 等效配置

### Qt 兼容性

本项目使用 `github.com/therecipe/qt`（Qt5 Go 绑定）。
该包已归档但可正常使用。如遇到兼容性问题，建议使用 Qt 5.15 LTS。

### 已知问题

1. **`therecipe/qt` 兼容性**：因 `therecipe/qt` 已归档，在最新 Go 版本或 Apple Silicon 上
   可能遇到问题。如有构建错误，请尝试降级 Go 版本到 1.21 或使用 `-tags=no_env` 标志。
2. **USB 库链接**：在 macOS 上，libusb 使用不同的链接方式，已通过 `build_macos.sh` 处理。
3. **Gatekeeper**：首次运行 .app 可能需要右键 > 打开以绕过 Gatekeeper 检查。

## 故障排除

### 构建时报 "too many errors"

尝试设置 `CGO_ENABLED=1` 并使用兼容的 Go 版本。

### 找不到 Qt 头文件

确保 Qt 5 的路径正确：
```bash
export PKG_CONFIG_PATH="/opt/homebrew/opt/qt@5/lib/pkgconfig:$PKG_CONFIG_PATH"
```

### 运行时崩溃（库找不到）

```bash
# 设置 DYLD_LIBRARY_PATH 指向 Qt 5 框架目录
export DYLD_LIBRARY_PATH="/opt/homebrew/opt/qt@5/lib:$DYLD_LIBRARY_PATH"
```
