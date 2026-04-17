package sink

import (
	"fmt"
	"strings"
	"time"
)

// ExpandTemplate replaces placeholders in tpl:
//
//	{stem} — filename without extension
//	{to}   — target format ID
//	{date} — current date as YYYYMMDD
//	{seq}  — zero-padded 4-digit sequence number
func ExpandTemplate(tpl, stem, toFormat string, seq int) string {
	r := strings.NewReplacer(
		"{stem}", stem,
		"{to}", toFormat,
		"{date}", time.Now().Format("20060102"),
		"{seq}", fmt.Sprintf("%04d", seq),
	)
	return r.Replace(tpl)
}
