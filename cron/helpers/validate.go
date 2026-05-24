package helpers

import (
	"strings"

	"github.com/robfig/cron/v3"
)

var parser = cron.NewParser(
	cron.Second |
		cron.Minute |
		cron.Hour |
		cron.Dom |
		cron.Month |
		cron.Dow,
)

func IsValidCron(cronSchedule string) bool {
	_, err := parser.Parse(cronSchedule)
	if err != nil {
		return false
	}

	if strings.HasPrefix(cronSchedule, "*") {
		// explicitly disallow schedules where the seconds is *
		return false
	}

	return true
}
