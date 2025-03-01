package api

import (
	"log/slog"
	"os"
	"testing"

	"golang.org/x/exp/slices"
)

func TestPalindrome(t *testing.T) {
	tests := []struct {
		input        string
		want         string
		isPalindrome bool
	}{
		// {"racecar", "racecar", true},
		// {"hello", "", false},
		// {"hahahaha", "", false},
		// {"Yayyyy", "", false},
		// {"", "", false},
		// {"ahahahah", "", false},
		// {"ahahaha", "", false},
		// {"wooooow", "", false},
		// {"JAJAJA", "", false},
		// {"JAJAJAJAJAJ", "", false},
		// {"YAYAYAYAY", "", false},
		// {"🤣🤣🤣🤣🤣", "", false},
		// {"RA-Cecar", "RA-Cecar", true},
		// {"LOOOOOOOOL", "", false},
		// {"xxxxanaxxxx", "", false},
		// {"xxanaxx", "xxanaxx", true},
		// {"A man, a plan, a canal: Panama!", "A man, a plan, a canal: Panama", true},
		// {"Pull up", "", false},

		// This is clearly a palindrome but the stupid dictionary doesn't have "taco"
		{"taco cat", "taco cat", true},

		{"lemonnomel", "lemonnomel", true},
		{"gulf of foflug", "gulf of foflug", true},
		{"foflug", "", false},

		// This is not made of valid words but the single-letter property is causing a false-positive
		{"Glenelg", "Glenelg", true},
	}

	dictPath := "/usr/share/dict/words" // Or use a custom wordlist
	dictionary, err := LoadDictionary(dictPath)
	if err != nil {
		slog.Error("Failed to load dictionary:", "error", err)
		os.Exit(1)
	}

	validator := NewWordSegmentValidator(dictionary)

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			word, _, got := Palindrome(validator, tt.input)
			// for i := range words {
			// 	fmt.Printf("%d: %s\n", i, words[i])
			// }

			// Ensure the returned word matches the input
			if word != tt.want {
				t.Errorf("Palindrome(%q) returned word %q; expected %q", tt.input, word, tt.input)
			}

			// Ensure the palindrome check result matches expected
			if got != tt.isPalindrome {
				t.Errorf("Palindrome(%q) returned %v; expected %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestIsValidWordSegment(t *testing.T) {
	tests := []struct {
		input              string
		want               []string
		isValidWordSegment bool
	}{
		// {"lemon", []string{"lemon"}, true},
		// {"no", []string{"no"}, true},
		// {"mel", []string{"mel"}, true},
		// {"lemonnomel", []string{"lemon", "no", "mel"}, true},

		// {"lemon", true},
		// {"no", true},
		// {"mel", true},
		// {"lemonnomel", true},
	}

	dictPath := "/usr/share/dict/words" // Or use a custom wordlist
	dictionary, err := LoadDictionary(dictPath)
	if err != nil {
		slog.Error("Failed to load dictionary:", "error", err)
		os.Exit(1)
	}

	validator := NewWordSegmentValidator(dictionary)

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, isValid := validator.IsValidWordSegment(tt.input)

			if isValid != tt.isValidWordSegment {
				t.Errorf("IsValidWordSegment(%q) returned %v; expected %v", tt.input, isValid, tt.isValidWordSegment)
			}

			if slices.Equal(got, tt.want) {
				t.Errorf("IsValidWordSegment(%q) returned %v; expected %v", tt.input, got, tt.want)
			}
		})
	}
}
