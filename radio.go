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
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/dalefarnsworth-dmr/codeplug"
	l "github.com/dalefarnsworth-dmr/debug"
	"github.com/dalefarnsworth-dmr/dfu"
	"github.com/dalefarnsworth-dmr/ui"
	"github.com/dalefarnsworth-dmr/userdb"
	"github.com/therecipe/qt/core"
)

type modelURL struct {
	model string
	url   string
}

func writeMD380toolsUsers() {
	title := "写入用户数据库到对讲机"
	text := `
	The users database contains DMR ID numbers and callsigns of all registered
	users. It can only be be written to radios that have been upgraded to the
	md380tools firmware.  See https://github.com/travisgoodspeed/md380tools.

	WARNING: Corruption may occur if a signal is received while writing to the
	radio.  The radio should be tuned to an unprogrammed (or at least quiet)
	channel while writing the new user database.`

	cancel, download := userdbDialog(title, text)
	if cancel {
		return
	}

	locType := core.QStandardPaths__CacheLocation
	cacheDir := core.QStandardPaths_WritableLocation(locType)
	tmpFilename := filepath.Join(cacheDir, "users.tmp")

	msgs := []string{
		"正在从网站下载用户数据库...",
		"正在擦除对讲机用户数据库...",
		"正在写入用户数据库到对讲机...",
	}
	msgIndex := 0
	if !download {
		msgIndex = 1
	}

	filename := userdbFilename()
	os.MkdirAll(filepath.Dir(filename), os.ModeDir|0755)

	pd := ui.NewProgressDialog(msgs[msgIndex])
	pd.SetRange(userdb.MinProgress, userdb.MaxProgress)

	if download {
		db, err := userdb.New(userdb.CuratedUsers(), userdb.Abbreviate(true))
		if err != nil {
			title := fmt.Sprintf("下载用户数据库失败")
			ui.ErrorPopup(title, err.Error())
			return
		}
		db.SetProgressCallback(func(cur int) error {
			if cur == userdb.MinProgress {
				pd.SetLabelText(msgs[msgIndex])
				msgIndex++
			}
			pd.SetValue(cur)
			if pd.WasCanceled() {
				return errors.New("cancelled")
			}
			return nil
		})
		err = db.WriteMD380ToolsFile(tmpFilename)
		if err != nil {
			os.Remove(tmpFilename)
			pd.Close()
			title := fmt.Sprintf("下载用户数据库失败")
			ui.ErrorPopup(title, err.Error())
			return
		}

		os.Rename(tmpFilename, filename)
	}

	pd.SetRange(dfu.MinProgress, dfu.MaxProgress)
	df, err := dfu.New(func(cur int) error {
		if cur == dfu.MinProgress {
			pd.SetLabelText(msgs[msgIndex])
			msgIndex++
		}
		pd.SetValue(cur)
		if pd.WasCanceled() {
			return errors.New("cancelled")
		}
		return nil

	})
	if err == nil {
		defer df.Close()
		db, err := userdb.New(userdb.FromFile(filename), userdb.Abbreviate(true))
		if err == nil {
			err = df.WriteMD380Users(db)
		}
	}
	if err != nil {
		pd.Close()
		title := fmt.Sprintf("写入用户数据库失败：%s", err.Error())
		ui.ErrorPopup(title, err.Error())
	}
}

