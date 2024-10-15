// #############################################################################
// # File: retry.go                                                            #
// # Project: retry                                                            #
// # Created Date: 2024/10/15 13:08:05                                         #
// # Author: realjf                                                            #
// # -----                                                                     #
// # Last Modified: 2024/10/15 13:08:16                                        #
// # Modified By: realjf                                                       #
// # -----                                                                     #
// #                                                                           #
// #############################################################################
package retry

import (
	"context"
	"fmt"
	"math/rand"
	"time"
)

func RetryWithBackoff(ctx context.Context, maxRetries int, retryDelaySeconds int, operation func() error) (err error) {

	stime := time.Now()
	rndEng := rand.New(rand.NewSource(stime.UnixNano()))
	for attempt := 0; attempt < maxRetries; attempt++ {
		fmt.Printf("Attempt %d[%f]\n", attempt+1, time.Now().Sub(stime).Abs().Seconds())

		select {
		case <-ctx.Done():
			return ctx.Err()
		default:

		}

		if err = operation(); err == nil {
			fmt.Println("operation succeeded")
			return
		}

		fmt.Printf("Error: %s\n", err)
		sec := rndEng.Intn(3)
		time.Sleep(time.Duration(retryDelaySeconds*(attempt+1)+sec) * time.Second)
	}

	return fmt.Errorf("Max retries reached. Last error: %s\n", err)
}
