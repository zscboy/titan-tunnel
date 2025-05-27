package cmds

import "github.com/urfave/cli/v2"

var NodeCmds = &cli.Command{
	Name:  "node",
	Usage: "node manager node",
	Subcommands: []*cli.Command{
		nodeList,
	},
}
