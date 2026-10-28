package util

import "testing"

func TestValidateTrimRange(t *testing.T) {
	// Empty means "no trim": the input placeholder says so, so it must stay valid.
	if err := ValidateTrimRange(""); err != nil {
		t.Errorf("empty range should mean no trim, got %v", err)
	}

	valid := []string{
		"01:20-03:45",
		"0:30-1:00",
		"00:00:10-00:00:20",
		"30-90",
		"1:20-3:45",
		"10:00-10:30",
	}
	for _, in := range valid {
		if err := ValidateTrimRange(in); err != nil {
			t.Errorf("ValidateTrimRange(%q) = %v, want nil", in, err)
		}
	}

	invalid := []string{
		"abc",
		"1:20",
		"01:20-",
		"-03:45",
		"01:20-03:45-",
		"1:2-3:4",
		"01:20:30:40-01:20:50:60",
		"01:20 - 03:45",
		"01:20--03:45",
		"03:45-01:20", // reversed: would download nothing
		"01:20-01:20", // zero length
	}
	for _, in := range invalid {
		if err := ValidateTrimRange(in); err == nil {
			t.Errorf("ValidateTrimRange(%q) = nil, want an error", in)
		}
	}
}
