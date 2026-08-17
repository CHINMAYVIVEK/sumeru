package scheduler

import "context"

// ExecuteCronForTest runs one cron tick handler (tests).
func ExecuteCronForTest(ctx context.Context, id int64, name, eventName, code string) {
	executeCron(ctx, id, name, eventName, code)
}
