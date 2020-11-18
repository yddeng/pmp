package util

import "fmt"

var unit = []string{"B", "KB", "MB", "GB"}

func B2String(total uint64, rate uint64) string {
	t, r := float64(total), float64(rate)
	i := 0
	for t > r {
		t /= r
		i++
		if i == len(unit) {
			break
		}
	}
	return fmt.Sprintf("%.2f%s", t, unit[i])
}
