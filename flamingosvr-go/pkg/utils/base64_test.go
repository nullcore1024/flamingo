package utils

import (
	"testing"
)

func TestBase64Encode(t *testing.T) {
	testCases := []struct {
		name  string
		input []byte
		want  string
	}{
		{
			name:  "empty string",
			input: []byte{},
			want:  "",
		},
		{
			name:  "hello world",
			input: []byte("hello world"),
			want:  "aGVsbG8gd29ybGQ=",
		},
		{
			name:  "test data",
			input: []byte("test data 123"),
			want:  "dGVzdCBkYXRhIDEyMw==",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := Base64Encode(tc.input)
			if got != tc.want {
				t.Errorf("Base64Encode() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestBase64Decode(t *testing.T) {
	testCases := []struct {
		name  string
		input string
		want  []byte
		error bool
	}{
		{
			name:  "empty string",
			input: "",
			want:  []byte{},
			error: false,
		},
		{
			name:  "hello world",
			input: "aGVsbG8gd29ybGQ=",
			want:  []byte("hello world"),
			error: false,
		},
		{
			name:  "invalid base64",
			input: "invalid",
			want:  nil,
			error: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Base64Decode(tc.input)
			if tc.error && err == nil {
				t.Errorf("Base64Decode() should return error for input %v", tc.input)
			}
			if !tc.error && err != nil {
				t.Errorf("Base64Decode() returned error: %v", err)
			}
			if !tc.error && string(got) != string(tc.want) {
				t.Errorf("Base64Decode() = %v, want %v", string(got), string(tc.want))
			}
		})
	}
}
