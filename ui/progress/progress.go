package progress

import (
	"fmt"
	"strings"
)

func Body(current, total int64, width, pulseWidth int) string {
	if total <= 0 {
		frame := int(current / 1024)
		return fmt.Sprintf("%s %d bytes", Indeterminate(frame, width, pulseWidth), current)
	}
	percentage := float64(current) / float64(total) * 100
	if percentage > 100 {
		percentage = 100
	}
	if percentage < 0 {
		percentage = 0
	}
	filled := int(percentage / 100 * float64(width))
	bar := strings.Repeat("=", filled) + strings.Repeat("-", width-filled)
	return fmt.Sprintf("[%s] %.0f%%", bar, percentage)
}

func Indeterminate(frame, width, pulseWidth int) string {
	if width < 1 {
		return "[]"
	}
	if pulseWidth < 1 {
		pulseWidth = 1
	}
	if pulseWidth > width {
		pulseWidth = width
	}
	travel := width - pulseWidth
	if travel == 0 {
		return "[" + strings.Repeat("=", width) + "]"
	}
	cycle := travel * 2
	position := frame % cycle
	if position < 0 {
		position += cycle
	}
	forward := position <= travel
	if !forward {
		position = cycle - position
	}

	bar := []byte(strings.Repeat("-", width))
	for index := range pulseWidth {
		bar[position+index] = '='
	}
	if forward {
		bar[position+pulseWidth-1] = '>'
	} else {
		bar[position] = '<'
	}
	return "[" + string(bar) + "]"
}
