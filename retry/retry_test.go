// #############################################################################
// # File: retry_test.go                                                       #
// # Project: retry                                                            #
// # Created Date: 2024/10/15 13:08:37                                         #
// # Author: realjf                                                            #
// # -----                                                                     #
// # Last Modified: 2024/10/15 13:08:59                                        #
// # Modified By: realjf                                                       #
// # -----                                                                     #
// #                                                                           #
// #############################################################################
package retry_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/realjf/utils/retry"
)

func exampleOperation() error {
	if time.Now().Second()%9 != 0 {
		return errors.New("Transient error occurred")
	}
	return nil
}

func TestRetryBackoff(t *testing.T) {
	maxRetries := 5
	retryDelay := 3
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	if err := retry.RetryWithBackoff(ctx, maxRetries, retryDelay, exampleOperation); err != nil {
		t.Fatal(err)
		return
	}
	t.Log("done")
}
