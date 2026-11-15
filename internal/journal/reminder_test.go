package journal

import (
	"strings"
	"testing"
	"unicode/utf8"
)

func TestReminderText(t *testing.T) {
	for _, tt := range []struct{ name, input, want string }{
		{"empty", "  ", ""},
		{"trim without inventing", "  The blue seal is broken.  ", "The blue seal is broken."},
		{"bounded multilingual", strings.Repeat("雪", 201), strings.Repeat("雪", 200) + "…"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got := reminderText(tt.input)
			if got != tt.want || !utf8.ValidString(got) {
				t.Fatalf("reminder text=%q want %q", got, tt.want)
			}
		})
	}
}
