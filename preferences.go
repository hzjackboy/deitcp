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
	"github.com/dalefarnsworth-dmr/ui"
)

func (edt *editor) preferences() {
	dialog := ui.NewDialog("偏好设置")

	loadSettings()

	row := dialog.AddHbox()
	groupBox := row.AddGroupbox("选项")
	form := groupBox.AddForm()

	editButtonsEnabled := settings.editButtonsEnabled
	checked := editButtonsEnabled
	checkbox := ui.NewCheckboxWidget(checked, func(checked bool) {
		editButtonsEnabled = checked
	})
	form.AddRow("在主窗口启用编辑按钮：", checkbox)

	gpsEnabled := settings.gpsEnabled
	checked = gpsEnabled
	checkbox = ui.NewCheckboxWidget(checked, func(checked bool) {
		gpsEnabled = checked
	})
	form.AddRow("显示 GPS 字段：", checkbox)

	uniqueContactNames := settings.uniqueContactNames
	checked = uniqueContactNames
	checkbox = ui.NewCheckboxWidget(checked, func(checked bool) {
		uniqueContactNames = checked
	})
	form.AddRow("要求联系人名称唯一：", checkbox)

	autosaveInterval := settings.autosaveInterval

	spinbox := ui.NewSpinboxWidget(autosaveInterval, 0, 60, func(i int) {
		autosaveInterval = i
	})
	form.AddRow("自动保存间隔(分钟)：", spinbox)

	var experimental bool
	if needExperimental {
		experimental = settings.experimental
		checked = experimental
		checkbox = ui.NewCheckboxWidget(checked, func(checked bool) {
			experimental = checked
		})
		form.AddRow("启用实验性功能：", checkbox)
	}

	dialog.AddSpace(2)

	row = dialog.AddHbox()

	cancelButton := ui.NewButtonWidget("Cancel", func() {
		dialog.Reject()
	})
	row.AddWidget(cancelButton)

	okButton := ui.NewButtonWidget("保存", func() {
		dialog.Accept()
	})
	row.AddWidget(okButton)

	if !dialog.Exec() {
		return
	}

	settings.editButtonsEnabled = editButtonsEnabled
	edt.updateButtons()

	settings.gpsEnabled = gpsEnabled
	edt.setGPSEnabled(gpsEnabled)
	cp := edt.codeplug
	if cp != nil {
		cp.SetGPSEnabled(gpsEnabled)
	}

	settings.uniqueContactNames = uniqueContactNames

	settings.experimental = experimental

	settings.autosaveInterval = autosaveInterval
	edt.setAutosaveInterval(autosaveInterval)

	edt.updateMenuBar()

	saveSettings()
}
