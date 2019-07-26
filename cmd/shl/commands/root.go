package commands

import (
	"github.com/abassian/shuffle/cmd/shl/commands/keys"
	"github.com/abassian/shuffle/cmd/shl/commands/run"
	"github.com/spf13/cobra"
)

//RootCmd is the root command for shl
var RootCmd = &cobra.Command{
	Use:   "shl",
	Short: "Shuffle",
}

func init() {
	RootCmd.AddCommand(
		run.RunCmd,
		keys.KeysCmd,
		VersionCmd,
	)
	//do not print usage when error occurs
	RootCmd.SilenceUsage = true
}
