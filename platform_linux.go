// Copyright 2017-2021 Dale Farnsworth. All rights reserved.
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

//go:build linux

package main

// platformInit 为Linux系统进行初始化
func platformInit() {
	// Linux不需要特殊初始化
}

// platformSaveDebuggingInfo Linux上保存调试信息
func platformSaveDebuggingInfo() {
	// 使用原有逻辑（通过debug包处理）
}
