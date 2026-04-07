package helpers

import "time"

// DateStringToTimeWithLayout converts a raw date string (potentially empty) into time.Time following the given layout.
//
// Empty strings or ill-formatted date strings will return the current timestamp.
func dateStringToTimeWithLayout(dateString, layout string) time.Time {
	if dateString == "" {
		return time.Now()
	}

	if date, err := time.Parse(layout, dateString); err != nil {
		// If it fails, we should return current time since the date is usually optional anyways
		return time.Now()
	} else {
		return date
	}
}

// DateStringToRFC3339Time converts a raw date string (potentially empty) into time.Time following the RFC-3339 date format.
func DateStringToRFC3339Time(dateString string) time.Time {
	return dateStringToTimeWithLayout(dateString, time.RFC3339)
}

// DateStringToRFC822Time converts a raw date string (potentially empty) into time.Time following the RFC-822 date format.
//
// RFC: https://www.rfc-editor.org/rfc/rfc822.html#section-5.1
func DateStringToRFC822Time(dateString string) time.Time {
	// Due to some weird behavior, the format is mapped to the RFC-1123 specification in Go
	return dateStringToTimeWithLayout(dateString, time.RFC1123)
}