func writeExpandedUsers(title, text string) {
	cancel, download := userdbDialog(title, text)
	if cancel {
		return
	}

	locType := core.QStandardPaths__CacheLocation
	cacheDir := core.QStandardPaths_WritableLocation(locType)
	tmpFilename := filepath.Join(cacheDir, "users.tmp")

	msgs := []string{
		"正在从网站下载用户数据库...",
		"Preparing to write user database to radio...",
		"正在擦除对讲机用户数据库...",
		"正在写入用户数据库到对讲机...",
	}
	msgIndex := 0
	if !download {
		msgIndex = 1
	}

	filename := userdbFilename()
	os.MkdirAll(filepath.Dir(filename), os.ModeDir|0755)

	pd := ui.NewProgressDialog(msgs[msgIndex])
	pd.SetRange(userdb.MinProgress, userdb.MaxProgress)

	if download {
		db, err := userdb.New(userdb.CuratedUsers(), userdb.Abbreviate(false))
		if err != nil {
			title := fmt.Sprintf("下载用户数据库失败")
			ui.ErrorPopup(title, err.Error())
			return
		}
		db.SetProgressCallback(func(cur int) error {
			if cur == userdb.MinProgress {
				pd.SetLabelText(msgs[msgIndex])
				msgIndex++
			}
			pd.SetValue(cur)
			if pd.WasCanceled() {
				return errors.New("cancelled")
			}
			return nil
		})
		err = db.WriteMD380ToolsFile(tmpFilename)
		if err != nil {
			os.Remove(tmpFilename)
			pd.Close()
			title := fmt.Sprintf("下载用户数据库失败")
			ui.ErrorPopup(title, err.Error())
			return
		}

		os.Rename(tmpFilename, filename)
	}

	pd.SetRange(dfu.MinProgress, dfu.MaxProgress)
	df, err := dfu.New(func(cur int) error {
		if cur == dfu.MinProgress {
			pd.SetLabelText(msgs[msgIndex])
			msgIndex++
		}
		pd.SetValue(cur)
		if pd.WasCanceled() {
			return errors.New("cancelled")
		}
		return nil

	})
	if err == nil {
		defer df.Close()
		if err == nil {
			db, err := userdb.New(userdb.FromFile(filename), userdb.Abbreviate(false))
			if err == nil {
				err = df.WriteUV380Users(db)
			}
		}
	}
	if err != nil {
		pd.Close()
		title := fmt.Sprintf("写入用户数据库失败：%s", err.Error())
		ui.ErrorPopup(title, err.Error())
	}
}

func writeMD2017Users() {
	title := "写入用户数据库到对讲机"
	text := `
	The users database contains DMR ID numbers and callsigns of all registered
	users. It can only be be written to MD-2017 radios.

	WARNING: This only works on MD-2017 radios with the "CSV" firmware versions.`

	writeExpandedUsers(title, text)
}

func writeUV380Users() {
	title := "写入用户数据库到对讲机"
	text := `
	The users database contains DMR ID numbers and callsigns of all registered
	users. It can only be be written to MD-UV380 radios.

	WARNING: This only works on MD-UV380 radios with the "CSV" firmware versions.`

	writeExpandedUsers(title, text)
}

func factoryFirmwareDialog(modelURLs []modelURL) {
	model, url := modelURLs[0].model, modelURLs[0].url
	if len(modelURLs) > 1 {
		title := "写入出厂固件到对讲机..."
		upgrade := false
		var canceled bool
		canceled, model, url = firmwareDialog(title, modelURLs, upgrade)
		if canceled {
			return
		}
	}

	msgs := []string{
		fmt.Sprintf("正在下载%s出厂固件...\n%s", model, url),
		"正在擦除对讲机固件...",
		fmt.Sprintf("正在写入%s出厂固件到对讲机...", model),
	}

	writeFirmware(url, msgs)
}

