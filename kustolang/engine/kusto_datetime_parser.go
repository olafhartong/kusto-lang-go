package kustolang

// Ported from src/Kusto.Language/Parser/KustoDateTimeParser.cs.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

import (
	"math"
	"strconv"
	"strings"
	"time"
	"unicode"
)

type dateTimeInterpretation string

const (
	datetimeInvalid         dateTimeInterpretation = "Invalid"
	datetimeMinMax          dateTimeInterpretation = "MinMax"
	datetimeNow             dateTimeInterpretation = "Now"
	datetimeNull            dateTimeInterpretation = "Null"
	datetimeFixedWidth      dateTimeInterpretation = "FixedWidth"
	datetimeTicks           dateTimeInterpretation = "Ticks"
	datetimeEpochSeconds    dateTimeInterpretation = "EpochSeconds"
	datetimeGeneral         dateTimeInterpretation = "General"
	datetimeEpochTicks      int64                  = 621355968000000000
	datetimeMinNumericTicks int64                  = 599266080000000000
	datetimeMaxTicks        int64                  = 3155378975999999999
	datetimeTicksPerDay     int64                  = 864000000000
)

// parseKustoDateTime models service literal interpretation for diagnostics.
// SyntaxToken.Value deliberately retains the separate upstream token parser.
func parseKustoDateTime(text string) (dateTimeInterpretation, time.Time) {
	text = strings.TrimSpace(text)
	switch text {
	case "":
		return datetimeInvalid, time.Time{}
	case "null":
		return datetimeNull, time.Time{}
	case "min":
		return datetimeMinMax, time.Time{}
	case "max":
		return datetimeMinMax, datetimeFromTicks(datetimeMaxTicks)
	}
	if asciiLower(text) == "now" {
		return datetimeNow, time.Time{}
	}
	if value, ok := parseFixedDateTime(text); ok {
		return datetimeFixedWidth, value
	}
	if datetimeSignedDigits(text) {
		// Upstream accumulates the positive magnitude first, including for negatives.
		magnitude := strings.TrimLeft(text, "+-")
		if number, err := strconv.ParseInt(magnitude, 10, 64); err == nil {
			if strings.HasPrefix(text, "-") {
				number = -number
			}
			if number < 0 || number > datetimeMaxTicks {
				return datetimeInvalid, time.Time{}
			}
			if number >= datetimeMinNumericTicks {
				return datetimeTicks, datetimeFromTicks(number)
			}
			if number > (datetimeMaxTicks-datetimeEpochTicks)/10000000 {
				return datetimeInvalid, time.Time{}
			}
			return datetimeEpochSeconds, datetimeFromTicks(datetimeEpochTicks + number*10000000)
		}
	}
	if value, recognized, ok := parseCalendarDateTime(text); recognized {
		if ok {
			return datetimeGeneral, value
		}
		return datetimeInvalid, time.Time{}
	}
	// The source falls back to invariant DateTime.TryParse for additional textual
	// forms. time.Parse is strict about calendar ranges and never rolls invalid
	// authored dates into another month.
	for _, layout := range []string{time.RFC1123, time.RFC1123Z, time.RFC822, time.RFC822Z, "Mon, 02 Jan 2006 15:04:05 GMT", "Monday, January 2, 2006", "January 2, 2006", "Jan 2, 2006", "2 January 2006", "2 Jan 2006", "2006 January 2", "January 2006", "Jan 2006", "2006-01", "01/2006", "2006.1.2"} {
		if value, err := time.ParseInLocation(layout, text, time.UTC); err == nil && value.Year() >= 1 && value.Year() <= 9999 {
			return datetimeGeneral, value.UTC()
		}
	}
	normalized := strings.Join(strings.Fields(strings.Map(func(c rune) rune {
		if c == ',' || c == '.' || c == '-' {
			return ' '
		}
		return c
	}, text)), " ")
	for _, layout := range []string{"Jan 2 2006", "January 2 2006", "2 Jan 2006", "2 January 2006", "2006 Jan 2", "2006 January 2", "Jan 2 06", "January 2 06", "2 Jan 06", "2 January 06"} {
		if value, err := time.ParseInLocation(layout, normalized, time.UTC); err == nil && value.Year() >= 1 {
			if value.Year() >= 2050 && strings.HasSuffix(layout, " 06") {
				value = value.AddDate(-100, 0, 0)
			}
			return datetimeGeneral, value
		}
	}
	// Invariant DateTime.TryParse also accepts dates without a year and times
	// without a date. Their current-calendar defaults matter only internally:
	// these General interpretations produce no ambiguity warning.
	now := time.Now().UTC()
	for _, layout := range []string{"Jan 2", "January 2", "2 Jan", "2 January", "1/2"} {
		if value, err := time.Parse(layout, normalized); err == nil {
			if validDatetimeParts(now.Year(), int(value.Month()), value.Day(), 0, 0, 0) {
				return datetimeGeneral, time.Date(now.Year(), value.Month(), value.Day(), 0, 0, 0, 0, time.UTC)
			}
		}
	}
	p := datetimeScanner{s: []rune(text)}
	h, m, sec, frac, zone, ok := p.timeOfDay(false)
	if ok && p.p == len(p.s) && validDatetimeParts(now.Year(), int(now.Month()), now.Day(), h, m, sec) && zone >= -840 && zone <= 840 {
		value := time.Date(now.Year(), now.Month(), now.Day(), h, m, sec, frac*100, time.UTC).Add(-time.Duration(zone) * time.Minute)
		return datetimeGeneral, value
	}
	return datetimeInvalid, time.Time{}
}
func datetimeSignedDigits(s string) bool {
	if s != "" && (s[0] == '+' || s[0] == '-') {
		s = s[1:]
	}
	return s != "" && datetimeDigits(s)
}
func datetimeDigits(s string) bool {
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}
func datetimeFromTicks(ticks int64) time.Time {
	return time.Date(1, 1, 1, 0, 0, 0, 0, time.UTC).AddDate(0, 0, int(ticks/datetimeTicksPerDay)).Add(time.Duration(ticks%datetimeTicksPerDay) * 100)
}
func datetimeToTicks(t time.Time) int64 {
	y := int64(t.Year() - 1)
	days := 365*y + y/4 - y/100 + y/400 + int64(t.YearDay()-1)
	return days*datetimeTicksPerDay + int64(t.Hour()*3600+t.Minute()*60+t.Second())*10000000 + int64(t.Nanosecond()/100)
}
func validDatetimeParts(y, m, d, h, min, sec int) bool {
	if y < 1 || y > 9999 || m < 1 || m > 12 || d < 1 || h < 0 || h > 23 || min < 0 || min > 59 || sec < 0 || sec > 59 {
		return false
	}
	return d <= time.Date(y, time.Month(m)+1, 0, 0, 0, 0, 0, time.UTC).Day()
}
func parseFixedDateTime(s string) (time.Time, bool) {
	n := len(s)
	y, m, d, h, min, sec, frac := 0, 1, 1, 0, 0, 0, 0
	number := func(start, length int) int { v, _ := strconv.Atoi(s[start : start+length]); return v }
	if (n == 4 || n == 6 || n == 8 || n == 10 || n == 12 || n == 14) && datetimeDigits(s) {
		y = number(0, 4)
		if n >= 6 {
			m = number(4, 2)
		}
		if n >= 8 {
			d = number(6, 2)
		}
		if n >= 10 {
			h = number(8, 2)
		}
		if n >= 12 {
			min = number(10, 2)
		}
		if n >= 14 {
			sec = number(12, 2)
		}
	} else if (n == 17 || (n >= 19 && n <= 25)) && s[8] == ' ' && s[11] == ':' && s[14] == ':' && datetimeDigits(s[:8]+s[9:11]+s[12:14]+s[15:17]) {
		y, m, d, h, min, sec = number(0, 4), number(4, 2), number(6, 2), number(9, 2), number(12, 2), number(15, 2)
		if n > 17 {
			if s[17] != '.' || !datetimeDigits(s[18:]) {
				return time.Time{}, false
			}
			frac = number(18, n-18)
			for i := n - 18; i < 7; i++ {
				frac *= 10
			}
		}
	} else {
		return time.Time{}, false
	}
	if !validDatetimeParts(y, m, d, h, min, sec) {
		return time.Time{}, false
	}
	return time.Date(y, time.Month(m), d, h, min, sec, frac*100, time.UTC), true
}

