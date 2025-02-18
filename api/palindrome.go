package api

import (
	"unicode"
)

func Palindrome(s string) (string, bool) {
	if len(s) < 6 {
		return "", false
	}

	var cleaned []rune
	var start int = -1
	var end int = -1
	charSet := make(map[rune]bool) // Track unique characters
	hasLettersOrDigits := false    // Ensure at least one real letter/digit exists

	// Iterate over runes, filtering out non-English characters and punctuation
	for i, r := range s {
		if containsEmoji(r) || !isEnglishLetter(r) {
			continue
		}
		lowerR := unicode.ToLower(r)
		if start == -1 {
			start = i // Mark first valid character
		}
		end = i
		cleaned = append(cleaned, lowerR)
		charSet[lowerR] = true
		hasLettersOrDigits = true
	}

	if start == -1 {
		return "", false
	}

	original := []rune(s[start : end+1])

	// If all characters are the same, return false
	if len(charSet) == 1 {
		return "", false
	}

	// Ensure we have at least one valid letter or digit
	if !hasLettersOrDigits {
		return "", false
	}

	// Convert cleaned runes to a string for regex check
	cleanedStr := string(cleaned)
	if checkTwoLetterRepetition(cleanedStr) {
		return "", false
	}

	// If the string starts and ends with the same character, with a single repeated character in the middle, return false
	if len(charSet) == 2 && cleaned[0] == cleaned[len(cleaned)-1] {
		midChar := cleaned[1]
		for _, r := range cleaned[2 : len(cleaned)-1] {
			if r != midChar {
				break
			}
		}
		return "", false
	}

	// If the string is too short, return false
	if len(cleaned) < 6 {
		return "", false
	}

	// Check palindrome property
	a, b := 0, len(cleaned)-1
	for a < b {
		if cleaned[a] != cleaned[b] {
			return "", false
		}
		a++
		b--
	}

	return string(original), true
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

// checkTwoLetterRepetition verifies if the string follows the alternating "XYXYXY" pattern
func checkTwoLetterRepetition(s string) bool {
	if len(s) < 4 {
		return false
	}

	x, y := s[0], s[1]

	matchesXY := true
	matchesYX := true

	for i := 0; i < len(s); i++ {
		if i%2 == 0 && s[i] != x {
			matchesXY = false
		}
		if i%2 == 1 && s[i] != y {
			matchesXY = false
		}

		// Alternate pattern (shifted by one position)
		if i%2 == 0 && s[i] != y {
			matchesYX = false
		}
		if i%2 == 1 && s[i] != x {
			matchesYX = false
		}
	}

	return (matchesXY || matchesYX)
}
