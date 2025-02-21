package api

import "testing"

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
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			word, got := Palindrome(tt.input)

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
