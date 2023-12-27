package retry

import (
	"time"

	"github.com/spf13/cobra"

	"github.com/oneblock-ai/okr/pkg/retry"
)

func NewRetry() *cobra.Command {
	r := &Retry{}
	cmd := &cobra.Command{
		Use:   "retry",
		Short: "Retry command until it succeeds",
		RunE:  r.Run,
	}
	r.init(cmd)
	return cmd
}

type Retry struct {
	SleepFirst bool
}

func (p *Retry) Run(cmd *cobra.Command, args []string) error {
	if p.SleepFirst {
		time.Sleep(5 * time.Second)
	}
	return retry.Retry(cmd.Context(), 15*time.Second, args)
}

func (p *Retry) init(apiCmd *cobra.Command) {
	f := apiCmd.Flags()
	f.BoolVar(&p.SleepFirst, "s", false, "Sleep 5 seconds before running command")
}
