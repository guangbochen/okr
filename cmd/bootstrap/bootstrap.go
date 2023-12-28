package bootstrap

import (
	"github.com/spf13/cobra"

	"github.com/oneblock-ai/okr/pkg/okr"
)

func NewBootstrap() *cobra.Command {
	b := Bootstrap{}
	cmd := &cobra.Command{
		Use:   "bootstrap [flags]",
		Short: "Bootstrap a Kubernetes(k3s) and KubeRay cluster",
		RunE:  b.Run,
	}
	b.init(cmd)
	return cmd
}

type Bootstrap struct {
	Force bool `usage:"Run bootstrap even if already bootstrapped" short:"f"`
}

func (b *Bootstrap) Run(cmd *cobra.Command, args []string) error {
	r := okr.New(okr.Config{
		Force:      b.Force,
		DataDir:    okr.DefaultDataDir,
		ConfigPath: okr.DefaultConfigFile,
	})
	return r.Run(cmd.Context())
}

func (b *Bootstrap) init(apiCmd *cobra.Command) {
	f := apiCmd.Flags()
	f.BoolVar(&b.Force, "force", false, "Run bootstrap even if already bootstrapped")
}
