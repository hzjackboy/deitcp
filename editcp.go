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
	"bytes"
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/dalefarnsworth-dmr/codeplug"
	l "github.com/dalefarnsworth-dmr/debug"
	"github.com/dalefarnsworth-dmr/ui"
	"github.com/therecipe/qt/core"
)

// This turns on the experimental checkbox in preferences
const needExperimental = false

const autosaveSuffix = ".autosave"
const maxRecentFiles = 10

var editorOpened = false
var editors []*editor

type editorSettings struct {
	sortAvailableChannels  bool
	sortAvailableChannelsB bool
	sortAvailableContacts  bool
	codeplugDirectory      string
	autosaveInterval       int
	recentFiles            []string
	model                  string
	freqRange              string
	editButtonsEnabled     bool
	gpsEnabled             bool
	experimental           bool
	uniqueContactNames     bool
}

var appSettings *ui.AppSettings
var settings editorSettings

type editor struct {
	app           *ui.App
	codeplug      *codeplug.Codeplug
	mainWindow    *ui.MainWindow
	prefWindow    *ui.Window
	autosaveTimer *core.QTimer
	codeplugHash  [sha256.Size]byte
	codeplugCount int
}

func checkAutosave(filename string) {
	asFilename := filename + autosaveSuffix
	asInfo, err := os.Stat(asFilename)
	if err != nil {
		return
	}

	fileInfo, err := os.Stat(filename)
	if err != nil {
		return
	}

	if fileInfo.ModTime().After(asInfo.ModTime()) {
		os.Remove(asFilename)
		return
	}

	backupFilename := filename + ".backup"
	title := "发现自动保存文件"
	msg := "自动保存的备份 %s 已存在。 " +
		"要从此备份中恢复文件吗？"
	msg = fmt.Sprintf(msg, filename)
	switch ui.YesNoPopup(title, msg) {
	case ui.PopupYes:
		os.Rename(filename, backupFilename)
		os.Rename(asFilename, filename)
		msg := fmt.Sprintf("%s has been saved as %s",
			filename, backupFilename)
		ui.InfoPopup("备份已创建", msg)
	default:
		break
	}
}

func (edt *editor) revertFile() error {
	var err error

	cp := edt.codeplug
	if cp.Changed() {
		title := fmt.Sprintf("Revert %s", cp.Filename())
		msg := cp.Filename() + " has been modified.\n"
		msg += "确定要放弃更改吗？"
		switch ui.YesNoPopup(title, msg) {
		case ui.PopupYes:
			err := edt.codeplug.Revert()
			if err != nil {
				ui.ErrorPopup("还原失败", err.Error())
			}
			edt.updateMenuBar()
			edt.mainWindow.CodeplugChanged(nil)

		default:
			break
		}
	}

	return err
}

func (edt *editor) save() string {
	cp := edt.codeplug
	if cp.Filename() == "." || cp.FileType() != codeplug.FileTypeRdt {
		return edt.saveAs("")
	}
	return edt.saveAs(edt.codeplug.Filename())
}

func baseFilename(filename string) string {
	base := filepath.Base(filename)
	ext := filepath.Ext(base)
	if ext != "" {
		base = strings.TrimSuffix(base, ext)
	}

	return base
}

func (edt *editor) saveAs(filename string) string {
	cp := edt.codeplug
	if filename == "" {
		dir := settings.codeplugDirectory
		base := baseFilename(edt.codeplug.Filename())
		ext := cp.Ext()
		dir = filepath.Join(dir, base+"."+ext)
		filename = ui.SaveFilename("保存写频文件", dir, ext)
		if filename == "" {
			return ""
		}
		settings.codeplugDirectory = filepath.Dir(filename)
		saveSettings()
	}

	valid := cp.Valid()
	edt.updateMenuBar()
	if !valid {
		fmtStr := `
%d records with invalid field values were found in the codeplug.

Click on Cancel and then select "Menu->Edit->Show Invalid Fields" to view them.

Or, click on Ignore to continue saving the file.`
		msg := fmt.Sprintf(fmtStr, len(cp.Warnings()))
		title := fmt.Sprintf("%s: save warning", filename)
		rv := ui.WarningPopup(title, msg)
		if rv != ui.PopupIgnore {
			return ""
		}
	}

	err := cp.SaveAs(filename)
	if err != nil {
		title := fmt.Sprintf("%s: save failed", filename)
		ui.ErrorPopup(title, err.Error())
		return ""
	}

	edt.updateFilename()

	autosaveFilename := cp.Filename() + autosaveSuffix
	os.Remove(autosaveFilename)
	return filename
}

func (edt *editor) setGPSEnabled(gpsEnabled bool) {
	edt.updateButtons()
	w := edt.mainWindow.RecordWindows()[codeplug.RtChannels_md380]
	if w != nil {
		w.RecordFunc()()
	}
}

func (edt *editor) setAutosaveInterval(seconds int) {
	if seconds == 0 {
		edt.autosaveTimer.Stop()
		return
	}
	if edt.autosaveTimer == nil {
		edt.autosaveTimer = core.NewQTimer(nil)
		edt.autosaveTimer.ConnectTimeout(func() {
			edt.autosave()
		})
	}
	edt.autosaveTimer.Start(seconds * 60 * 1000)
}

