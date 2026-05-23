package apperrors

import "errors"

var (
	ErrFeedDBError       = errors.New("database operation on 'feed' table failed")
	ErrFeedAlreadyExists = errors.New("failed to insert 'feed' row as it already exists")
	ErrFeedNotFound      = errors.New("feed does not exist")
	ErrFeedUpdateFailed  = errors.New("feed failed to update")
)
