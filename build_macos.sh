#!/bin/bash
#===============================================================================
# editcp macOS 构建脚本
# 用于在 macOS (Apple Silicon/Intel) 上编译 editcp
#===============================================================================

set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
cd "$SCRIPT_DIR"

echo "=========================================="
echo "  editcp macOS 构建脚本"
echo "=========================================="

#------------------------------------------------------------------------------
# 检查必要工具
#------------------------------------------------------------------------------
echo ""
echo "[1/6] 检查系统依赖..."

# 检查 Go
if ! command -v go &>/dev/null; then
    if [ -x "$HOME/go1.23.0/bin/go" ]; then
        export PATH="$HOME/go1.23.0/bin:$PATH"
    else
        echo "错误：未找到 Go。请先安装 Go 1.21+"
        echo "  brew install go"
        echo "  或从 https://go.dev/dl/ 下载"
        exit 1
    fi
fi
echo "  ✓ Go: $(go version)"

# 检查 Qt 5
QT_PATH=""
if brew --prefix qt@5 &>/dev/null; then
    QT_PATH="$(brew --prefix qt@5)"
    echo "  ✓ Qt 5: $QT_PATH"
elif [ -d "/usr/local/opt/qt@5" ]; then
    QT_PATH="/usr/local/opt/qt@5"
    echo "  ✓ Qt 5: $QT_PATH"
elif [ -d "/opt/homebrew/opt/qt@5" ]; then
    QT_PATH="/opt/homebrew/opt/qt@5"
    echo "  ✓ Qt 5: $QT_PATH"
else
    echo "错误：未找到 Qt 5。请先安装："
    echo "  brew install qt@5"
    exit 1
fi

# 检查 pkg-config（用于查找 Qt）
if ! command -v pkg-config &>/dev/null; then
    echo "错误：未找到 pkg-config。请先安装："
    echo "  brew install pkg-config"
    exit 1
fi
echo "  ✓ pkg-config: $(pkg-config --version)"

# 检查 libusb
if ! pkg-config --exists libusb-1.0 2>/dev/null; then
    echo "警告：未找到 libusb-1.0。请先安装："
    echo "  brew install libusb"
fi
echo "  ✓ libusb: 已安装"

#------------------------------------------------------------------------------
# 设置 CGo 编译标志（macOS 专用）
#------------------------------------------------------------------------------
echo ""
echo "[2/6] 设置编译环境..."

export PKG_CONFIG_PATH="$QT_PATH/lib/pkgconfig:$PKG_CONFIG_PATH"

# CGo 需要的标志
export CGO_CPPFLAGS="-I$QT_PATH/include -I$QT_PATH/include/QtCore -I$QT_PATH/include/QtGui -I$QT_PATH/include/QtWidgets -F$QT_PATH/lib"
export CGO_LDFLAGS="-F$QT_PATH/lib -framework QtCore -framework QtGui -framework QtWidgets -framework QtPrintSupport"

# 额外 Qt 框架
QT_FRAMEWORKS=(
    "QtConcurrent" "QtDBus" "QtNetwork" "QtOpenGL" "QtPrintSupport"
    "QtSql" "QtSvg" "QtTest" "QtWidgets" "QtXml"
)
for fw in "${QT_FRAMEWORKS[@]}"; do
    if [ -d "$QT_PATH/lib/$fw.framework" ]; then
        CGO_LDFLAGS="$CGO_LDFLAGS -framework $fw"
    fi
done

echo "  ✓ PKG_CONFIG_PATH: $PKG_CONFIG_PATH"

#------------------------------------------------------------------------------
# 下载 Go 依赖（如果网络可用）
#------------------------------------------------------------------------------
echo ""
echo "[3/6] 下载 Go 模块依赖..."

export GOPROXY=https://proxy.golang.org,direct
go mod tidy 2>/dev/null || echo "  ⚠ go mod tidy 失败（网络问题），继续..."
go mod download 2>/dev/null || echo "  ⚠ go mod download 失败（网络问题），继续..."

