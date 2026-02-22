package utils

import (
	"testing"
)

func TestMD5(t *testing.T) {
	testCases := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "empty string",
			input: "",
			want:  "d41d8cd98f00b204e9800998ecf8427e",
		},
		{
			name:  "hello world",
			input: "hello world",
			want:  "5eb63bbbe01eeed093cb22bb8f5acdc3",
		},
		{
			name:  "test data",
			input: "test data 123",
			want:  "3ea8b8314067daab252a9ed2c5783cc2",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := MD5(tc.input)
			if got != tc.want {
				t.Errorf("MD5() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestMD5Bytes(t *testing.T) {
	testCases := []struct {
		name  string
		input []byte
		want  string
	}{
		{
			name:  "empty bytes",
			input: []byte{},
			want:  "d41d8cd98f00b204e9800998ecf8427e",
		},
		{
			name:  "hello world bytes",
			input: []byte("hello world"),
			want:  "5eb63bbbe01eeed093cb22bb8f5acdc3",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := MD5Bytes(tc.input)
			if got != tc.want {
				t.Errorf("MD5Bytes() = %v, want %v", got, tc.want)
			}
		})
	}
}
