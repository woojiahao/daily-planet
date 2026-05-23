package apperrors

import "errors"

var (
	ErrLoadFeedFailed           = errors.New("failed to load feed")
	ErrLoadFeedUnsuppportedType = errors.New("feed not of type RSS, Atom, or JSON")
)
