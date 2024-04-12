package validator

import (
	"testing"
)

func TestValid(t *testing.T) {
	v := New()
	want := true
	if got := v.Valid(); got != want {
		t.Errorf("Valid() = %v, want %v", got, want)
	}
}

func TestAddError(t *testing.T) {
	v := New()
	tests := []struct {
		name    string
		key     string
		message string
		want    int
	}{
		{"add new error", "first error", "first error message", 1},
		{"add existing error", "first error", "first error message", 1},
		{"add another error", "second error", "second error message", 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v.AddError(tt.key, tt.message)
			if got := len(v.Errors); got != tt.want {
				t.Errorf("AddError(%v, %v), want %v,  got %v", tt.key, tt.message, tt.want, got)
			}
		})
	}
}

func TestCheck(t *testing.T) {
	v := New()
	var str string
	var num int
	tests := []struct {
		name    string
		ok      bool
		key     string
		message string
		want    int
	}{
		{"passed validation 1", str == "", "", "", 0},
		{"passed validation 2", num == 0, "", "", 0},
		{"failed validation", str != "", "sample text", "must be provided", 1},
		{"failed validation 2", str != "", "sample text", "must be provided", 1},
		{"failed validation 3", num != 0, "sample num", "must be provided", 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v.Check(tt.ok, tt.key, tt.message)
			if got := len(v.Errors); got != tt.want {
				t.Errorf("Check(%v, %v, %v), want %v, got %v", tt.ok, tt.key, tt.message, tt.want, got)
			}
		})
	}
}

func TestMatches(t *testing.T) {
	tests := []struct {
		name  string
		email string
		want  bool
	}{
		{"valid", "test@example.com", true},
		{"invalid (no @ symbol)", "test.example.com", false},
		{"invalid (email ends with dot)", "test@example.com.", false},
		{"invalid (no prefix)", "@example.com", false},
		{"invalid (prefix contains space)", "te st@example.com", false},
		{"invalid (prefix contains two dots in a row)", "te..st@example", false},
		{"invalid (no TLD)", "test@example", false},
		{"invalid (domain ends with dot)", "test@example.", false},
		{"invalid (no domain, only TLD)", "test@.com", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Matches(tt.email, EmailRX); got != tt.want {
				t.Errorf("Matches(%v, %v) = %v, want %v", tt.email, EmailRX, got, tt.want)
			}
		})
	}
}

// func TestErrMessage(t *testing.T) {
// 	err := map[string]string{"first error": "first error message", "second error": "second error message"}
// 	want := "first error: first error message; second error: second error message."
// 	if got := ErrMessage(err); got.Error() != want {
// 		t.Errorf("ErrMEssage(%v), want %v, got %v", err, want, got)
// 	}
// }
