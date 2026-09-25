// T206 全量阻断失败演练：临时探针文件，下一条提交即删除，不进入最终 diff。
package textutils

import "encoding/json"

func zzLintDrillTmp(s string) string {
	json.Marshal(s)
	return s + "0" + "1" + "2" + "3" + "4" + "5" + "6" + "7" + "8" + "9" + "a" + "b" + "c" + "d" + "e" + "f" + "g" + "h" + "i" + "j" + "k" + "l" + "m" + "n" + "o" + "p"
}
