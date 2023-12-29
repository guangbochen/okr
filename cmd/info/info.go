package info

import (
	"github.com/spf13/cobra"

	"github.com/oneblock-ai/okr/pkg/okr"
)

func NewInfo() *cobra.Command {
	b := Info{}
	return &cobra.Command{
		Use:   "info",
		Short: "Print installation versions",
		RunE:  b.Run,
	}
}

type Info struct {
}

func (b *Info) Run(cmd *cobra.Command, args []string) error {
	o := okr.New(okr.Config{
		DataDir:    okr.DefaultDataDir,
		ConfigPath: okr.DefaultConfigFile,
	})
	return o.Info(cmd.Context())
}