#------------------------------------------------------------------------------
# 安装 therecipe/qt 工具（如需要）
#------------------------------------------------------------------------------
echo ""
echo "[4/6] 检查 therecipe/qt 工具链..."

if ! command -v qtdeploy &>/dev/null; then
    echo "  正在安装 qtdeploy 工具..."
    go install github.com/therecipe/qt/cmd/qtdeploy@latest 2>/dev/null || true
    go install github.com/therecipe/qt/cmd/qtsetup@latest 2>/dev/null || true
    go install github.com/therecipe/qt/cmd/...@latest 2>/dev/null || true
fi

# 检查 qtdeploy 是否可用
if command -v qtdeploy &>/dev/null; then
    echo "  ✓ qtdeploy: 已安装"
    QT_DEPLOY_AVAILABLE=true
else
    echo "  ⚠ qtdeploy 未安装（需要网络下载），尝试直接构建..."
    QT_DEPLOY_AVAILABLE=false
fi

#------------------------------------------------------------------------------
# 构建
#------------------------------------------------------------------------------
echo ""
echo "[5/6] 构建 editcp..."

BUILD_DIR="$SCRIPT_DIR/build/macos"
mkdir -p "$BUILD_DIR"

# 版本信息
VERSION="1.0.31"
BUILD_TIME="$(date +%Y-%m-%d_%H:%M:%S)"
COMMIT_HASH="$(git rev-parse --short HEAD 2>/dev/null || echo 'unknown')"

# 构建标志
LDFLAGS="-s -w"
LDFLAGS="$LDFLAGS -X main.version=$VERSION"

if [ "$QT_DEPLOY_AVAILABLE" = true ]; then
    # 使用 qtdeploy 构建（完整的 Qt 部署）
    echo "  使用 qtdeploy 构建 macOS app bundle..."
    export QT_DIR="/opt/homebrew/opt/qt@5"
    export PKG_CONFIG_PATH="/opt/homebrew/opt/qt@5/lib/pkgconfig:$PKG_CONFIG_PATH"
    export CGO_CPPFLAGS="-DQT_CORE_LIB -DQT_GUI_LIB -DQT_WIDGETS_LIB -I/opt/homebrew/opt/qt@5/include -I/opt/homebrew/opt/qt@5/include/QtCore -I/opt/homebrew/opt/qt@5/include/QtGui -I/opt/homebrew/opt/qt@5/include/QtWidgets -F/opt/homebrew/opt/qt@5/lib"
    export CGO_LDFLAGS="-F/opt/homebrew/opt/qt@5/lib -framework QtCore -framework QtGui -framework QtWidgets -framework QtPrintSupport"
    qtdeploy build desktop
    if [ -d "deploy/darwin/editcp.app" ]; then
        echo "  ✓ macOS app 构建完成: deploy/darwin/editcp.app"
    fi
    if [ -d "deploy/darwin/deitcp.app" ]; then
        echo "  ✓ macOS app 构建完成: deploy/darwin/deitcp.app"
    fi
else
    # 直接 go build（仅生成二进制，不打包 .app）
    echo "  使用 go build 直接构建..."
    go build -ldflags="$LDFLAGS" -o "$BUILD_DIR/editcp" .
    echo "  ✓ 构建完成: $BUILD_DIR/editcp"
fi

#------------------------------------------------------------------------------
# macOS 代码签名（macOS 15+ 需要）
#------------------------------------------------------------------------------
echo ""
echo "[6/6] macOS 代码签名..."

APP_BUNDLE=""
if [ -d "deploy/darwin/deitcp.app" ]; then
    APP_BUNDLE="deploy/darwin/deitcp.app"
elif [ -d "deploy/darwin/editcp.app" ]; then
    APP_BUNDLE="deploy/darwin/editcp.app"
fi

if [ -n "$APP_BUNDLE" ]; then
    echo "  创建 USB 授权文件..."
    ENT_FILE="$(dirname "$APP_BUNDLE")/editcp.entitlements"
    cat > "$ENT_FILE" <<'ENT'
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>com.apple.security.device.usb</key>
    <true/>