type datetimeScanner struct {
	s []rune
	p int
}

func (p *datetimeScanner) consume(c rune) bool {
	if p.p < len(p.s) && p.s[p.p] == c {
		p.p++
		return true
	}
	return false
}
func (p *datetimeScanner) space() {
	for p.p < len(p.s) && unicode.IsSpace(p.s[p.p]) {
		p.p++
	}
}
func (p *datetimeScanner) digits() int {
	i := p.p
	for i < len(p.s) && p.s[i] >= '0' && p.s[i] <= '9' {
		i++
	}
	return i - p.p
}
func (p *datetimeScanner) letters() int {
	i := p.p
	for i < len(p.s) && unicode.IsLetter(p.s[i]) {
		i++
	}
	return i - p.p
}
func (p *datetimeScanner) number(max int) (int, bool) {
	start := p.p
	v := 0
	for p.p < len(p.s) && p.p-start < max && p.s[p.p] >= '0' && p.s[p.p] <= '9' {
		v = v*10 + int(p.s[p.p]-'0')
		p.p++
	}
	return v, p.p > start
}
func (p *datetimeScanner) two() (int, bool) {
	if p.digits() < 2 {
		return 0, false
	}
	return p.number(2)
}
func (p *datetimeScanner) fraction() (int, bool) {
	n := p.digits()
	if n == 0 {
		return 0, false
	}
	f, base := 0.0, 0.1
	for i := 0; i < n; i++ {
		f += float64(p.s[p.p+i]-'0') * base
		base *= 0.1
	}
	p.p += n
	return int(math.RoundToEven(f * 10000000)), true
}
func (p *datetimeScanner) word(s string) bool {
	n := p.letters()
	return n == len(s) && asciiLower(string(p.s[p.p:p.p+n])) == s
}
func (p *datetimeScanner) zone() (int, bool) {
	if p.p == len(p.s) {
		return 0, true
	}
	if p.consume('Z') || p.consume('z') {
		return 0, true
	}
	negative := p.consume('-')
	if !negative && !p.consume('+') {
		return 0, true
	}
	n := p.digits()
	h, m := 0, 0
	switch n {
	case 1, 2:
		h, _ = p.number(n)
		if p.consume(':') {
			n = p.digits()
			if n < 1 || n > 2 {
				return 0, false
			}
			m, _ = p.number(n)
		}
	case 3, 4:
		v, _ := p.number(n)
		h, m = v/100, v%100
	default:
		return 0, false
	}
	if m > 59 {
		return 0, false
	}
	offset := h*60 + m
	if negative {
		offset = -offset
	}
	return offset, true
}
func (p *datetimeScanner) isoTime() (h, m, s, frac, zone int, ok bool) {
	p.space()
	h, ok = p.two()
	if !ok {
		return
	}
	p.space()
	if !p.consume(':') {
		ok = false
		return
	}
	p.space()
	m, ok = p.two()
	if !ok {
		return
	}
	p.space()
	if p.consume(':') {
		p.space()
		s, ok = p.two()
		if !ok {
			return
		}
		if p.consume('.') || p.consume(',') {
			frac, ok = p.fraction()
			if !ok {
				return
			}
		}
		p.space()
	}
	zone, ok = p.zone()
	ok = ok && p.p == len(p.s)
	return
}
func (p *datetimeScanner) timeOfDay(bare bool) (h, m, s, frac, zone int, ok bool) {
	h, hasHour := p.number(2)
	hasMinutes := hasHour && p.consume(':')
	if hasMinutes {
		m, ok = p.number(2)
		if !ok {
			return
		}
		if p.consume(':') {
			s, ok = p.number(2)
			if !ok {
				return
			}
			if p.consume('.') {
				frac, ok = p.fraction()
				if !ok {
					return
				}
			}
		}
	}
	p.space()
	am, pm := p.word("am"), p.word("pm")
	if am || pm {
		if !hasHour && !bare {
			return
		}
		if am {
			if h > 12 {
				h = -1
			} else if h == 12 {
				h = 0
			}
		} else if h < 12 {
			h += 12
		}
		p.p += 2
		p.space()
	} else if hasHour && !hasMinutes {
		return
	}
	if p.word("gmt") {
		p.p += 3
		ok = true
		return
	}
	zone, ok = p.zone()
	return
}
func (p *datetimeScanner) weekdaySeparators() bool {
	punctuation := false
	for p.p < len(p.s) {
		c := p.s[p.p]
		if !unicode.IsSpace(c) && c != ',' && c != '.' {
			break
		}
		if !unicode.IsSpace(c) {
			if punctuation {
				return false
			}
			punctuation = true
		}
		p.p++
	}
	return true
}