func (edt *editor) autosave() {
	cp := edt.codeplug
	if cp == nil {
		return
	}

	filename := cp.Filename() + autosaveSuffix

	hash := cp.CurrentHash()
	if hash == edt.codeplugHash {
		return
	}
	edt.codeplugHash = hash

	err := cp.SaveToFile(filename)
	if err != nil {
		os.Remove(filename)
	}
}

func displayPreviousPanic(text string) {
	var removeFile bool

	mw := ui.NewMainWindow()
	mw.Resize(600, 400)
	mw.SetTitle("上次崩溃信息")
	mw.ConnectClose(func() bool {
		if removeFile {
			l.RemovePreviousPanicFile()
		}

		return true
	})

	vBox := mw.AddVbox()
	vBox.AddSpace(1)
	vBox.AddLabel("editcp 上次崩溃时留下以下信息：")

	tEdit := vBox.AddTextEdit()
	tEdit.SetReadOnly(true)
	tEdit.SetNoLineWrap()
	tEdit.SetPlainText(text)

	label := fmt.Sprintf("(This message is contained in the file '%s')", l.PreviousPanicFilename)
	vBox.AddLabel(label)

	form := vBox.AddForm()
	cb := ui.NewCheckboxWidget(false, func(checked bool) {
		removeFile = checked
	})
	cb.SetLabel("删除此信息？")
	form.AddWidget(cb)

	hBox := vBox.AddHbox()
	hBox.AddFiller()
	hBox.SetFixedHeight()

	button := ui.NewButtonWidget("关闭", func() {
		mw.Close()
	})
	hBox.AddWidget(button)

	mw.Show()
}

func main() {
	platformInit()
	l.WindowsSaveDebuggingInfo()
	args := os.Args[1:]
	for i := len(args) - 1; i >= 0; i-- {
		switch args[i] {
		case "--experimental":
			settings.experimental = true
			args = append(args[:i], args[i+1:]...)
		}
	}

	app, err := ui.NewApp()
	if err != nil {
		l.Println(err.Error())
		return
	}
	app.SetOrganizationName("codeplug")
	app.SetApplicationName("写频编辑器")
	appSettings = app.NewSettings()
	loadSettings()

	filenames := args
	if len(filenames) == 0 {
		filenames = []string{""}
	}

	for _, filename := range filenames {
		newEditor(app, codeplug.FileTypeNone, filename)
	}

	panicString := l.PreviousPanicString()
	if len(panicString) > 0 {
		displayPreviousPanic(panicString)
	}

	if len(editors) == 0 && len(panicString) == 0 {
		return
	}

	app.Exec()

	saveSettings()
}

func (edt *editor) titleSuffix() string {
	suffix := ""
	if edt.codeplugCount > 1 {
		suffix = fmt.Sprintf(" #%d", edt.codeplugCount)
	}
	return suffix
}

func deleteEditor(i int) {
	copy(editors[i:], editors[i+1:])
	editors[len(editors)-1] = nil
	editors = editors[:len(editors)-1]
}

func deleteString(strs *[]string, i int) {
	copy((*strs)[i:], (*strs)[i+1:])
	(*strs)[len(*strs)-1] = ""
	*strs = (*strs)[:len(*strs)-1]
}

func (edt *editor) openCodeplug(fType codeplug.FileType, filename string) {
	if fType == codeplug.FileTypeNone {
		if absPath, err := filepath.Abs(filename); err == nil {
			filename = absPath
		}
		fileInfo, err := os.Stat(filename)
		if err != nil {
			ui.ErrorPopup(filename, err.Error())
			removeRecentFile(filename)
			return
		}

		for _, cp := range codeplug.Codeplugs() {
			xfileInfo, err := os.Stat(cp.Filename())
			if err == nil && os.SameFile(xfileInfo, fileInfo) {
				edt.codeplug = cp
				break
			}
		}
	}

	if edt.codeplug == nil {
		checkAutosave(filename)

		cp, err := codeplug.NewCodeplug(fType, filename)
		if err != nil {
			ui.ErrorPopup("写频文件错误", err.Error())
			return
		}

		cp.SetUniqueContactNames(settings.uniqueContactNames)
		cp.SetGPSEnabled(settings.gpsEnabled)

		typ, freqRange := typeFrequencyRange(cp)

		if typ == "" || freqRange == "" {
			return
		}

		err = cp.Load(typ, freqRange)
		if err != nil {
			ui.ErrorPopup("写频数据加载错误", err.Error())
			return
		}
		if !cp.Valid() {
			fmtStr := `
%d records with invalid field values were found in the codeplug.

Select "Menu->Edit->Show Invalid Fields" to view them.`
			msg := fmt.Sprintf(fmtStr, len(cp.Warnings()))
			ui.InfoPopup("写频警告", msg)
		}
		edt.updateMenuBar()

		edt.codeplug = cp
		edt.codeplugHash = edt.codeplug.CurrentHash()
		loadSettings()
		edt.setAutosaveInterval(settings.autosaveInterval)
	}

	if fType == codeplug.FileTypeNone {
		addRecentFile(filename)
	}

	highCount := 0
	cp := edt.codeplug
	for _, edt := range editors {
		if edt.codeplug == cp && edt.codeplugCount > highCount {
			highCount = edt.codeplugCount
		}
	}
	edt.codeplugCount = highCount + 1
}