func (edt *editor) addRadioMenu(menu *ui.Menu) {
	cp := edt.codeplug
	mb := edt.mainWindow.MenuBar()
	menu = mb.AddMenu("对讲机")

	menu.AddAction("从对讲机读取写频", func() {
		// macOS 上要跳过 RadioExists()，因为它先打开 USB 设备调用 enterDfuMode，
		// 会发送 detach 信号让对讲机重启进刷机模式（断开USB重连），
		// 然后 Close() 关闭 libusb context。后续 ReadRadio 重新打开设备时
		// 可能因为设备正在重连或 libusb context 状态不一致而失败。
		//
		// 正确做法：直接进入 ReadRadio，由 ReadRadio 内部 dfu.New 处理设备查找和打开。

		edt := newEditor(edt.app, codeplug.FileTypeNew, "")
		if edt == nil || edt.codeplug == nil {
			return
		}

		cp := edt.codeplug

		msgs := []string{
			"准备从对讲机读取写频...",
			"正在从对讲机读取写频...",
		}
		msgIndex := 0
		pd := ui.NewProgressDialog(msgs[msgIndex])
		pd.SetRange(codeplug.MinProgress, codeplug.MaxProgress)
		err := cp.ReadRadio(func(cur int) error {
			if cur == codeplug.MinProgress {
				pd.SetLabelText(msgs[msgIndex])
				msgIndex++
			}

			pd.SetValue(cur)
			if pd.WasCanceled() {
				return errors.New("cancelled")
			}
			return nil
		})
		if err != nil {
			pd.Close()
			title := "读取写频失败"
			ui.ErrorPopup(title, err.Error()+"\n\n请检查：\n1. 对讲机是否已进入刷机模式（开机时按住PTT+PTT上方按键）\n2. 写频线连接是否正常\n3. USB 权限是否已授权（系统设置 > 隐私 > USB）")
			edt.FreeCodeplug()
			return
		}
		pd.Close()

		if !cp.Valid() {
			fmtStr := `
	%d records with invalid field values were found in the codeplug.

	Select "Menu->Edit->Show Invalid Fields" to view them.`
			msg := fmt.Sprintf(fmtStr, len(cp.Warnings()))
			ui.InfoPopup("codeplug warning", msg)
		}
		edt.updateMenuBar()
	})

	menu.AddAction("写入写频到对讲机", func() {
		valid := cp.Valid()
		edt.updateMenuBar()
		if !valid {
			fmtStr := `
	%d records with invalid field values were found in the codeplug.

	Click on Cancel and then select "Menu->Edit->Show Invalid Fields" to view them.

	Or, click on Ignore to continue writing to the radio.`
			msg := fmt.Sprintf(fmtStr, len(cp.Warnings()))
			title := "写入警告"
			rv := ui.WarningPopup(title, msg)
			if rv != ui.PopupIgnore {
				return
			}
		}

		title := "写入写频到对讲机"
		model := codeplug.ModelTypes(cp.Model())
		freq := cp.FrequencyRange()
		warn := `

	WARNING: Corruption may occur if a signal is received
	while writing to the radio.  The radio should be tuned
	to an unprogrammed (or at least quiet) channel while
	writing the new codeplug.`
		msg := fmt.Sprintf("%s\n\nWrite %s %s codeplug to radio?\n", warn, model, freq)
		if ui.YesNoPopup(title, msg) != ui.PopupYes {
			return
		}

		msgs := []string{
			"准备写入写频到对讲机...",
			"正在擦除对讲机写频...",
			"正在写入写频到对讲机...",
		}
		msgIndex := 0

		pd := ui.NewProgressDialog(msgs[msgIndex])
		pd.SetRange(codeplug.MinProgress, codeplug.MaxProgress)
		err := cp.WriteRadio(func(cur int) error {
			if cur == codeplug.MinProgress {
				pd.SetLabelText(msgs[msgIndex])
				msgIndex++
			}
			pd.SetValue(cur)
			if pd.WasCanceled() {
				return errors.New("cancelled")
			}
			return nil
		})
		if err != nil {
			pd.Close()
			title := "写入写频到对讲机失败"
			ui.ErrorPopup(title, err.Error())
		}
	}).SetEnabled(cp != nil && cp.Loaded())

	menu.AddSeparator()

	fwMenu := menu.AddMenu("写入出厂固件到对讲机...")

	fwMenu.AddAction("写入MD-380出厂固件..."), func() {
		dir := "https://farnsworth.org/dale/dmr/factory_firmware/md380/"

		modelURLs := []modelURL{
			modelURL{"MD-380 old (D03.20)", dir + "D003.020.bin"},
			modelURL{"MD-380 (D13.20)", dir + "D013.020.bin"},
			modelURL{"MD-380 new (D13.34)", dir + "D013.034.bin"},
			modelURL{"MD-380 newer (D14.04", dir + "D014.004.bin"},
			modelURL{"MD-380 newest (D15.01", dir + "D015.001.bin"},
			modelURL{"MD-380G (S13.20)", dir + "S013.020.bin"},
		}

		factoryFirmwareDialog(modelURLs)
	})

	fwMenu.AddAction("写入MD-390出厂固件..."), func() {
		dir := "https://farnsworth.org/dale/dmr/factory_firmware/md390/"

		modelURLs := []modelURL{
			modelURL{"MD-390 (D13.20)", dir + "D013.020.bin"},
			modelURL{"MD-390G (S13.20)", dir + "S013.020.bin"},
		}

		factoryFirmwareDialog(modelURLs)
	})

	fwMenu.AddAction("写入RT3出厂固件", func() {
		dir := "https://farnsworth.org/dale/dmr/factory_firmware/rt3/"

		modelURLs := []modelURL{
			modelURL{"RT3 (D03.20)", dir + "D003.020.bin"},
		}

		factoryFirmwareDialog(modelURLs)
	})

	fwMenu.AddAction("写入RT8出厂固件", func() {
		dir := "https://farnsworth.org/dale/dmr/factory_firmware/rt8/"

		modelURLs := []modelURL{
			modelURL{"RT8 (S13.20)", dir + "S013.020.bin"},
		}

		factoryFirmwareDialog(modelURLs)
	})

	/*
		fwMenu.AddAction("Write MD-UV380 factory firmware...", func() {
			dir := "https://farnsworth.org/dale/dmr/factory_firmware/uv380/"

			modelURLs := []modelURL{
				modelURL{"MD-UV380 REC (D17.05)", dir + "MD-UV380(REC)-D17.05.bin"},
				modelURL{"MD-UV380 CSV (V17.05)", dir + "MD-UV380(CSV)-V17.05.bin"},
			}

			factoryFirmwareDialog(modelURLs)
		})

		fwMenu.AddAction("Write MD-UV390 factory firmware...", func() {
			dir := "https://farnsworth.org/dale/dmr/factory_firmware/uv390/"

			modelURLs := []modelURL{
				modelURL{"MD-UV390 GPS-REC (S17.05)", dir + "MD-UV390(GPS-REC)-S17.05.bin"},
				modelURL{"MD-UV390 GPS-CSV (P17.05)", dir + "MD-UV390(CSV-GPS)-P17.05.bin"},
			}

			factoryFirmwareDialog(modelURLs)
		})
	*/

	menu.AddSeparator()
	writeUsersMenu := menu.AddMenu("写入用户数据库到对讲机...")

	writeUsersMenu.AddAction("Write md380tools user database to radio...", writeMD380toolsUsers)
	writeUsersMenu.AddAction("写入MD2017用户数据库到对讲机..."), writeMD2017Users)
	writeUsersMenu.AddAction("写入MD-UV380用户数据库到对讲机..."), writeUV380Users)

	menu.AddSeparator()

	md380toolsMenu := menu.AddMenu("md380tools...")

	md380toolsMenu.AddAction("Write user database to radio...", writeMD380toolsUsers)

	md380toolsMenu.AddAction("写入md380tools固件到对讲机...", func() {
		path := "https://farnsworth.org/dale/md380tools/firmware/"
		nonGpsURL := path + "D13.20.bin"
		gpsURL := path + "S13.20.bin"

		modelURLs := []modelURL{
			modelURL{"MD-380 (D13.20)", nonGpsURL},
			modelURL{"MD-380G (S13.20)", gpsURL},
			modelURL{"MD-390 (D13.20)", nonGpsURL},
			modelURL{"MD-390G (S13.20)", gpsURL},
			modelURL{"RT3 (D13.20)", nonGpsURL},
			modelURL{"RT8 (S13.20)", gpsURL},
		}

		title := "写入md380tools固件到对讲机..."
		upgrade := true
		canceled, model, url := firmwareDialog(title, modelURLs, upgrade)
		if canceled {
			return
		}

		msgs := []string{
			fmt.Sprintf("正在下载md380tools %s固件...\n%s", model, url),
			"正在擦除对讲机固件...",
			fmt.Sprintf("正在写入md380tools %s固件到对讲机...", model),
		}

		writeFirmware(url, msgs)
	})
	md380toolsMenu.AddAction("写入KD4Z md380tools固件到对讲机...", func() {
		path := "https://farnsworth.org/dale/md380tools/kd4z/"
		nonGpsURL := path + "firmware-noGPS.bin"
		gpsURL := path + "firmware-GPS.bin"

		modelURLs := []modelURL{
			modelURL{"MD-380 (D13.20)", nonGpsURL},
			modelURL{"MD-380G (S13.20)", gpsURL},
			modelURL{"MD-390 (D13.20)", nonGpsURL},
			modelURL{"MD-390G (S13.20)", gpsURL},
			modelURL{"RT3 (D13.20)", nonGpsURL},
			modelURL{"RT8 (S13.20)", gpsURL},
		}

		title := "写入KD4Z md380tools固件到对讲机..."
		upgrade := true
		canceled, model, url := firmwareDialog(title, modelURLs, upgrade)
		if canceled {
			return
		}

		msgs := []string{
			fmt.Sprintf("正在下载KD4Z md380tools %s固件...\n%s", model, url),
			"正在擦除对讲机固件...",
			fmt.Sprintf("Writing KD4Z md380tools %s firmware to radio...", model),
		}

		writeFirmware(url, msgs)
	})
}

