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

package main

import (
	"strings"

	"github.com/dalefarnsworth-dmr/ui"
)

func errorText(edt *editor) string {
	cp := edt.codeplug
	status := "未发现错误"
	errMsg := ""

	if !cp.Valid() {
		status = "以下字段值无效："
		errMsg = strings.Join(cp.Warnings(), "\n")
	}
	edt.updateMenuBar()

	return status + "\n\n" + errMsg
}

func checkCodeplug(edt *editor) {
	w := edt.mainWindow.NewWindow()
	cp := edt.codeplug
	w.SetTitle(cp.Filename() + edt.titleSuffix() + " 无效字段")

	windowBox := w.AddVbox()

	var t *ui.TextEdit

	b := windowBox.AddButton("重新扫描无效值")
	b.ConnectClicked(func() {
		t.SetPlainText(errorText(edt))
	})

	t = windowBox.AddTextEdit()
	t.SetPlainText(errorText(edt))
	t.SetReadOnly(true)

	w.Show()
}