func (edt *editor) FreeCodeplug() {
	if edt.codeplug == nil {
		return
	}
	edt.codeplug.Free()
	if len(editors) > 1 {
		edt.mainWindow.Close()
		return
	}
	edt.codeplug = nil
	edt.codeplugCount--
	edt.updateFilename()
	edt.updateMenuBar()
	edt.updateButtons()
}

func typeFreqRanges(cp *codeplug.Codeplug, typ string) (rangesA, rangesB []string) {
	_, freqRangesMap := cp.TypesFrequencyRanges()
	rangeMapA := make(map[string]bool)
	rangeMapB := make(map[string]bool)

	freqRanges := freqRangesMap[typ]
	for _, r := range freqRanges {
		ranges := strings.Split(r, "_")
		rangeMapA[ranges[0]+" MHz"] = true
		if len(ranges) > 1 {
			rangeMapB[ranges[1]+" MHz"] = true
		}
	}

	rangesA = make([]string, 0)
	for r := range rangeMapA {
		rangesA = append(rangesA, r)
	}
	sort.Strings(rangesA)

	rangesB = make([]string, 0)
	for r := range rangeMapB {
		rangesB = append(rangesB, r)
	}
	sort.Strings(rangesB)

	return rangesA, rangesB
}

func typeFrequencyRange(cp *codeplug.Codeplug) (typ string, freqRange string) {
	types, freqRangesMap := cp.TypesFrequencyRanges()
	if len(types) == 1 {
		typ = types[0]
		ranges := freqRangesMap[typ]
		if ranges != nil && len(ranges) == 1 {
			return typ, ranges[0]

		}
	}

	typ = settings.model

	mOpts := append([]string{"<选择型号>"}, types...)

	var vOptsA []string
	var rangeA string
	var rangeB string
	var rangesA = make([]string, 0)
	var rangesB = make([]string, 0)

	rangesA, rangesB = typeFreqRanges(cp, typ)
	settingRanges := strings.Split(settings.freqRange, "_")

	rangeA = settingRanges[0] + " MHz"
	if len(rangesB) == 0 {
		vOptsA = append([]string{"<选择频率范围>"}, rangesA...)
	} else {
		vOptsA = append([]string{"<选择频率范围 A>"}, rangesA...)
	}
	if len(settingRanges) > 1 {
		rangeB = settingRanges[1] + " MHz"
	}
	vOptsB := append([]string{"<选择频率范围 B>"}, rangesB...)

	dialog := ui.NewDialog("选择写频类型")

	cancelButton := ui.NewButtonWidget("取消", func() {
		dialog.Reject()
	})
	okButton := ui.NewButtonWidget("确定", func() {
		dialog.Accept()
	})
	opt := vOptsA[0]
	enable := containsString(rangeA, vOptsA[1:])
	if enable {
		opt = rangeA
	}
	if len(rangesB) != 0 {
		enable = enable && containsString(rangeB, vOptsB[1:])
	}
	okButton.SetEnabled(enable)

	vCbA := ui.NewComboboxWidget(opt, vOptsA, func(index int) {
		if index < 0 {
			return
		}
		rangeA = vOptsA[index]
		enable := containsString(rangeA, vOptsA[1:])
		rangesA, rangesB = typeFreqRanges(cp, typ)
		if len(rangesB) != 0 {
			enable = enable && containsString(rangeB, rangesB)
		}
		okButton.SetEnabled(enable)
	})
	vCbA.SetEnabled(containsString(typ, mOpts[1:]))

	opt = vOptsB[0]
	if containsString(rangeB, vOptsB[1:]) {
		opt = rangeB
	}

	vCbB := ui.NewComboboxWidget(opt, vOptsB, func(index int) {
		if index < 0 {
			return
		}
		rangeB = vOptsB[index]
		enable := containsString(rangeA, vOptsA[1:])
		rangesA, rangesB = typeFreqRanges(cp, typ)
		if len(rangesB) != 0 {
			enable = enable && containsString(rangeB, rangesB)
		}
		okButton.SetEnabled(enable)
	})
	vCbB.SetEnabled(containsString(typ, mOpts[1:]))

	if len(types) == 1 {
		mOpts = types
	}

	var form *ui.Form
	var mCb *ui.FieldWidget

	mCb = ui.NewComboboxWidget(typ, mOpts, func(index int) {
		if index < 0 {
			return
		}
		typ = mOpts[index]

		rangesA, rangesB = typeFreqRanges(cp, typ)
		settingRanges := strings.Split(settings.freqRange, "_")
		rangeA = settingRanges[0] + " MHz"
		if len(rangesB) == 0 {
			vOptsA = append([]string{"<选择频率范围>"}, rangesA...)
		} else {
			vOptsA = append([]string{"<选择频率范围 A>"}, rangesA...)
		}
		if len(settingRanges) > 1 {
			rangeB = settingRanges[1] + " MHz"
		}
		vOptsB = append([]string{"<选择频率范围 B>"}, rangesB...)
		vCbA.SetEnabled(containsString(typ, mOpts[1:]))

		opt := vOptsA[0]
		enable := containsString(rangeA, vOptsA[1:])
		if enable {
			opt = rangeA
		}
		if len(rangesB) > 1 {
			enable = enable && containsString(rangeB, vOptsB[1:])
		}
		okButton.SetEnabled(enable)

		ui.UpdateComboboxWidget(vCbA, opt, vOptsA)

		vCbA.SetLabel("")
		if len(rangesB) > 1 {
			vCbA.SetLabel("A")
		}

		opt = vOptsB[0]
		if containsString(rangeB, vOptsB[1:]) {
			opt = rangeB
		}

		if len(rangesB) > 1 {
			vCbB.SetEnabled(containsString(typ, mOpts[1:]))
			ui.UpdateComboboxWidget(vCbB, opt, vOptsB)
			vCbB.SetVisible(true)
		} else {
			vCbB.SetVisible(false)
		}
	})

	dialog.AddLabel("选择写频的型号和频率范围。")
	form = dialog.AddForm()
	form.AddRow("", mCb)
	form.AddRow("", vCbA)
	if len(rangesB) > 1 {
		vCbA.SetLabel("A")
	}
	form.AddRow("B", vCbB)
	vCbB.SetVisible(len(rangesB) > 1)

	row := dialog.AddHbox()
	row.AddWidget(cancelButton)
	row.AddWidget(okButton)

	if !dialog.Exec() {
		return "", ""
	}

	freqRange = rangeA
	if len(rangesB) > 1 {
		freqRange += "_" + rangeB
	}
	freqRange = strings.Replace(freqRange, " MHz", "", -1)

	if containsString(typ, types) {
		settings.model = typ
		settings.freqRange = freqRange
	}
	saveSettings()

	return typ, freqRange
}

