package api

import "testing"

func TestPalindrome(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"racecar", true},
		{"hello", false},
		{"hahahaha", false},
		{"", false},
		{"ahahahah", false},
		{"ahahaha", false},
		{"wooooow", false},
		{"JAJAJA", false},
		{"JAJAJAJAJAJ", false},
		{"YAYAYAYAY", false},
		{"🤣🤣🤣🤣🤣", false},
		{"RA-Cecar", true},
		{"LOOOOOOOOL", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := Palindrome(tt.input)
			if got != tt.want {
				t.Errorf("Palindrome(%q) = %v; want %v", tt.input, got, tt.want)
			}
		})
	}
}
