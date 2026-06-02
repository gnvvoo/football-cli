package api

import (
	"strings"
	"unicode"

	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

// normalize : 발음 구별 부호 제거 후 소문자 변환
// Modrić → modric, Müller → muller
func normalize(s string) string {
	t := transform.Chain(
		norm.NFD,
		transform.RemoveFunc(func(r rune) bool {
			return unicode.Is(unicode.Mn, r) // Mn: 발음 구별 부호 카테고리
		}),
		norm.NFC,
	)
	result, _, _ := transform.String(t, s)
	return strings.ToLower(result)
}

// containsNormalized : 발음 구별 부호 무시하고 부분 문자열 매칭
func containsNormalized(s, query string) bool {
	return strings.Contains(normalize(s), normalize(query))
}