func containsString(str string, strs []string) bool {
	found := false
	for _, s := range strs {
		if s == str {
			found = true
		}
	}
	return found
}

func newEditor(app *ui.App, fType codeplug.FileType, filename string) *editor {
	var edt *editor
	for _, ed := range editors {
		if ed.codeplug == nil {
			edt = ed
			break
		}
	}

	if edt == nil {
		edt = new(editor)
		edt.app = app
		editors = append(editors, edt)
	}

	mw := edt.mainWindow
	if mw == nil {
		mw = ui.NewMainWindow()
		edt.mainWindow = mw

		mw.Resize(600, 50)
	}

	if filename != "" || fType != codeplug.FileTypeNone {
		edt.openCodeplug(fType, filename)
	}

	cp := edt.codeplug
	if cp != nil {
		mw.SetCodeplug(cp)
		cp.SetUniqueContactNames(settings.uniqueContactNames)
		cp.SetGPSEnabled(settings.gpsEnabled)
	}

	edt.updateFilename()

	mw.ConnectClose(func() bool {
		if cp != nil {
			count := 0
			for _, edt := range editors {
				if edt.codeplug == cp {
					count++
				}
			}

			if count == 1 && cp.Changed() {
				title := fmt.Sprintf("保存 %s", cp.Filename())
				msg := cp.Filename() + " has been modified.\n"
				msg += "是否保存更改？"
				switch ui.SavePopup(title, msg) {
				case ui.PopupSave:
					if edt.save() == "" {
						return false
					}

				case ui.PopupDiscard:
					break

				case ui.PopupCancel:
					return false
				}
			}
		}

		for i, editor := range editors {
			if editor == edt {
				deleteEditor(i)
				break
			}
		}

		if cp != nil {
			asFilename := cp.Filename() + autosaveSuffix
			os.Remove(asFilename)
		}
		return true
	})

	edt.updateMenuBar()

	edt.updateButtons()

	mw.Show()

	if len(editors) > 1 && cp == nil {
		mw.Close()
		return nil
	}

	editorOpened = true
	return edt
}

