package commands

import (
	"encoding/hex"
	"fmt"

	"github.com/Baptist-Publication/chorus/eth/crypto"
	gcommon "github.com/Baptist-Publication/chorus/module/lib/go-common"
	agcrypto "github.com/Baptist-Publication/chorus/module/lib/go-crypto"
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
			{
				Name:     "cal",
				Action:   calculatePrivPubAddr,
				Usage:    "calculate public key and address from private key",
				Category: "Account",
				Flags: []cli.Flag{
					anntoolFlags.privkey,
				},
			},
			{
				Name:     "geneth",
				Action:   generateSecpPrivPubAddr,
				Usage:    "generate new private-pub key pair",
				Category: "Account",
			},
		},
	}
)

func generatePrivPubAddr(ctx *cli.Context) error {
	sk := agcrypto.GenPrivKeyEd25519()
	pk := sk.PubKey().(*agcrypto.PubKeyEd25519)

	fmt.Printf("privkey: %X\n", sk[:])
	fmt.Printf("pubkey: %X\n", pk[:])

	return nil
}

func calculatePrivPubAddr(ctx *cli.Context) error {
	if !ctx.IsSet("privkey") {
		return cli.NewExitError("private key is required", -1)
	}

	skBs, err := hex.DecodeString(gcommon.SanitizeHex(ctx.String("privkey")))
	if err != nil {
		return cli.NewExitError(err.Error(), -1)
	}

	var sk agcrypto.PrivKeyEd25519
	copy(sk[:], skBs)

	pk := sk.PubKey().(*agcrypto.PubKeyEd25519)
	addr := pk.Address()

	fmt.Printf("pubkey : %X\n", pk[:])
	fmt.Printf("address: %X\n", addr)

	return nil
}

func generateSecpPrivPubAddr(ctx *cli.Context) error {
	sk, _ := crypto.GenerateKey()
	pk := crypto.FromECDSAPub(&sk.PublicKey)
	addr := crypto.PubkeyToAddress(sk.PublicKey)

	fmt.Printf("privkey: %X\n", crypto.FromECDSA(sk))
	fmt.Printf("pubkey: %X\n", pk[:])
	fmt.Printf("address : %X\n", addr[:])

	return nil
}
