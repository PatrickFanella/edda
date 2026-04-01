package llmutil

import "testing"

func TestStripMarkdownFences(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "no fence returns input unchanged",
			in:   `{"key":"value"}`,
			want: `{"key":"value"}`,
		},
		{
			name: "json fence stripped",
			in:   "```json\n{\"key\":\"value\"}\n```",
			want: `{"key":"value"}`,
		},
		{
			name: "plain fence stripped",
			in:   "```\n{\"key\":\"value\"}\n```",
			want: `{"key":"value"}`,
		},
		{
			name: "nested backticks only outermost stripped",
			in:   "```json\n{\"code\":\"use ```triple``` here\"}\n```",
			want: "{\"code\":\"use ```triple``` here\"}",
		},
		{
			name: "whitespace around fence trimmed",
			in:   "  \n```json\n{\"a\":1}\n```\n  ",
			want: `{"a":1}`,
		},
		{
			name: "empty input returns empty",
			in:   "",
			want: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := StripMarkdownFences(tc.in)
			if got != tc.want {
				t.Errorf("StripMarkdownFences()\ngot:  %q\nwant: %q", got, tc.want)
			}
		})
	}
}
