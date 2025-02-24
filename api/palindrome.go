package api

import (
	"fmt"
	"strings"
	"unicode"
)

func Palindrome(validator *WordSegmentValidator, s string) (string, []string, bool) {
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
		return "", nil, false
	}

	original := []rune(s[start : end+1])

	// Convert cleaned runes to a string
	cleanedStr := string(cleaned)

	// If the string is too short, return false
	if len(cleanedStr) < 7 {
		return "", nil, false
	}

	// If all characters are the same, return false
	if len(charSet) == 1 {
		return "", nil, false
	}

	// Ensure we have at least one valid letter or digit
	if !hasLettersOrDigits {
		return "", nil, false
	}

	// Check against 2 letter repeating patterns
	if checkTwoLetterRepetition(cleanedStr) {
		return "", nil, false
	}

	// Check for more than 2 of the same letter in a row
	if checkRepeats(cleanedStr) {
		return "", nil, false
	}

	// If the string starts and ends with the same character, with a single repeated character in the middle, return false
	if len(charSet) == 2 && cleaned[0] == cleaned[len(cleaned)-1] {
		midChar := cleaned[1]
		for _, r := range cleaned[2 : len(cleaned)-1] {
			if r != midChar {
				break
			}
		}
		return "", nil, false
	}

	// Check palindrome property
	a, b := 0, len(cleaned)-1
	for a < b {
		if cleaned[a] != cleaned[b] {
			return "", nil, false
		}
		a++
		b--
	}

	// // Tokenize and check for valid words
	results, ok := validator.IsValidWordSegment(cleanedStr)
	if !ok {
		return "", nil, false
	}

	// if !canSegment(cleanedStr, dictionary) {
	// 	return "", false
	// }

	return string(original), results, true
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

// check for more than 2 of a single character in a row
func checkRepeats(s string) bool {
	l := 0

	for i := range len(s) {
		if i > l+2 {
			return true
		}
		if s[l] != s[i] {
			l = i
		}
	}

	return false
}

// WordSegmentValidator contains a dictionary of valid words
type WordSegmentValidator struct {
	dictionary map[string]struct{}
}

// NewWordSegmentValidator creates a new validator with the given dictionary
func NewWordSegmentValidator(dictionary map[string]struct{}) *WordSegmentValidator {
	return &WordSegmentValidator{
		dictionary: dictionary,
	}
}

// // IsValidWordSegment checks if the input string can be segmented into valid dictionary words
// func (v *WordSegmentValidator) IsValidWordSegment(s string) bool {
// 	if len(s) == 0 {
// 		return true
// 	}

// 	// dp[i] represents whether the substring s[0:i] can be segmented into valid words
// 	dp := make([]bool, len(s)+1)
// 	dp[0] = true

// 	for i := 1; i <= len(s); i++ {
// 		for j := 0; j < i; j++ {
// 			if dp[j] {
// 				if _, ok := v.dictionary[s[j:i]]; ok {
// 					dp[i] = true
// 					break
// 				}
// 			}
// 		}
// 	}

// 	return dp[len(s)]
// }

// IsValidWordSegment returns true if there are valid word segmentations and all valid combinations.
func (v *WordSegmentValidator) IsValidWordSegment(s string) ([]string, bool) {
	var results []string

	// Backtracking function to explore valid segmentations.
	var backtrack func(start int, path []string)
	backtrack = func(start int, path []string) {
		if start == len(s) {
			// Only append if the entire string is segmented into valid words.
			results = append(results, strings.Join(path, " "))
			return
		}

		// Explore all possible substrings from the current start position.
		for end := start + 1; end <= len(s); end++ {
			word := s[start:end]
			if _, exists := v.dictionary[word]; exists {
				// If it's a valid word, continue exploring.
				backtrack(end, append(path, word))
			}
		}
	}

	// Start backtracking from the beginning of the string.
	backtrack(0, []string{})

	// Return true if at least one valid segmentation exists.
	return results, len(results) > 0
}

// CanSegment checks if the input string can be segmented into valid words
func canSegment(input string, dictionary map[string]struct{}) bool {
	n := len(input)
	dp := make([]bool, n+1)
	dp[0] = true

	// fmt.Println(dictionary)

	for i := 1; i <= n; i++ {
		for j := 0; j < i; j++ {
			word := input[j:i]

			// Only update if the prefix is valid and the current word is valid.
			if dp[j] && isValidWord(word, dictionary) {
				dp[i] = true
				break
			}
		}
	}

	fmt.Println(dp)
	return dp[n]
}

// isValidWord checks if a word exists in the dictionary
func isValidWord(word string, dictionary map[string]struct{}) bool {
	fmt.Println(word)
	_, exists := dictionary[word]
	return exists
}
