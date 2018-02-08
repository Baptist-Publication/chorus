// Copyright © 2017 Baptist-Publication Technology
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"
	"os"

	acfg "github.com/Baptist-Publication/angine/config"
	agtypes "github.com/Baptist-Publication/angine/types"
	crypto "github.com/Baptist-Publication/chorus-module/lib/go-crypto"
	libcrypto "github.com/Baptist-Publication/chorus-module/xlib/crypto"
	"github.com/Baptist-Publication/chorus/src/chain/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func init() {
	RootCmd.AddCommand(resetCmd)
	RootCmd.AddCommand(changePwdCmd)
}

var resetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Reset PrivValidator, clean the data and shards",
	Long:  ``,
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		angineconf := acfg.GetConfig(viper.GetString("runtime"))
		pwd, err := libcrypto.InputPasswdForDecrypt()
		if err != nil {
			cmd.Println(err)
			return
		}
		pwd2 := make([]byte, len(pwd))
		copy(pwd2, pwd)
		conf := config.GetConfig(viper.GetString("runtime"), pwd)
		os.RemoveAll(conf.GetString("db_dir"))
		os.RemoveAll(conf.GetString("shards"))
		resetPrivValidator(angineconf.GetString("priv_validator_file"), pwd2)
	},
}

var changePwdCmd = &cobra.Command{
	Use:   "resetpwd",
	Short: "Reset pwd of privkey's encryption",
	Long:  ``,
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		resetPwd(cmd)
	},
}

func resetPrivValidator(privValidatorFile string, pwd []byte) {
	var (
		privValidator *agtypes.PrivValidator
		logger        *zap.Logger
	)

	if _, err := os.Stat(privValidatorFile); err == nil {
		privValidator = agtypes.LoadPrivValidator(logger, privValidatorFile, pwd)
		privValidator.Reset()
		fmt.Println("Reset PrivValidator", "file", privValidatorFile)
	} else {
		privValidator, err = agtypes.GenPrivValidator(logger, pwd)
		if err != nil {
			panic(err)
		}
		privValidator.SetFile(privValidatorFile)
		privValidator.Save()
		fmt.Println("Generated PrivValidator", "file", privValidatorFile)
	}
}

func resetPwd(cmd *cobra.Command) {
	pwd, err := libcrypto.InputPasswdForDecrypt()
	if err != nil {
		cmd.Println(err)
		return
	}
	angineconf := acfg.GetConfig(viper.GetString("runtime"))
	filePath := angineconf.GetString("priv_validator_file")
	privValidator := agtypes.LoadPrivValidator(nil, filePath, pwd)
	if privValidator == nil {
		cmd.Println("priv_validator can't find")
		return
	}
	privEd, ok := privValidator.GetPrivKey().(*crypto.PrivKeyEd25519)
	if !ok {
		cmd.Println("Privkey's format error,not Ed25519")
		return
	}
	var pwd2 []byte
	pwd2, err = libcrypto.InputPasswdForEncrypt()
	if err != nil {
		cmd.Println(err)
		return
	}
	err = privEd.ChangePwd(pwd2)
	if err != nil {
		cmd.Println(err)
		return
	}
	privValidator.Save()
}
