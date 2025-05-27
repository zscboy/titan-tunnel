package cmds

import "github.com/urfave/cli/v2"

var DownloadCmds = &cli.Command{
	Name:  "download",
	Usage: "download manager image download",
	Subcommands: []*cli.Command{
		downloadimage,
		downloadTaskDelete,
		downloadTaskList,
	},
}