func (edt *editor) updateMenuBar() {
	cp := edt.codeplug
	mb := edt.mainWindow.MenuBar()
	mb.Clear()
	menu := mb.AddMenu("文件")
	menu.AddAction("新建...", func() {
		newEditor(edt.app, codeplug.FileTypeNew, "")
	})
	menu.AddAction("打开...", func() {
		dir := settings.codeplugDirectory
		exts := edt.codeplug.AllExts()
		filenames := ui.OpenCPFilenames("打开写频文件", dir, exts)
		for _, filename := range filenames {
			if filename != "" {
				newEditor(edt.app, codeplug.FileTypeNone, filename)
			}
		}
	})
	recentMenu := menu.AddMenu("最近打开...")
	recentMenu.ConnectAboutToShow(func() {
		edt.updateRecentMenu(recentMenu)
	})
	recentMenu.SetEnabled(len(settings.recentFiles) != 0)

	menu.AddAction("还原", func() {
		edt.revertFile()
	}).SetEnabled(cp != nil)

	menu.AddSeparator()

	menu.AddAction("转换写频到新型号...", func() {
		edt.convertCodeplug()
	}).SetEnabled(cp != nil)

	menu.AddSeparator()

	importMenu := menu.AddMenu("导入...")
	importMenu.AddAction("导入文本文件...", func() {
		edt.importText()
	})

	importMenu.AddAction("导入表格文件...", func() {
		edt.importXLSX()
	})

	importMenu.AddAction("导入JSON文件...", func() {
		edt.importJSON()
	})

	exportMenu := menu.AddMenu("导出...")
	exportMenu.SetEnabled(cp != nil)

	exportMenu.AddAction("导出为文本...", func() {
		edt.exportText()
	})

	exportMenu.AddAction("导出为文本(每条记录一行)...", func() {
		edt.exportTextOneLineRecords()
	})

	exportMenu.AddAction("导出为表格...", func() {
		edt.exportXLSX()
	})

	exportMenu.AddAction("导出为JSON...", func() {
		edt.exportJSON()
	})

	menu.AddSeparator()

	menu.AddAction("保存", func() {
		edt.save()
	}).SetEnabled(cp != nil)

	menu.AddAction("另存为...", func() {
		edt.saveAs("")
	}).SetEnabled(cp != nil)

	menu.AddSeparator()

	menu.AddAction("关闭", func() {
		edt.mainWindow.Close()
	})

	menu.AddAction("退出", func() {
		for i := len(editors) - 1; i >= 0; i-- {
			editors[i].mainWindow.Close()
		}
	})

	var showInvalidAction *ui.Action
	menu = mb.AddMenu("编辑")
	menu.ConnectAboutToShow(func() {
		showInvalidAction.SetEnabled(cp != nil && len(cp.Warnings()) != 0)
	})
	menu.AddAction("基本信息", func() {
		basicInformation(edt)
	}).SetEnabled(cp != nil)

	menu.AddAction("常规设置", func() {
		generalSettings(edt)
	}).SetEnabled(cp != nil)

	menu.AddAction("菜单项", func() {
		menuItems(edt)
	}).SetEnabled(cp != nil)

	menu.AddAction("按键定义", func() {
		buttonDefinitions(edt)
	}).SetEnabled(cp != nil)

	menu.AddAction("短信", func() {
		textMessages(edt)
	}).SetEnabled(cp != nil)

	menu.AddAction("隐私设置", func() {
		privacySettings(edt)
	}).SetEnabled(cp != nil)

	menu.AddAction("信道", func() {
		channels(edt)
	}).SetEnabled(cp != nil)

	menu.AddAction("联系人", func() {
		contacts(edt)
	}).SetEnabled(cp != nil)

	menu.AddAction("接收组列表", func() {
		groupLists(edt)
	}).SetEnabled(cp != nil)

	menu.AddAction("扫描列表", func() {
		scanLists(edt)
	}).SetEnabled(cp != nil)

	menu.AddAction("区域", func() {
		zones(edt)
	}).SetEnabled(cp != nil)

	if cp != nil && cp.HasRecordType(codeplug.RtGPSSystems) {
		menu.AddAction("GPS系统", func() {
			gpsSystems(edt)
		}).SetEnabled(cp != nil && settings.gpsEnabled)
	}

	menu.AddSeparator()

	showInvalidAction = menu.AddAction("显示无效字段", func() {
		checkCodeplug(edt)
	})
	showInvalidAction.SetEnabled(cp != nil && len(cp.Warnings()) != 0)

	menu.AddSeparator()

	menu.AddAction("偏好设置...", func() {
		edt.preferences()
	})

	edt.addRadioMenu(menu)

	windowsMenu := mb.AddMenu("窗口")
	windowsMenu.ConnectAboutToShow(func() {
		edt.updateWindowsMenu(windowsMenu)
	})

	menu = mb.AddMenu("帮助")
	menu.AddAction("关于...", func() {
		about()
	})
	menu.AddAction("致谢...", func() {
		thanks()
	})
}

