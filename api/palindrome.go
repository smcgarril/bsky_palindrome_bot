package api

import (
	"regexp"
	"unicode"
)

// Regex to detect repetitive laughter patterns (strictly repeated units)
var laughRegex = regexp.MustCompile(`^(ha)+$|^(ha)+[h]+$|^(ah)+$|^(ah)+[a]+$|^(he)+$|^(he)+[h]+$|^(eh)+$|^(eh)+[e]+$|^(ho)+$|^(ho)+[h]+$|^(oh)+$|^(oh)+[o]+$|^(hi)+$|^(hi)+[h]+$|^(ih)+$|^(ih)+[i]+$|^(hu)+$|^(hu)+[h]+$|^(uh)+$|^(uh)+[u]+$|^(ja)+$|^(ja)+[j]+$|^(aj)+$|^(aj)+[a]+$|^(je)+$|^(je)+[j]+$|^(ej)+$|^(ej)+[e]+$|^(ji)+$|^(ji)+[j]+$|^(ij)+$|^(ij)+[i]+$|^(jo)+$|^(jo)+[j]+$|^(oj)+$|^(oj)+[o]+$|^(ju)+$|^(ju)+[j]+$|^(uj)+$|^(uj)+[u]+$|^(lo)+$|^(lo)+[l]+$|^(ol)+$|^(ol)+[o]+$|^(le)+$|^(le)+[l]+$|^(el)+$|^(el)+[e]+$|^(lmao)+$|^(rofl)+$|^(wo)+$|^(wo)+[w]+$|^(ow)+$|^(ow)+[o]+$`)

func Palindrome(s string) bool {
	var cleaned []rune
	charSet := make(map[rune]bool) // Track unique characters
	hasLettersOrDigits := false    // Ensure at least one real letter/digit exists

	// Iterate over runes, filtering out non-English characters and punctuation
	for _, r := range s {
		if containsEmoji(r) || !isEnglishLetter(r) {
			return false
		}
		lowerR := unicode.ToLower(r)
		cleaned = append(cleaned, lowerR)
		charSet[lowerR] = true
		hasLettersOrDigits = true
	}

	// If all characters are the same, return false
	if len(charSet) == 1 {
		return false
	}

	// Ensure we have at least one valid letter or digit
	if !hasLettersOrDigits {
		return false
	}

	// Convert cleaned runes to a string for regex check
	cleanedStr := string(cleaned)
	if laughRegex.MatchString(cleanedStr) {
		return false
	}

	// Check palindrome property
	a, b := 0, len(cleaned)-1
	for a < b {
		if cleaned[a] != cleaned[b] {
			return false
		}
		a++
		b--
	}

	return true
}

// Only allow English letters
func isEnglishLetter(r rune) bool {
	return (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z')
}

// Check if a rune is an emoji.
func containsEmoji(r rune) bool {
	return (r >= 0x1F600 && r <= 0x1F64F) || // Emoticons
		(r >= 0x1F300 && r <= 0x1F5FF) || // Misc symbols and pictographs
		(r >= 0x1F680 && r <= 0x1F6FF) || // Transport and map symbols
		(r >= 0x1F700 && r <= 0x1F77F) || // Alchemical symbols
		(r >= 0x1F780 && r <= 0x1F7FF) || // Geometric shapes
		(r >= 0x1F800 && r <= 0x1F8FF) || // Supplemental arrows
		(r >= 0x1F900 && r <= 0x1F9FF) || // Supplemental symbols and pictographs
		(r >= 0x2600 && r <= 0x26FF) || // Misc symbols
		(r >= 0x2700 && r <= 0x27BF) || // Dingbats
		(r >= 0xFE00 && r <= 0xFE0F) || // Variation selectors
		(r >= 0x1FA70 && r <= 0x1FAFF) // Extended pictographic characters
}
