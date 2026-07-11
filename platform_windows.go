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

//go:build windows

package main

// platformInit 为Windows系统进行初始化
func platformInit() {
	// Windows的初始化由l.WindowsSaveDebuggingInfo()完成
}

// platformSaveDebuggingInfo Windows上保存调试信息
func platformSaveDebuggingInfo() {
	// 保持原有逻辑
}