func (edt *editor) updateButtons() {
	cp := edt.codeplug

	row := edt.mainWindow.AddHbox()
	row.Clear()

	if !settings.editButtonsEnabled {
		return
	}

	column := row.AddVbox()

	biButton := column.AddButton("基本信息")
	biButton.SetEnabled(cp != nil)
	biButton.ConnectClicked(func() { basicInformation(edt) })

	gsButton := column.AddButton("常规设置")
	gsButton.SetEnabled(cp != nil)
	gsButton.ConnectClicked(func() { generalSettings(edt) })

	miButton := column.AddButton("菜单项")
	miButton.SetEnabled(cp != nil)
	miButton.ConnectClicked(func() { menuItems(edt) })

	bdButton := column.AddButton("按键定义")
	bdButton.SetEnabled(cp != nil)
	bdButton.ConnectClicked(func() { buttonDefinitions(edt) })

	tmButton := column.AddButton("短信")
	tmButton.SetEnabled(cp != nil)
	tmButton.ConnectClicked(func() { textMessages(edt) })

	psButton := column.AddButton("隐私设置")
	psButton.SetEnabled(cp != nil)
	psButton.ConnectClicked(func() { privacySettings(edt) })

	ciButton := column.AddButton("信道")
	ciButton.SetEnabled(cp != nil)
	ciButton.ConnectClicked(func() { channels(edt) })

	dcButton := column.AddButton("联系人")
	dcButton.SetEnabled(cp != nil)
	dcButton.ConnectClicked(func() { contacts(edt) })

	glButton := column.AddButton("接收组列表")
	glButton.SetEnabled(cp != nil)
	glButton.ConnectClicked(func() { groupLists(edt) })

	slButton := column.AddButton("扫描列表")
	slButton.SetEnabled(cp != nil)
	slButton.ConnectClicked(func() { scanLists(edt) })

	ziButton := column.AddButton("区域")
	ziButton.SetEnabled(cp != nil)
	ziButton.ConnectClicked(func() { zones(edt) })

	if cp != nil && cp.HasRecordType(codeplug.RtGPSSystems) {
		gpButton := column.AddButton("GPS系统")
		gpButton.SetEnabled(cp != nil && settings.gpsEnabled)
		gpButton.ConnectClicked(func() { gpsSystems(edt) })
	}

	row.AddSeparator()

	column = row.AddVbox()

	column.AddFiller()
	row.AddFiller()
}

func (edt *editor) updateFilename() {
	title := "写频编辑器"
	cp := edt.codeplug
	if cp != nil {
		filename := cp.Filename()
		title = filename + edt.titleSuffix()
		if _, err := os.Stat(filename); err == nil {
			settings.codeplugDirectory = filepath.Dir(filename)
			saveSettings()
		}
		addRecentFile(filename)
	}

	edt.mainWindow.SetTitle(title)
}

func (edt *editor) updateWindowsMenu(menu *ui.Menu) {
	menu.Clear()

	mainWindows := ui.MainWindows()
	sort.Slice(mainWindows, func(i, j int) bool {
		return mainWindows[i].Title() < mainWindows[j].Title()
	})
	for i := range mainWindows {
		mw := mainWindows[i]
		menu.AddAction(mw.Title(), func() {
			mw.Show()
		})
		windows := make([]*ui.Window, 0, 16)
		for _, w := range mw.RecordWindows() {
			windows = append(windows, w)
		}
		sort.Slice(windows, func(i, j int) bool {
			return windows[i].Title() < windows[j].Title()
		})
		for i := range windows {
			w := windows[i]
			menu.AddAction(w.Title(), func() {
				w.Show()
			})
		}
	}
}

func (edt *editor) updateRecentMenu(menu *ui.Menu) {
	menu.Clear()

	loadSettings()

	for i := range settings.recentFiles {
		filename := settings.recentFiles[i]
		menu.AddAction(filename, func() {
			newEditor(edt.app, codeplug.FileTypeNone, filename)
		})
	}
	menu.SetEnabled(len(settings.recentFiles) != 0)
}

func addRecentFile(name string) {
	if _, err := os.Stat(name); err != nil {
		return
	}

	if len(settings.recentFiles) > 0 && name == settings.recentFiles[0] {
		return
	}

	removeRecentFile(name)

	settings.recentFiles = append([]string{name}, settings.recentFiles...)

	if len(settings.recentFiles) > maxRecentFiles {
		settings.recentFiles = settings.recentFiles[:maxRecentFiles]
	}

	saveSettings()
}

func removeRecentFile(name string) {
	for i, n := range settings.recentFiles {
		if n == name {
			deleteString(&settings.recentFiles, i)
			break
		}
	}
	saveSettings()
}

func (edt *editor) exportText() {
	dir := settings.codeplugDirectory
	base := baseFilename(edt.codeplug.Filename())
	ext := "txt"
	dir = filepath.Join(dir, base+"."+ext)
	filename := ui.SaveFilename("导出为文本文件", dir, ext)
	if filename == "" {
		return
	}
	settings.codeplugDirectory = filepath.Dir(filename)
	saveSettings()

	err := edt.codeplug.ExportText(filename)
	if err != nil {
		title := fmt.Sprintf("Export to %s", filename)
		ui.ErrorPopup(title, err.Error())
		return
	}
}

func (edt *editor) exportTextOneLineRecords() {
	dir := settings.codeplugDirectory
	base := baseFilename(edt.codeplug.Filename())
	ext := "txt"
	dir = filepath.Join(dir, base+"."+ext)
	filename := ui.SaveFilename("导出为文本文件", dir, ext)
	if filename == "" {
		return
	}
	settings.codeplugDirectory = filepath.Dir(filename)
	saveSettings()

	err := edt.codeplug.ExportTextOneLineRecords(filename)
	if err != nil {
		title := fmt.Sprintf("Export to %s", filename)
		ui.ErrorPopup(title, err.Error())
		return
	}
}

