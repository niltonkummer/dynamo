package dynamo

import (
	"context"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/smithy-go"
	"github.com/aws/smithy-go/transport/http"
	"github.com/cenkalti/backoff"
)

// RetryTimeout defines the maximum amount of time that requests will
// attempt to automatically retry for. In other words, this is the maximum
// amount of time that dynamo operations will block.
// RetryTimeout is only considered by methods that do not take a context.
// Higher values are better when using tables with lower throughput.
var RetryTimeout = 1 * time.Minute

func defaultContext() (aws.Context, context.CancelFunc) {
	if RetryTimeout == 0 {
		return aws.BackgroundContext(), (func() {})
	}
	return context.WithDeadline(aws.BackgroundContext(), time.Now().Add(RetryTimeout))
}

func retry(ctx aws.Context, f func() error) error {
	var err error
	var next time.Duration
	b := backoff.WithContext(backoff.NewExponentialBackOff(), ctx)
	for {
		if err = f(); err == nil {
			return nil
		}

		if !canRetry(err) {
			return err
		}

		if next = b.NextBackOff(); next == backoff.Stop {
			return err
		}

		if err = aws.SleepWithContext(ctx, next); err != nil {
			return err
		}
	}
}

func canRetry(err error) bool {
	var resErr *http.ResponseError
	if errors.As(err, &resErr) {
		switch resErr.HTTPStatusCode() {
		case 500, 503:
			return true
		case 400:
			var apiErr smithy.APIError
			if errors.As(err, &apiErr) {
				switch apiErr.ErrorCode() {
				case "ProvisionedThroughputExceededException",
					"ThrottlingException":
					return true
				}
			}
		}
	}
	return false
}
