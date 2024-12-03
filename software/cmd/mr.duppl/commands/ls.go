package commands

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/buglloc/mr.duppl/software/pkg/dupplcap"
)

var lsArgs struct {
	Name bool
	Path bool
}

var lsCmd = &cobra.Command{
	Use:           "ls",
	SilenceUsage:  true,
	SilenceErrors: true,
	Short:         "List devices",
	RunE: func(_ *cobra.Command, _ []string) error {
		if !lsArgs.Name && !lsArgs.Path {
			lsArgs.Name = true
			lsArgs.Path = true
		}

		ifaces, err := dupplcap.Ifaces()
		if err != nil {
			return fmt.Errorf("unable to get information about interfaces: %w", err)
		}

		for _, iface := range ifaces {
			switch {
			case lsArgs.Name && lsArgs.Path:
				fmt.Println(iface.Name, iface.Path)
			case lsArgs.Name:
				fmt.Println(iface.Name)
			case lsArgs.Path:
				fmt.Println(iface.Path)
			}
		}

		return nil
	},
}

func init() {
	flags := lsCmd.PersistentFlags()
	flags.BoolVarP(&lsArgs.Name, "name", "n", false, "display device name")
	flags.BoolVarP(&lsArgs.Path, "path", "p", false, "display device path")
}
