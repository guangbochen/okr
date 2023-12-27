package main

import (
	"os"

	"github.com/rancher/wrangler/v2/pkg/signals"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/oneblock-ai/okr/cmd/bootstrap"
	"github.com/oneblock-ai/okr/cmd/info"
	"github.com/oneblock-ai/okr/cmd/probe"
	"github.com/oneblock-ai/okr/cmd/retry"
)

type OKR struct {
}

func main() {
	cmd := New()

	ctx := signals.SetupSignalContext()
	if err := cmd.ExecuteContext(ctx); err != nil {
		logrus.Error(err.Error())
		os.Exit(1)
	}

}

func New() *cobra.Command {
	o := &OKR{}
	var rootCmd = &cobra.Command{
		Use:  "okr",
		Long: "OKR is a command line tool for Oneblock to bootstrap k3s and KubeRay cluster",
		RunE: o.Run,
	}

	rootCmd.AddCommand(
		bootstrap.NewBootstrap(),
		info.NewInfo(),
		probe.NewProbe(),
		retry.NewRetry(),
	)

	rootCmd.InitDefaultHelpCmd()
	return rootCmd
}

func (o *OKR) Run(cmd *cobra.Command, args []string) error {
	return cmd.Help()
}
