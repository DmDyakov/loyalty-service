package luhn

import "testing"

func TestValid(t *testing.T) {
	tests := []struct {
		number string
		valid  bool
	}{
		{"49927398716", true},
		{"12345678903", true},
		{"0", true},
		{"00", true},
		{"1234", false},
		{"", false},
		{"abc", false},
		{"1234a", false},
		{" 123", false},
	}

	for _, tt := range tests {
		t.Run(tt.number, func(t *testing.T) {
			got := Valid(tt.number)
			if got != tt.valid {
				t.Errorf("Valid(%q) = %v, want %v", tt.number, got, tt.valid)
			}
		})
	}
}
