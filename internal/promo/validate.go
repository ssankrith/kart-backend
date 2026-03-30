package promo

import "unicode/utf8"

// CouponCodePreludeOK is true if code has UTF-8 length 8–10 and is ASCII (byte length equals rune count).
func CouponCodePreludeOK(code string) bool {
	n := utf8.RuneCountInString(code)
	if n < 8 || n > 10 {
		return false
	}
	return len(code) == n
}