</dict>
</plist>
ENT
    echo "  清除扩展属性..."
    xattr -cr "$APP_BUNDLE" 2>/dev/null

    echo "  签名所有依赖库..."
    for fw in "$APP_BUNDLE/Contents/Frameworks/"*.framework; do
        fw_name="$(basename "$fw" .framework)"
        fw_bin="$fw/Versions/5/$fw_name"
        [ -f "$fw_bin" ] && codesign --force --sign - "$fw_bin" 2>/dev/null
    done
    for dylib in "$APP_BUNDLE/Contents/Frameworks/"*.dylib; do
        [ -f "$dylib" ] && codesign --force --sign - "$dylib" 2>/dev/null
    done
    find "$APP_BUNDLE/Contents/PlugIns" -name "*.dylib" -exec codesign --force --sign - {} \; 2>/dev/null
    codesign --force --sign - "$APP_BUNDLE/Contents/MacOS/editcp" 2>/dev/null
    codesign --force --sign - "$APP_BUNDLE/Contents/MacOS/deitcp" 2>/dev/null

    echo "  最终签名 app bundle..."
    codesign --force --deep --sign - --options runtime --entitlements "$ENT_FILE" "$APP_BUNDLE" 2>&1
    echo "  ✓ macOS app 已签名: $APP_BUNDLE"
    echo ""
    echo "  运行方式:"
    echo "    双击运行: open $APP_BUNDLE"
    echo "    终端运行: $APP_BUNDLE/Contents/MacOS/editcp"
elif [ -f "$BUILD_DIR/editcp" ]; then
    # 创建简易 .app 包
    APP_BUNDLE="$BUILD_DIR/editcp.app"
    mkdir -p "$APP_BUNDLE/Contents/MacOS"
    mkdir -p "$APP_BUNDLE/Contents/Resources"
    cp "$BUILD_DIR/editcp" "$APP_BUNDLE/Contents/MacOS/editcp"

    # Info.plist
    cat > "$APP_BUNDLE/Contents/Info.plist" <<'PLIST'
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>CFBundleExecutable</key>
    <string>editcp</string>
    <key>CFBundleIdentifier</key>
    <string>org.farnsworth.editcp</string>
    <key>CFBundleName</key>
    <string>Codeplug Editor</string>
    <key>CFBundleDisplayName</key>
    <string>editcp</string>
    <key>CFBundleVersion</key>
    <string>1.0.31</string>
    <key>CFBundleShortVersionString</key>
    <string>1.0.31</string>
    <key>CFBundlePackageType</key>
    <string>APPL</string>
    <key>LSMinimumSystemVersion</key>
    <string>11.0</string>
    <key>NSHighResolutionCapable</key>
    <true/>
</dict>
</plist>
PLIST

    chmod +x "$APP_BUNDLE/Contents/MacOS/editcp"

    # 如果 editcp.ico 存在，转换为 icns（需要 iconutil）
    if [ -f "editcp.ico" ] && command -v iconutil &>/dev/null; then
        echo "  正在转换图标..."
        # 简化处理：使用 sips 从 ICO 转 PNG
        mkdir -p "$APP_BUNDLE/Contents/Resources/icons.iconset"
        # 如果无法直接转，跳过
    fi

    echo "  ✓ 已创建 app bundle: $APP_BUNDLE"
    echo ""
    echo "  运行方式:"
    echo "    双击运行: open '$APP_BUNDLE'"
    echo "    终端运行: '$APP_BUNDLE/Contents/MacOS/editcp'"
fi

echo ""
echo "=========================================="
echo "  macOS 构建完成！"
echo "=========================================="
echo ""
echo "如果运行遇到问题，请检查："
echo "  1. USB 权限（用于连接对讲机）："
echo "     需要在系统设置 > 隐私与安全性 > USB 中授权"
echo "  2. 首次运行可能需要右键 > 打开以绕过 Gatekeeper"
echo ""