func writeFirmware(url string, msgs []string) {
	tmpFile, err := ioutil.TempFile("", "editcp")
	if err != nil {
		title := fmt.Sprintf("临时文件创建失败：%s", err.Error())
		ui.ErrorPopup(title, err.Error())
		return
	}

	filename := tmpFile.Name()
	defer os.Remove(filename)

	msgIndex := 0
	pd := ui.NewProgressDialog(msgs[msgIndex])
	pd.SetRange(dfu.MinProgress, dfu.MaxProgress)

	df, err := dfu.New(func(cur int) error {
		if cur == dfu.MinProgress {
			pd.SetLabelText(msgs[msgIndex])
			msgIndex++
		}
		pd.SetValue(cur)
		if pd.WasCanceled() {
			return errors.New("cancelled")
		}
		return nil

	})
	if err != nil {
		pd.Close()
		title := "固件写入失败"
		ui.ErrorPopup(title, err.Error())
		return
	}
	defer df.Close()

	pd.SetRange(userdb.MinProgress, userdb.MaxProgress)
	err = download(url, filename, func(cur int) bool {
		if cur == dfu.MinProgress {
			pd.SetLabelText(msgs[msgIndex])
			msgIndex++
		}
		pd.SetValue(cur)
		if pd.WasCanceled() {
			return false
		}
		return true
	})
	if err != nil {
		pd.Close()
		title := "固件写入失败"
		ui.ErrorPopup(title, err.Error())
		return
	}

	file, err := os.Open(filename)
	if err != nil {
		l.Fatalf("writeFirmware: %s", err.Error())
	}

	defer file.Close()

	err = df.WriteFirmware(file)
	if err != nil {
		pd.Close()
		title := "write of new firmware failed"
		ui.ErrorPopup(title, err.Error())
		return
	}

	msg := "关闭对讲机后重新开机。"
	ui.InfoPopup("固件写入完成", msg)
}

