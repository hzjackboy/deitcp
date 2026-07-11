// Package lang 提供中英文界面切换功能
package lang

import "sync"

var (
	mu          sync.RWMutex
	currentLang = ZH // 默认中文
)

// Lang 语言类型
type Lang int

const (
	ZH Lang = iota // 中文
	EN             // 英文
)

// SetLang 设置当前语言
func SetLang(l Lang) {
	mu.Lock()
	defer mu.Unlock()
	currentLang = l
}

// GetLang 获取当前语言
func GetLang() Lang {
	mu.RLock()
	defer mu.RUnlock()
	return currentLang
}

// T 根据当前语言返回对应文本
// zh: 中文文本, en: 英文文本
func T(zh, en string) string {
	mu.RLock()
	defer mu.RUnlock()
	if currentLang == EN {
		return en
	}
	return zh
}


// RecordTypeEN 返回记录类型名称的英文版本
var recordTypeEN = map[string]string{
	"信道列表": "Channels",
	"区域": "Zones",
	"联系人": "Contacts",
}

// RecordTypeCN 返回记录类型名称的中文版本
var recordTypeCN = map[string]string{}

func init() {
	for k, v := range recordTypeEN {
		recordTypeCN[v] = k
	}
}

// TRecordType 根据地返回记录类型显示名称
func TRecordType(name string) string {
	mu.RLock()
	defer mu.RUnlock()
	if currentLang == EN {
		if en, ok := recordTypeEN[name]; ok {
			return en
		}
	} else {
		if cn, ok := recordTypeCN[name]; ok {
			return cn
		}
	}
	return name
}
