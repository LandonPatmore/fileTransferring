package shared

import "testing"

func TestCheckByteArrayEquality(t *testing.T) {
	v := BlockNumberChecker([] byte{0, 0}, [] byte{0, 0})

	if !v {
		t.Errorf("Got = %v; want true", v)
	}
}