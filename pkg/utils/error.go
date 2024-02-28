package utils

import (
	"fmt"

	"github.com/getsentry/sentry-go"
	"github.com/m-mizutani/goerr"
)

func HandleError(msg string, err error) {
	// Sending error to Sentry
	hub := sentry.CurrentHub().Clone()
	hub.ConfigureScope(func(scope *sentry.Scope) {
		if goErr := goerr.Unwrap(err); goErr != nil {
			for k, v := range goErr.Values() {
				scope.SetExtra(fmt.Sprintf("%v", k), v)
			}
		}
	})
	evID := hub.CaptureException(err)

	logger.Error(msg, ErrLog(err), "sentry.EventID", evID)
}
