// Copyright 2017-2021 Dale Farnsworth. All rights reserved.

// Dale Farnsworth
// 1007 W Mendoza Ave
// Mesa, AZ  85210
// USA
//
// dale@farnsworth.org

// This file is part of Editcp.
//
// Editcp is free software: you can redistribute it and/or modify
// it under the terms of version 3 of the GNU General Public License
// as published by the Free Software Foundation.
//
// Editcp is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with Editcp.  If not, see <http://www.gnu.org/licenses/>.

//go:build darwin

package main

import (
	"os"
	"os/signal"
	"syscall"
)

// platformInit 执行Darwin专用的初始化
// 在main函数开始时调用
func platformInit() {
	// macOS需要注册信号处理，使Ctrl+C能干净退出
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		os.Exit(0)
	}()
}

// platformSaveDebuggingInfo macOS上保存调试信息
// 在macOS上使用标准错误输出，不依赖Windows事件日志
func platformSaveDebuggingInfo() {
	// macOS不需要Windows的调试保存机制
	// 保留前一次崩溃信息（通过debug包处理）
}

// platformCacheDir 返回macOS上的缓存目录
func platformCacheDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "/tmp"
	}
	cacheDir := home + "/Library/Caches/editcp"
	os.MkdirAll(cacheDir, 0755)
	return cacheDir
}

// platformSettingsDir 返回macOS上的配置目录
func platformSettingsDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "."
	}
	settingsDir := home + "/Library/Preferences/editcp"
	os.MkdirAll(settingsDir, 0755)
	return settingsDir
}
