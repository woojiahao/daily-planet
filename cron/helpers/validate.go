package helpers

import (
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
	return err == nil
}
