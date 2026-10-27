package kustolang

import (
	"encoding/hex"
	"fmt"
	"math"
	"math/big"
	"strings"
	"time"
)

// Adapted from KustoCompleter.CreateValueExamples and CultureUtilities.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

func invariantTimeSpanText(value TimeSpanValue) string {
	negative := value < 0
	ticks := uint64(value)
	if negative {
		ticks = -ticks
	}
	days := ticks / uint64(timeSpanTicksPerDay)
	ticks %= uint64(timeSpanTicksPerDay)
	hours := ticks / uint64(timeSpanTicksPerHour)
	ticks %= uint64(timeSpanTicksPerHour)
	minutes := ticks / uint64(timeSpanTicksPerMinute)
	ticks %= uint64(timeSpanTicksPerMinute)
	seconds, fraction := ticks/uint64(timeSpanTicksPerSecond), ticks%uint64(timeSpanTicksPerSecond)
	text := fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds)
	if days > 0 {
		text = fmt.Sprintf("%d.%s", days, text)
	}
	if fraction > 0 {
		text += fmt.Sprintf(".%07d", fraction)
	}
	if negative {
		text = "-" + text
	}
	return text
}

func invariantCompletionValue(value any) string {
	switch v := value.(type) {
	case time.Time:
		return "datetime(" + v.Format("01/02/2006 15:04:05") + ")"
	case TimeSpanValue:
		return "timespan(" + invariantTimeSpanText(v) + ")"
	case time.Duration:
		return "timespan(" + invariantTimeSpanText(TimeSpanValue(v/100)) + ")"
	case [16]byte:
		text := hex.EncodeToString(v[:])
		return "guid(" + text[:8] + "-" + text[8:12] + "-" + text[12:16] + "-" + text[16:20] + "-" + text[20:] + ")"
	case *big.Rat:
		text := strings.TrimRight(strings.TrimRight(v.FloatString(28), "0"), ".")
		return "decimal(" + text + ")"
	case bool:
		if v {
			return "True"
		}
		return "False"
	case float64:
		if math.IsInf(v, 1) {
			return "Infinity"
		}
		if math.IsInf(v, -1) {
			return "-Infinity"
		}
	}
	return fmt.Sprint(value)
}