func (edt *editor) convertCodeplug() {
	body := edt.codeplug.TextLines()
	body = body[1:]

	edt = newEditor(edt.app, codeplug.FileTypeNew, "")
	if edt == nil {
		return
	}
	cp := edt.codeplug
	header := cp.TextLines()[:1]

	cp.RemoveAllRecords()

	body = append(header, body...)
	text := strings.Join(body, "")

	reader := bytes.NewReader([]byte(text))
	cp.ImportText(reader)

	if !cp.Valid() {
		fmtStr := `
%d records with invalid field values were found in the codeplug.

Select "Menu->Edit->Show Invalid Fields" to view them.`
		msg := fmt.Sprintf(fmtStr, len(cp.Warnings()))
		ui.InfoPopup("写频警告", msg)
	}
	edt.updateMenuBar()
}

func (edt *editor) importText() {
	dir := settings.codeplugDirectory
	filename := ui.OpenTextFilename("导入文本文件", dir)
	if filename == "" {
		return
	}
	settings.codeplugDirectory = filepath.Dir(filename)
	saveSettings()

	newEditor(edt.app, codeplug.FileTypeText, filename)
}

func (edt *editor) importXLSX() {
	dir := settings.codeplugDirectory
	filename := ui.OpenXLSXFilename("导入表格文件", dir)
	if filename == "" {
		return
	}
	settings.codeplugDirectory = filepath.Dir(filename)
	saveSettings()

	newEditor(edt.app, codeplug.FileTypeXLSX, filename)
}

func (edt *editor) exportXLSX() {
	dir := settings.codeplugDirectory
	base := baseFilename(edt.codeplug.Filename())
	ext := "xlsx"
	dir = filepath.Join(dir, base+"."+ext)
	filename := ui.SaveFilename("导出为表格文件", dir, ext)
	if filename == "" {
		return
	}
	settings.codeplugDirectory = filepath.Dir(filename)
	saveSettings()

	err := edt.codeplug.ExportXLSX(filename)
	if err != nil {
		title := fmt.Sprintf("Export to %s", filename)
		ui.ErrorPopup(title, err.Error())
		return
	}
}

func (edt *editor) importJSON() {
	dir := settings.codeplugDirectory
	filename := ui.OpenJSONFilename("导入JSON文件", dir)
	if filename == "" {
		return
	}
	settings.codeplugDirectory = filepath.Dir(filename)
	saveSettings()

	newEditor(edt.app, codeplug.FileTypeJSON, filename)
}

func (edt *editor) exportJSON() {
	dir := settings.codeplugDirectory
	base := baseFilename(edt.codeplug.Filename())
	ext := "json"
	dir = filepath.Join(dir, base+"."+ext)
	filename := ui.SaveFilename("导出为JSON文件", dir, ext)
	if filename == "" {
		return
	}
	settings.codeplugDirectory = filepath.Dir(filename)
	saveSettings()

	err := edt.codeplug.ExportJSON(filename)
	if err != nil {
		title := fmt.Sprintf("Export to %s", filename)
		ui.ErrorPopup(title, err.Error())
		return
	}
}

func about() {
	msg := fmt.Sprintf("editcp Version %s\n", version)
	msg += `
editcp is free software licensed
under version 3 of the GPL.

Copyright 2017-2021 Dale Farnsworth.  All rights reserved.

Dale Farnsworth
1007 W Mendoza Ave
Mesa, AZ  85210
USA

dale@farnsworth.org

The source code for editcp may be found at
https://github.com/dalefarnsworth-dmr/editcp
`
	ui.InfoPopup("关于 editcp", msg)
}

func thanks() {
	msgs := []string{
		"衷心感谢以下人士：",
		"  José Melo, CT4TX, for creating the nice logo",
		"  Ron McMurdy, W5QLD, for reporting bugs",
		"  Markus Lenggenhager, HB9BRJ, for reporting bugs",
		"  Roy G. Jackson, KW4G, for reporting bugs",
		"  Kevin Otte, N8VNR, for reporting bugs",
		"  Andreas Krüger, DJ3EI, for reporting bugs",
		"  Martin Jones, KI0KO, for reporting bugs",
		"  Marco Carrara, IW2KWD, for suggesting improvements",
		"  Bob Finch, W9YA, for reporting bugs",
		"  Tyler Tidman, VA3DGN, for reporting bugs + RT90 & MD-9600 support",
		"",
		"I apologize to the many contributors whom I've missed.",
		"If you've reported a fix, a bug or suggestion, let me",
		"know and I'll add you to this list.",
		"Dale Farnsworth, NO7K, dale@farnsworth.org",
	}

	msg := strings.Join(msgs, "\n")
	ui.InfoPopup("致谢", msg)
}

type fillRecord func(*editor, *ui.HBox)

func (edt *editor) recordWindow(rType codeplug.RecordType) *ui.Window {
	windows := edt.mainWindow.RecordWindows()
	return windows[rType]
}

