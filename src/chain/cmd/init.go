// Copyright © 2017 ZhongAn Technology
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
	"path"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/Baptist-Publication/angine"
	cmn "github.com/Baptist-Publication/chorus-module/lib/go-common"
)

const (
	CONFPATH = "confpath"
)

// abcCmd represents the abc command
var initCmd = &cobra.Command{
	Use:       "init",
	Short:     "create config, genesis, priv files in the runtime directory",
	Long:      `create config, genesis, priv files in the runtime directory`,
	Args:      cobra.OnlyValidArgs,
	ValidArgs: []string{"chainid"},
	Run: func(cmd *cobra.Command, args []string) {
		chainID := cmd.Flag("chainid").Value.String()
		initCivilConfig(viper.GetString("config"))
		angine.Initialize(&angine.Tunes{
			Runtime: viper.GetString("runtime"),
		}, chainID)
	},
}

func init() {

	initCmd.Flags().String("chainid", "", "name of the chain")
	RootCmd.AddCommand(initCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// abcCmd.PersistentFlags().String("foo", "", "A help for foo")
	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// abcCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func initCivilConfig(pathStr string) {
	if pathStr == "" {
		pathStr = CivilPath()
	}

	var dirpath, filepath string
	if path.Ext(pathStr) == "" {
		dirpath = pathStr
		filepath = path.Join(pathStr, ".chorus.toml")
	} else {
		dirpath = path.Dir(pathStr)
		filepath = pathStr
	}

	if err := cmn.EnsureDir(dirpath, 0700); err != nil {
		cmn.PanicSanity(err)
	}
	if !cmn.FileExists(filepath) {
		cmn.MustWriteFile(filepath, []byte(configTemplate), 0644)
		fmt.Println("path of the chorus config file: " + filepath)
	}
}

var civilPath string

func CivilPath() string {
	if len(civilPath) == 0 {
		civilPath = viper.GetString(CONFPATH)
		if len(civilPath) == 0 {
			civilPath, _ = homedir.Dir()
		}
	}
	return civilPath
}

var configTemplate = `#toml configuration for chorus
environment = "production"                # log mode, e.g. "development"/"production"
p2p_laddr = "0.0.0.0:46656"               # p2p port that this node is listening
rpc_laddr = "0.0.0.0:46657"               # rpc port this node is exposing
event_laddr = "0.0.0.0:46658"             # chorus uses a exposed port for events function
log_path = ""                             # 
seeds = ""                                # peers to connect when the node is starting
signbyCA = ""                             # you must require a signature from a valid CA if the blockchain is a permissioned blockchain
enable_incentive = true					  # to enable block coinbase transaction in angine
`