func userdbFilename() string {
	locType := core.QStandardPaths__CacheLocation
	cacheDir := core.QStandardPaths_WritableLocation(locType)

	name := "usersDB.bin"

	return filepath.Join(cacheDir, name)
}

func userdbDialog(title string, labelText string) (canceled, download bool) {
	loadSettings()

	usersFilename := userdbFilename()

	download = true
	if fileYounger(usersFilename, 12*time.Hour) {
		download = false
	}

	downloadCheckbox := ui.NewCheckboxWidget(download, func(checked bool) {
		download = checked
	})
	downloadCheckbox.SetEnabled(fileExists(usersFilename))

	dialog := ui.NewDialog(title)

	filenameBox := ui.NewHbox()
	filenameBox.AddLabel("   " + usersFilename)

	dialog.AddLabel(labelText[1:])

	form := dialog.AddForm()
	form.AddRow("下载新的用户数据库文件", downloadCheckbox)

	dialog.AddLabel("文件名：")
	dialog.AddExistingHbox(filenameBox)

	row := dialog.AddHbox()
	cancelButton := ui.NewButtonWidget("Cancel", func() {
		dialog.Reject()
	})
	row.AddWidget(cancelButton)

	saveButton := ui.NewButtonWidget("写入", func() {
		dialog.Accept()
	})
	row.AddWidget(saveButton)

	saved := dialog.Exec()
	return !saved, download
}

