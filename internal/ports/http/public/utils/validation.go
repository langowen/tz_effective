package utils

import (
	"fmt"
	"regexp"
)

var uuidRegex = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

func ValidateUUID(uuid string) error {
	if !uuidRegex.MatchString(uuid) {
		return fmt.Errorf("invalid UUID format: %s", uuid)
	}
	return nil
}

func ValidateDate(date string) error {
	dateRegex := regexp.MustCompile(`^(0[1-9]|1[0-2])-\d{4}$`)
	if !dateRegex.MatchString(date) {
		return fmt.Errorf("invalid date format: %s, expected format MM-YYYY", date)
	}
	return nil
}
