package main

import "testing"

func TestIsBlank(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"empty", "", true},
		{"spaces", "   ", true},
		{"tabsAndNewline", "\t\n", true},
		{"nonBlank", "data", false},
		{"surroundedBySpace", "  value  ", false},
	}

	for _, tt := range tests {
		caseData := tt
		t.Run(caseData.name, func(t *testing.T) {
			if got := IsBlank(caseData.input); got != caseData.want {
				t.Fatalf("IsBlank(%q) = %v, want %v", caseData.input, got, caseData.want)
			}
		})
	}
}
