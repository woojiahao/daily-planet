package apperrors

import "errors"

var (
	ErrCronEngineConfigurationNotFound  = errors.New("configuration not scheduled")
	ErrCronEngineInvalidCronString      = errors.New("invalid cron string")
	ErrCronEngineScheduleAlreadyRunning = errors.New("schedule for configuration is already running. call Cancel on it first")
	ErrCronEngineScheduleError          = errors.New("failed to schedule configuration on cron engine")
)
