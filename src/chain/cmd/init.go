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
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/Baptist-Publication/angine"
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
		// initCivilConfig(viper.GetString("config"))
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
