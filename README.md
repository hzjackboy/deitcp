# 🎛️ editcp / Codeplug Editor

[![License: GPL v3](https://img.shields.io/badge/License-GPLv3-blue.svg)](https://www.gnu.org/licenses/gpl-3.0)

**editcp** — DMR 对讲机写频编辑工具，适用于 Tytera MD380/MD390、Alinco DJ-MD40 等机型。  
**editcp** — A codeplug editor for Tytera MD380/MD390, Alinco DJ-MD40, and compatible DMR radios.

---

## 🇨🇳 中文说明

### 简介

editcp 是一个开源的 DMR 对讲机写频（codeplug）编辑工具，功能类似于 TYT 官方 CPS 软件。
本项目基于 [DaleFarnsworth-DMR/editcp](https://github.com/DaleFarnsworth-DMR/editcp) 进行了 **macOS 移植和适配**。

### 主要功能

- 编辑 General Settings、Channels、Contacts、Zones、Group Lists、Scan Lists
- 拖拽排序列表项
- 同时打开多个写频文件，跨文件拖拽复制
- 无限撤销/重做
- 输入验证和写频完整性检查
- 导入/导出为文本文件、Excel 表格、JSON 格式
- **读/写对讲机写频数据**（通过 USB/DFU 协议）
- 写入固件和用户数据库（支持 md380tools）

### macOS 构建

```bash
# 前置要求（通过 Homebrew 安装）
brew install go qt@5 pkg-config libusb

# 一键构建
./build_macos.sh

# 或使用 Makefile
make -f Makefile.macos build
```

构建完成后，App 位于 `deploy/darwin/deitcp.app`。

### macOS 使用注意事项

1. **代码签名**：首次运行如果闪退，终端执行以下命令：
   ```bash
   codesign --force --deep --sign - deploy/darwin/deitcp.app
   ```
2. **USB 权限**：系统设置 → 隐私与安全性 → USB 配件 → 允许
3. **对讲机刷机模式**：关机 → 按住 **PTT + PTT 上方按键** → 开机（LED 红绿交替闪烁）
4. **首次运行**：如果提示"无法验证开发者"，右键 → 打开

---

## 🇬🇧 English README

### Introduction

editcp is an open-source codeplug editor for DMR radios, similar in purpose to the official TYT CPS software.  
This fork is a **macOS port** based on [DaleFarnsworth-DMR/editcp](https://github.com/DaleFarnsworth-DMR/editcp).

### Features

- Edit General Settings, Channels, Contacts, Zones, Group Lists, and Scan Lists
- Drag-and-drop reordering of list items
- Open multiple codeplugs simultaneously and copy items between them
- Unlimited undo/redo
- Comprehensive input validation and codeplug integrity checking
- Export/Import to/from plain text, Excel spreadsheets, and JSON
- **Read/Write codeplug data to/from radio** (via USB DFU protocol)
- Firmware and user database writing (md380tools support)

### macOS Build

```bash
# Prerequisites (install via Homebrew)
brew install go qt@5 pkg-config libusb

# One-click build
./build_macos.sh

# Or use the Makefile
make -f Makefile.macos build
```

After building, the app bundle is at `deploy/darwin/deitcp.app`.

### macOS Usage Notes

1. **Code Signing**: If the app crashes on first launch, run:
   ```bash
   codesign --force --deep --sign - deploy/darwin/deitcp.app
   ```
2. **USB Permission**: System Settings → Privacy & Security → USB Accessories → Allow
3. **Radio Bootloader Mode**: Power off → Hold **PTT + Button above PTT** → Power on (LED flashes green/red)
4. **Gatekeeper**: Right-click → Open if "unverified developer" warning appears

---

## 🛠️ Building on Linux / Windows

See the [original project](https://github.com/DaleFarnsworth-DMR/editcp) for Linux/Windows build instructions.

```bash
# Linux
make

# macOS
make -f Makefile.macos build

# Windows (cross-compile from Linux with Docker)
make windows
```

## 📄 License

GNU General Public License v3.0. See [LICENSE](LICENSE).

## 👤 Author

**Dale Farnsworth** — [dale@farnsworth.org](mailto:dale@farnsworth.org)  
macOS port by [hzjackboy](https://github.com/hzjackboy)