func (edt *editor) newRecordWindow(rType codeplug.RecordType, writable bool, fillRecord fillRecord) {
	windows := edt.mainWindow.RecordWindows()
	w := windows[rType]
	if w != nil {
		w.Show()
		return
	}

	w = edt.mainWindow.NewRecordWindow(rType, writable)
	windows[rType] = w

	w.ConnectClose(func() bool {
		delete(windows, rType)
		return true
	})

	cp := edt.codeplug
	r := cp.Record(rType)
	w.SetTitle(cp.Filename() + edt.titleSuffix() + " " + r.TypeName())

	windowBox := w.AddHbox()
	var recordFunc func()

	if cp.MaxRecords(rType) == 1 {
		selectorBox := windowBox.AddVbox()
		recordFunc = func() {
			selectorBox.Clear()
			recordBox := selectorBox.AddHbox()
			fillRecord(edt, recordBox)
			w.Show()
		}
	} else {
		windowBox.AddRecordList(rType)
		selectorBox := windowBox.AddVbox()
		recordFunc = func() {
			selectorBox.Clear()
			recordBox := selectorBox.AddHbox()
			fillRecord(edt, recordBox)
			addRecordSelector(selectorBox, writable)
			w.Show()
		}
	}

	w.SetRecordFunc(recordFunc)
	recordFunc()

}

func addRecordSelector(box *ui.VBox, writable bool) {
	w := box.Window()
	cp := w.MainWindow().Codeplug()
	rl := w.RecordList()
	rType := w.RecordType()
	row := box.AddHbox()
	row.SetFixedHeight()

	decrement := row.AddButton("<")
	decrement.ConnectClicked(func() {
		rIndex := rl.Current()
		if rIndex <= 0 {
			return
		}

		rIndex--
		rl.SetCurrent(rIndex)
	})

	rIndex := rl.Current()
	records := cp.Records(rType)
	row.AddButton(fmt.Sprintf("%d of %d", rIndex+1, len(records)))
	increment := row.AddButton(">")
	increment.ConnectClicked(func() {
		rIndex := rl.Current()
		records := cp.Records(rType)
		if rIndex >= len(records)-1 {
			return
		}

		rIndex++
		rl.SetCurrent(rIndex)
	})

	if writable {
		row.AddSpace(3)
		add := row.AddButton("添加")
		add.ConnectClicked(func() {
			err := rl.AddSelected()
			if err != nil {
				ui.ErrorPopup("添加记录", err.Error())
				return
			}
		})

		dup := row.AddButton("复制")
		dup.ConnectClicked(func() {
			err := rl.DupSelected()
			if err != nil {
				ui.ErrorPopup("复制记录", err.Error())
				return
			}
		})

		row.AddSpace(3)
		delete := row.AddButton("删除")
		delete.ConnectClicked(func() {
			err := rl.RemoveSelected()
			if err != nil {
				ui.ErrorPopup("删除记录", err.Error())
				return
			}
		})
	}

	row.AddFiller()
}

func currentRecord(w *ui.Window) *codeplug.Record {
	rIndex := 0
	rl := w.RecordList()
	if rl != nil {
		rIndex = rl.Current()
	}
	records := w.MainWindow().Codeplug().Records(w.RecordType())
	if rIndex >= len(records) {
		rIndex = len(records) - 1
	}

	return records[rIndex]
}

func loadSettings() {
	as := appSettings
	as.Sync()
	settings.sortAvailableChannels = as.Bool("sortAvailableChannels", false)
	settings.sortAvailableChannelsB = as.Bool("sortAvailableChannelsB", false)
	settings.sortAvailableContacts = as.Bool("sortAvailableContacts", false)
	settings.codeplugDirectory = as.String("codeplugDirectory", "")
	settings.autosaveInterval = as.Int("autosaveInterval", 1)
	settings.model = as.String("model", "")
	settings.freqRange = as.String("frequencyRange", "")
	settings.editButtonsEnabled = as.Bool("editButtonsEnabled", true)
	settings.gpsEnabled = as.Bool("displayGPS", true)
	settings.experimental = as.Bool("experimental", false)
	settings.uniqueContactNames = as.Bool("uniqueContactNames", true)

	size := as.BeginReadArray("recentFiles")
	settings.recentFiles = make([]string, size)
	for i := 0; i < size; i++ {
		as.SetArrayIndex(i)
		settings.recentFiles[i] = as.String("filename", "")
	}
	as.EndArray()
}

func saveSettings() {
	as := appSettings
	as.SetBool("sortAvailableChannels", settings.sortAvailableChannels)
	as.SetBool("sortAvailableChannelsB", settings.sortAvailableChannelsB)
	as.SetBool("sortAvailableContacts", settings.sortAvailableContacts)
	as.SetString("codeplugDirectory", settings.codeplugDirectory)
	as.SetInt("autosaveInterval", settings.autosaveInterval)
	as.SetString("model", settings.model)
	as.SetString("frequencyRange", settings.freqRange)
	as.SetBool("editButtonsEnabled", settings.editButtonsEnabled)
	as.SetBool("displayGPS", settings.gpsEnabled)
	as.SetBool("experimental", settings.experimental)
	as.SetBool("uniqueContactNames", settings.uniqueContactNames)

	as.BeginWriteArray("recentFiles", len(settings.recentFiles))
	for i, name := range settings.recentFiles {
		as.SetArrayIndex(i)
		as.SetString("filename", name)
	}
	as.EndArray()

	as.Sync()
}
