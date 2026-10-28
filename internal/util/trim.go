package util

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// trimRangeRe matches a "start-end" pair where each side is SS, MM:SS or HH:MM:SS.
// [0-9] documents intent and rejects other numeral systems.
var trimRangeRe = regexp.MustCompile(`^([0-9]{1,2}(?::[0-9]{2}){0,2})-([0-9]{1,2}(?::[0-9]{2}){0,2})$`)

// ValidateTrimRange checks a user-entered video trim range such as "01:20-03:45".
// An empty string means "no trim" and is valid. Start must be before end, otherwise
// yt-dlp silently downloads nothing.
func ValidateTrimRange(s string) error {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}

	m := trimRangeRe.FindStringSubmatch(s)
	if m == nil {
		return fmt.Errorf("khoảng cắt %q không hợp lệ. Dùng dạng MM:SS-MM:SS hoặc HH:MM:SS-HH:MM:SS, ví dụ 01:20-03:45", s)
	}

	start, okStart := trimSeconds(m[1])
	end, okEnd := trimSeconds(m[2])
	if !okStart || !okEnd {
		return fmt.Errorf("khoảng cắt %q không hợp lệ: phút/giây phải nhỏ hơn 60", s)
	}
	if start >= end {
		return fmt.Errorf("khoảng cắt %q không hợp lệ: thời điểm bắt đầu phải nhỏ hơn thời điểm kết thúc", s)
	}
	return nil
}

// trimSeconds converts "SS", "MM:SS" or "HH:MM:SS" into seconds.
func trimSeconds(v string) (int, bool) {
	parts := strings.Split(v, ":")
	total := 0
	for i, p := range parts {
		n, err := strconv.Atoi(p)
		if err != nil {
			return 0, false
		}
		// Only the leading component may exceed 59 (it is hours or a bare second count).
		if i > 0 && n > 59 {
			return 0, false
		}
		total = total*60 + n
	}
	return total, true
}
