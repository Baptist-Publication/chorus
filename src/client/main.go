package main

import (
	"os"

	"github.com/Baptist-Publication/chorus/src/client/commands"
	"github.com/Baptist-Publication/chorus/src/client/commons"
	"gopkg.in/urfave/cli.v1"
)

func main() {
	app := cli.NewApp()
	app.Name = "chorustool"
	app.Version = "0.2"

	app.Commands = []cli.Command{
		commands.AccountCommands,
		commands.QueryCommands,
		commands.TxCommands,
		commands.InfoCommand,
		commands.EVMCommands,
		commands.ShareCommands,
	}

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "callmode",
			Usage:       "rpc call mode: sync or commit",
			Value:       "sync",
			Destination: &commons.CallMode,
		},
		cli.StringFlag{
			Name:        "backend",
			Value:       "tcp://localhost:46657",
			Destination: &commons.QueryServer,
			Usage:       "rpc address of the node",
		},
		cli.StringFlag{
			Name:  "target",
			Value: "",
			Usage: "specify the target chain for the following command",
		},
	}

	app.Before = func(ctx *cli.Context) error {
		if commons.CallMode == "sync" || commons.CallMode == "commit" {
			return nil
		}

		return cli.NewExitError("invalid sync mode", 127)
	}

	_ = app.Run(os.Args)
}
