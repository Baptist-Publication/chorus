package commands

import (
	"fmt"

	"github.com/Baptist-Publication/chorus-module/lib/go-crypto"
	libcrypto "github.com/Baptist-Publication/chorus-module/xlib/crypto"
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
				Flags: []cli.Flag{
					anntoolFlags.passwd,
				},
			},
			{
				Name:     "cal",
				Action:   calculatePrivPubAddr,
				Usage:    "calculate public key and address from private key",
				Category: "Account",
				Flags: []cli.Flag{
					anntoolFlags.privkey,
					anntoolFlags.passwd,
				},
			},
		},
	}
)

func generatePrivPubAddr(ctx *cli.Context) error {
	var pwd []byte
	var err error
	if ctx.IsSet(anntoolFlags.passwd.GetName()) {
		pwd = []byte(ctx.String(anntoolFlags.passwd.GetName()))
	} else {
		pwd, err = libcrypto.InputPasswdForEncrypt()
		if err != nil {
			return nil
		}
	}
	sk, err := crypto.GenPrivKeyEd25519(pwd)
	if err != nil {
		return err
	}
	defer sk.Destroy()
	pk := sk.PubKey().(*crypto.PubKeyEd25519)

	fmt.Printf("ori-privkey: %X\n", sk.KeyBytes())
	fmt.Printf("privkey: %X\n", sk.Bytes())
	fmt.Printf("pubkey: %X\n", pk.Bytes())

	return nil
}

func calculatePrivPubAddr(ctx *cli.Context) error {
	privKey, err := ParsePrivkey(ctx)
	if err != nil {
		return cli.NewExitError(err.Error(), 127)
	}
	defer privKey.Destroy()
	pk := privKey.PubKey().(*crypto.PubKeyEd25519)
	addr := pk.Address()

	fmt.Printf("pubkey : %X\n", pk[:])
	fmt.Printf("address: %X\n", addr)

	return nil
}
