package probe

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/oneblock-ai/okr/pkg/k3s/instructions/probe"
)

func NewProbe() *cobra.Command {
	p := Probe{}
	cmd := &cobra.Command{
		Use:   "probe",
		Short: "Run plan probes",
		RunE:  p.Run,
	}
	p.init(cmd)
	return cmd
}

type Probe struct {
	Interval string
	File     string
}

func (p *Probe) Run(cmd *cobra.Command, args []string) error {
	interval, err := time.ParseDuration(p.Interval)
	if err != nil {
		return fmt.Errorf("failed to parse duration %s: %w", p.Interval, err)
	}

	return probe.RunProbes(cmd.Context(), p.File, interval)
}

func (p *Probe) init(apiCmd *cobra.Command) {
	f := apiCmd.Flags()
	f.StringVar(&p.Interval, "interval", "5s", "Polling interval to run probes")
	f.StringVar(&p.File, "file", "/var/lib/oneblock-ai/okr/plan/plan.json", "Plan file")
}
