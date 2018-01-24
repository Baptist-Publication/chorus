package commands

import (
	"fmt"

	"github.com/Baptist-Publication/chorus-module/lib/go-crypto"
	"gopkg.in/urfave/cli.v1"
)

var (
	//AccountCommands defines a more git-like subcommand system
	AccountCommands = cli.Command{
		Name:     "account",
		Usage:    "operations for account",
		Category: "Account",
		Subcommands: []cli.Command{
			{
				Name:     "gen",
				Action:   generatePrivPubAddr,
				Usage:    "generate new private-pub key pair",
				Category: "Account",
			},
		},
	}
)

func generatePrivPubAddr(ctx *cli.Context) error {
	sk := crypto.GenPrivKeyEd25519()
	pk := sk.PubKey().(*crypto.PubKeyEd25519)

	fmt.Printf("privkey: %X\n", sk[:])
	fmt.Printf("pubkey: %X\n", pk[:])

	return nil
}