func firmwareDialog(title string, modelURLs []modelURL, upgrade bool) (canceled bool, model, url string) {

	models := make([]string, len(modelURLs))
	for i, modelURL := range modelURLs {
		models[i] = modelURL.model
	}

	model = models[0]
	modelCombobox := ui.NewComboboxWidget(model, models, func(index int) {
		if index < 0 {
			return
		}
		model = models[index]
	})

	dialog := ui.NewDialog(title)

	var labelText string
	if upgrade {
		labelText += `
	The md380tools firmware only works on MD380, MD380, RT3, and RT8 radios.`
	}

	labelText += `

	Before continuing, enable bootloader mode:
		1. Insert a cable into USB.
		2. Connect the cable to the radio.
		3. Power-on the radio by turning volume knob, while holding down
		   the PTT button and the button above PTT.

	While in bootloader mode, the LED will flash green and red.`

	if !upgrade {
		labelText += `

	Hint: If the display becomes flipped on the md380, try another
	md380 variant.`
	}

	dialog.AddLabel(labelText[1:])

	groupBox := dialog.AddGroupbox("选择对讲机型号")
	form := groupBox.AddForm()
	form.AddRow("对讲机型号", modelCombobox)

	row := dialog.AddHbox()

	cancelButton := ui.NewButtonWidget("Cancel", func() {
		dialog.Reject()
	})
	row.AddWidget(cancelButton)

	saveButton := ui.NewButtonWidget("更新固件", func() {
		dialog.Accept()
	})
	row.AddWidget(saveButton)

	saved := dialog.Exec()

	for _, modelURL := range modelURLs {
		if modelURL.model == model {
			url = modelURL.url
			break
		}
	}

	return !saved, model, url
}

var timeoutSeconds = 20

var tr = &http.Transport{
	TLSHandshakeTimeout:   time.Duration(timeoutSeconds) * time.Second,
	ResponseHeaderTimeout: time.Duration(timeoutSeconds) * time.Second,
}

var client = &http.Client{
	Transport: tr,
	Timeout:   time.Duration(timeoutSeconds) * time.Second,
}

type downloader struct {
	url               string
	filename          string
	progressCallback  func(progressCounter int) bool
	progressFunc      func() error
	progressIncrement int
	progressCounter   int
}

func newDownloader() *downloader {
	d := &downloader{
		progressFunc: func() error { return nil },
	}

	return d
}

func (d *downloader) setMaxProgressCount(max int) {
	d.progressFunc = func() error { return nil }
	if d.progressCallback != nil {
		d.progressIncrement = MaxProgress / max
		d.progressCounter = 0
		d.progressFunc = func() error {
			d.progressCounter += d.progressIncrement
			curProgress := d.progressCounter
			if curProgress > MaxProgress {
				curProgress = MaxProgress
			}

			if !d.progressCallback(d.progressCounter) {
				return errors.New("")
			}

			return nil
		}
		d.progressCallback(d.progressCounter)
	}
}

func (d *downloader) finalProgress() {
	//fmt.Fprintf(os.Stderr, "\nprogressMax %d\n", d.progressCounter/d.progressIncrement)
	if d.progressCallback != nil {
		d.progressCallback(MaxProgress)
	}
}

// Minimum and maximum progress values
const (
	MinProgress = 0
	MaxProgress = 1000000
)

func download(url, filename string, progress func(cur int) bool) error {
	d := newDownloader()
	d.url = url
	d.filename = filename
	d.progressCallback = progress
	return d.download()
}

func (d *downloader) download() (err error) {
	file, err := os.Create(d.filename)
	if err != nil {
		return wrapError("download", err)
	}
	defer func() {
		fErr := file.Close()
		if err == nil {
			err = fErr
		}
		return
	}()

	resp, err := client.Get(d.url)
	if err != nil {
		return wrapError("download", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return wrapError("download", errors.New(resp.Status))
	}
	length := resp.ContentLength
	if length < 0 {
		length = 1024 * 1024
	}

	bufSize := 16 * 1024

	d.setMaxProgressCount(int(length) / bufSize)

	buf := make([]byte, bufSize)
	for {
		err := d.progressFunc()
		if err != nil {
			return wrapError("download", err)
		}

		n, err := resp.Body.Read(buf)
		if n == 0 && err != nil {
			if err == io.EOF {
				break
			}
			return wrapError("download", err)
		}

		n, err = file.Write(buf)
		if err != nil {
			return wrapError("download", err)
		}
	}

	d.finalProgress()

	return nil
}

func wrapError(prefix string, err error) error {
	if err.Error() == "" {
		return err
	}
	return fmt.Errorf("%s: %s", prefix, err.Error())
}

func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}

func fileYounger(filename string, duration time.Duration) bool {
	fileInfo, err := os.Stat(filename)
	return err == nil && time.Since(fileInfo.ModTime()) < duration
}
