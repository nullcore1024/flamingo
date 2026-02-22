package utils

import (
	"testing"
)

func TestUUID(t *testing.T) {
	// 测试UUID生成是否成功
	uuid := UUID()
	if uuid == "" {
		t.Error("UUID() returned empty string")
	}

	// 测试生成的UUID是否有效
	if !IsValidUUID(uuid) {
		t.Error("UUID() returned invalid UUID")
	}
}

func TestUUIDV4(t *testing.T) {
	// 测试UUIDV4生成是否成功
	uuid := UUIDV4()
	if uuid == "" {
		t.Error("UUIDV4() returned empty string")
	}

	// 测试生成的UUID是否有效
	if !IsValidUUID(uuid) {
		t.Error("UUIDV4() returned invalid UUID")
	}
}

func TestIsValidUUID(t *testing.T) {
	testCases := []struct {
		name  string
		input string
		want  bool
	}{
		{
			name:  "valid UUID",
			input: "550e8400-e29b-41d4-a716-446655440000",
			want:  true,
		},
		{
			name:  "invalid UUID",
			input: "invalid-uuid",
			want:  false,
		},
		{
			name:  "empty string",
			input: "",
			want:  false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := IsValidUUID(tc.input)
			if got != tc.want {
				t.Errorf("IsValidUUID() = %v, want %v", got, tc.want)
			}
		})
	}
}
