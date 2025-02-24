package api

import (
	"log/slog"
	"os"
	"testing"
)

func TestPalindrome(t *testing.T) {
	tests := []struct {
		input        string
		want         string
		isPalindrome bool
	}{
		{"racecar", "racecar", true},
		{"hello", "", false},
		{"hahahaha", "", false},
		{"Yayyyy", "", false},
		{"", "", false},
		{"ahahahah", "", false},
		{"ahahaha", "", false},
		{"wooooow", "", false},
		{"JAJAJA", "", false},
		{"JAJAJAJAJAJ", "", false},
		{"YAYAYAYAY", "", false},
		{"🤣🤣🤣🤣🤣", "", false},
		{"RA-Cecar", "RA-Cecar", true},
		{"LOOOOOOOOL", "", false},
		{"xxxxanaxxxx", "", false},
		{"xxanaxx", "xxanaxx", true},
		{"taco cat", "taco cat", true},
		// {"lemonnomel", "", false},
		{"A man, a plan, a canal: Panama!", "A man, a plan, a canal: Panama", true},
		{"gulf of foflug", "gulf of foflug", true},
		{"foflug", "", false},
		// {"Glenelg", "", false},
		{"Pull up", "", false},
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
