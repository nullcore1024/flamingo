package utils

import (
	"testing"
)

func TestTrim(t *testing.T) {
	testCases := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "empty string",
			input: "",
			want:  "",
		},
		{
			name:  "string with spaces",
			input: "  hello world  ",
			want:  "hello world",
		},
		{
			name:  "string without spaces",
			input: "hello",
			want:  "hello",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := Trim(tc.input)
			if got != tc.want {
				t.Errorf("Trim() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestIsEmpty(t *testing.T) {
	testCases := []struct {
		name  string
		input string
		want  bool
	}{
		{
			name:  "empty string",
			input: "",
			want:  true,
		},
		{
			name:  "string with spaces",
			input: "   ",
			want:  true,
		},
		{
			name:  "non-empty string",
			input: "hello",
			want:  false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := IsEmpty(tc.input)
			if got != tc.want {
				t.Errorf("IsEmpty() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestContains(t *testing.T) {
	testCases := []struct {
		name   string
		input  string
		substr string
		want   bool
	}{
		{
			name:   "substring exists",
			input:  "hello world",
			substr: "world",
			want:   true,
		},
		{
			name:   "substring does not exist",
			input:  "hello world",
			substr: "test",
			want:   false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := Contains(tc.input, tc.substr)
			if got != tc.want {
				t.Errorf("Contains() = %v, want %v", got, tc.want)
			}
		})
	}
}
