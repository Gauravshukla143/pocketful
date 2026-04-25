package utils

import (
	"regexp"
)

// IsValidPAN validates a PAN card number format (e.g., ABCDE1234F).
func IsValidPAN(pan string) bool {
	re := regexp.MustCompile(`^[A-Z]{5}[0-9]{4}[A-Z]{1}$`)
	return re.MatchString(pan)
}

// IsValidAadhaar validates a 12-digit Aadhaar number.
func IsValidAadhaar(aadhaar string) bool {
	re := regexp.MustCompile(`^\d{12}$`)
	return re.MatchString(aadhaar)
}

// IsValidEmail validates basic email format.
func IsValidEmail(email string) bool {
	re := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	return re.MatchString(email)
}
