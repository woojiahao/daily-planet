package apperrors

import "errors"

var (
	ErrConfigurationDBError      = errors.New("database operation on 'configuration' table failed")
	ErrConfigurationNotFound     = errors.New("configuration does not exist")
	ErrConfigurationUpdateFailed = errors.New("failed to update configuration")
)
