package cmds

import "github.com/urfave/cli/v2"

var ImageCmds = &cli.Command{
	Name:  "image",
	Usage: "image: list image, delete image",
	Subcommands: []*cli.Command{
		imageList,
		imageDelete,
	},
}
