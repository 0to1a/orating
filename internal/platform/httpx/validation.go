package httpx

import (
	"fmt"
	"regexp"
	"strings"
)

// Permissive on purpose — accepts single-label domains like `user@localhost`
// for dev/test. Real validation is via the OTP flow, not regex.
var emailRE = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+$`)

func IsValidEmail(s string) bool {
	return emailRE.MatchString(strings.TrimSpace(s))
}

func Required(field, s string) error {
	if strings.TrimSpace(s) == "" {
		return fmt.Errorf("%s is required: %w", field, ErrValidation)
	}
	return nil
}

func MaxLen(field, s string, max int) error {
	if len(s) > max {
		return fmt.Errorf("%s must be at most %d characters: %w", field, max, ErrValidation)
	}
	return nil
}

func MinLen(field, s string, min int) error {
	if len(s) < min {
		return fmt.Errorf("%s must be at least %d characters: %w", field, min, ErrValidation)
	}
	return nil
}
