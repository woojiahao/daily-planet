package helpers

import "github.com/adhocore/gronx"

func IsValidCron(cronSchedule string) bool {
	gron := gronx.New()
	return gron.IsValid(cronSchedule)
}