var datetimeMonthNames = []string{"january", "february", "march", "april", "may", "june", "july", "august", "september", "october", "november", "december"}
var datetimeDayNames = []string{"monday", "tuesday", "wednesday", "thursday", "friday", "saturday", "sunday"}

func datetimeName(names []string, s string) int {
	for i, n := range names {
		if (len(s) == 3 || len(s) == len(n)) && strings.HasPrefix(n, asciiLower(s)) {
			return i
		}
	}
	return -1
}
func (p *datetimeScanner) yearFirst() (y, m, d int, ok bool) {
	if len(p.s) < 5 {
		return
	}
	for _, c := range p.s[:4] {
		if c < '0' || c > '9' {
			return
		}
	}
	y, _ = p.number(4)
	if !p.consume('-') && !p.consume('/') {
		return
	}
	m, ok = p.number(2)
	if !ok {
		return
	}
	if !p.consume('-') && !p.consume('/') {
		ok = false
		return
	}
	d, ok = p.number(2)
	return
}
func (p *datetimeScanner) slashDate() (y, m, d int, ok bool) {
	m, ok = p.number(2)
	if !ok {
		return
	}
	if !p.consume('/') {
		ok = false
		return
	}
	d, ok = p.number(2)
	if !ok {
		return
	}
	if !p.consume('/') {
		ok = false
		return
	}
	start := p.p
	y, ok = p.number(4)
	if p.p-start <= 2 {
		y += 2000
		if y > 2049 {
			y -= 100
		}
	}
	return
}
func (p *datetimeScanner) monthDate() (y, m, d int, ok bool) {
	y, d = -1, -1
	for part := 0; part < 3; part++ {
		n := p.digits()
		if n > 0 {
			v, _ := p.number(n)
			if n == 4 && y < 0 {
				y = v
			} else if n <= 2 && d < 0 {
				d = v
			} else {
				return
			}
		} else {
			n = p.letters()
			i := datetimeName(datetimeMonthNames, string(p.s[p.p:p.p+n]))
			if i < 0 || m != 0 {
				return
			}
			m = i + 1
			p.p += n
		}
		if part < 2 {
			punctuation := false
			for p.p < len(p.s) {
				c := p.s[p.p]
				if !unicode.IsSpace(c) && c != ',' && c != '.' && c != '-' {
					break
				}
				if !unicode.IsSpace(c) {
					if punctuation {
						return
					}
					punctuation = true
				}
				p.p++
			}
		}
	}
	ok = y >= 0 && m > 0 && d >= 0
	return
}
func parseCalendarDateTime(text string) (time.Time, bool, bool) {
	p := datetimeScanner{s: []rune(text)}
	y, m, d, yearFirst := p.yearFirst()
	ok := yearFirst
	if !ok {
		p.p = 0
		y, m, d, ok = p.slashDate()
	}
	if !ok {
		p.p = 0
		y, m, d, ok = p.monthDate()
	}
	if !ok {
		return time.Time{}, false, false
	}
	h, min, sec, frac, zone := 0, 0, 0, 0, 0
	end := p.p
	weekday := -1
	p.space()
	isMark := p.p < len(p.s) && (p.s[p.p] == 'T' || p.s[p.p] == 't') && (p.p+1 == len(p.s) || !unicode.IsLetter(p.s[p.p+1]))
	if isMark {
		p.p++
		if !yearFirst {
			return time.Time{}, true, false
		}
		h, min, sec, frac, zone, ok = p.isoTime()
	} else {
		p.p = end
		if !p.weekdaySeparators() {
			return time.Time{}, false, false
		}
		n := p.letters()
		weekday = datetimeName(datetimeDayNames, string(p.s[p.p:p.p+n]))
		if weekday >= 0 {
			p.p += n
			if !p.weekdaySeparators() {
				return time.Time{}, false, false
			}
		} else {
			p.p = end
			p.space()
		}
		if p.p == end && p.p < len(p.s) && p.s[p.p] >= '0' && p.s[p.p] <= '9' {
			return time.Time{}, false, false
		}
		bare := weekday < 0 && p.s[end-1] >= '0' && p.s[end-1] <= '9'
		h, min, sec, frac, zone, ok = p.timeOfDay(bare)
		ok = ok && p.p == len(p.s)
	}
	if !ok || !validDatetimeParts(y, m, d, h, min, sec) || zone > 840 || zone < -840 {
		return time.Time{}, true, false
	}
	value := time.Date(y, time.Month(m), d, h, min, sec, 0, time.UTC)
	ticks := datetimeToTicks(value) + int64(frac)
	if ticks > datetimeMaxTicks {
		return time.Time{}, true, false
	}
	if weekday >= 0 && (ticks/datetimeTicksPerDay)%7 != int64(weekday) {
		return time.Time{}, true, false
	}
	ticks -= int64(zone) * 600000000
	if ticks < 0 {
		ticks += datetimeTicksPerDay
	}
	if ticks > datetimeMaxTicks {
		return time.Time{}, true, false
	}
	return datetimeFromTicks(ticks), true, true
}

func asciiLower(s string) string {
	b := []byte(s)
	for i, c := range b {
		if c >= 'A' && c <= 'Z' {
			b[i] = c - 'A' + 'a'
		}
	}
	return string(b)
}
