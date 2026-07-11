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
	"github.com/dalefarnsworth-dmr/codeplug"
	"github.com/dalefarnsworth-dmr/editcp/lang"
	"github.com/dalefarnsworth-dmr/ui"
)

func generalSettings(edt *editor) {
	writable := false
	edt.newRecordWindow(codeplug.RtGeneralSettings_md380, writable, gsRecord)
}

func gsRecord(edt *editor, recordBox *ui.HBox) {
	r := currentRecord(recordBox.Window())

	mainBox := recordBox.AddVbox()
	row := mainBox.AddHbox()

	column := row.AddVbox()
	groupBox := column.AddGroupbox(lang.T("保存", "Save"))
	form := groupBox.AddForm()
	form.AddFieldTypeRows(r,
		codeplug.FtGsSavePreamble,
		codeplug.FtGsSaveModeReceive,
	)

	groupBox = column.AddGroupbox(lang.T("提示音", "Alert Tone"))
	form = groupBox.AddForm()
	form.AddFieldTypeRows(r,
		codeplug.FtGsDisableAllTones,
		codeplug.FtGsKeypadTones,
		codeplug.FtGsChFreeIndicationTone,
		codeplug.FtGsTalkPermitTone,
		codeplug.FtGsCallAlertToneDuration,
	)

	groupBox = column.AddGroupbox(lang.T("扫描", "Scan"))
	form = groupBox.AddForm()
	form.AddFieldTypeRows(r,
		codeplug.FtGsScanDigitalHangTime,
		codeplug.FtGsScanAnalogHangTime,
	)

	groupBox = column.AddGroupbox(lang.T("单独工作者", "Lone Worker"))
	form = groupBox.AddForm()
	form.AddFieldTypeRows(r,
		codeplug.FtGsLoneWorkerResponseTime,
		codeplug.FtGsLoneWorkerReminderTime,
	)

	groupBox = column.AddGroupbox(lang.T("开机密码", "Power On Password"))
	form = groupBox.AddForm()
	form.AddFieldTypeRows(r,
		codeplug.FtGsPwAndLockEnable,
		codeplug.FtGsPowerOnPassword,
	)

	if r.HasFieldType(codeplug.FtGsCHVoiceAnnouncement) {
		groupBox = column.AddGroupbox(lang.T("语音播报", "Voice Announcement"))
		form = groupBox.AddForm()
		form.AddFieldTypeRows(r,
			codeplug.FtGsCHVoiceAnnouncement,
		)
	}
	column.AddFiller()

	column = row.AddVbox()
	form = column.AddForm()
	form.AddFieldTypeRows(r,
		codeplug.FtGsRadioName,
		codeplug.FtGsRadioID,
		codeplug.FtGsMonitorType,
		codeplug.FtGsVoxSensitivity,
		codeplug.FtGsTxPreambleDuration,
		codeplug.FtGsRxLowBatteryInterval,
		codeplug.FtGsChannelsHangTime,
		codeplug.FtGsBacklightColor,
		codeplug.FtGsLockUnlock,
		codeplug.FtGsPcProgPassword,
		codeplug.FtGsRadioProgPassword,
		codeplug.FtGsBacklightTime,
		codeplug.FtGsSetKeypadLockTime,
	)

	groupBox = column.AddGroupbox(lang.T("脱网", "Talkaround"))
	form = groupBox.AddForm()
	form.AddFieldTypeRows(r,
		codeplug.FtGsGroupCallHangTime,
		codeplug.FtGsPrivateCallHangTime,
	)

	column = row.AddVbox()
	form = column.AddForm()
	form.AddFieldTypeRows(r,
		codeplug.FtGsRadioID1,
		codeplug.FtGsRadioID2,
		codeplug.FtGsRadioID3,
		codeplug.FtGsMicLevel,
		codeplug.FtGsTxMode,
		codeplug.FtGsFreqChannelMode,
		codeplug.FtGsModeSelect,
		codeplug.FtGsModeSelectA,
		codeplug.FtGsModeSelectB,
		codeplug.FtGsTimeZone,
		codeplug.FtGsDisableAllLeds,
		codeplug.FtGsGroupCallMatch,
		codeplug.FtGsPrivateCallMatch,
		codeplug.FtGsEditRadioID,
		codeplug.FtGsPublicZone,
		codeplug.FtGsMenuControl,
		codeplug.FtGsTwoChannel,
		codeplug.FtGsEnableContactsCSV,
	)

	groupBox = column.AddGroupbox(lang.T("开机画面", "Intro Screen"))
	form = groupBox.AddForm()
	form.AddFieldTypeRows(r,
		codeplug.FtGsIntroScreen,
		codeplug.FtGsIntroScreenLine1,
		codeplug.FtGsIntroScreenLine2,
	)

	column.AddFiller()
}
