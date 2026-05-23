package apperrors

import "errors"

var ErrCacheDBError = errors.New("database operation on 'cache' table failed")
