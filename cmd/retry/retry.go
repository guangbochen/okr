package retry

import (
	"time"

	"github.com/spf13/cobra"

	"github.com/oneblock-ai/okr/pkg/k3s/retry"
)

func NewRetry() *cobra.Command {
	r := &Retry{}
	cmd := &cobra.Command{
		Use:                "retry",
		Short:              "Retry command until it succeeds",
		Hidden:             true,
		DisableFlagParsing: true,
		RunE:               r.Run,
	}
	return cmd
}

type Retry struct {
	SleepFirst bool `usage:"Sleep 5 seconds before running command"`
}

func (p *Retry) Run(cmd *cobra.Command, args []string) error {
	if p.SleepFirst {
		time.Sleep(5 * time.Second)
	}
	return retry.Retry(cmd.Context(), 15*time.Second, args)
}
