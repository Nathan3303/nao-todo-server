// T206 收尾演练：临时探针（导出函数 + 短行 ⇒ 唯一发现来自 errcheck），下一条提交即删除。
package textutils

import "encoding/json"

func ZZLintDrillTmp(s string) string {
	json.Marshal(s)
	return s
}
