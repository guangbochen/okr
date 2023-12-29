package retry

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"time"

	"github.com/sirupsen/logrus"
)

func Retry(ctx context.Context, interval time.Duration, args []string) error {
	for {
		var stderr bytes.Buffer
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Stdout = os.Stdout
		cmd.Stdin = os.Stdin
		cmd.Stderr = os.Stderr
		cmd.Stderr = &stderr
		err := cmd.Run()
		if err != nil {
			logrus.Errorf("will retry failed command %v: %v, %v", args, err.Error(), stderr.String())
			select {
			case <-time.After(interval):
				continue
			case <-ctx.Done():
				return ctx.Err()
			}
		}
		return nil
	}
}
