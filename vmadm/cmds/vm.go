package cmds

import "github.com/urfave/cli/v2"

var VmCmds = &cli.Command{
	Name:  "vm",
	Usage: "vm operation",
	Subcommands: []*cli.Command{
		vmCreate,
		vmDelete,
		vmStart,
		vmStop,
		vmList,
	},
}
