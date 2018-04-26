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
	"os"
	"path"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/Baptist-Publication/chorus/chain/log"
	"github.com/Baptist-Publication/chorus/chain/node"
)

// nodeCmd represents the node command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run a blockchain full-capacity node",
	Long:  ``,
	Args:  cobra.NoArgs,
	PreRun: func(cmd *cobra.Command, args []string) {

	},
	Run: func(cmd *cobra.Command, args []string) {
		env := viper.GetString("environment")
		logpath := viper.GetString("log_path")
		if logpath == "" {
			var err error
			if logpath, err = os.Getwd(); err != nil {
				cmd.Println(err)
				os.Exit(1)
			}
		}
		viper.Set("log_path", logpath)
		logger := log.Initialize(env, path.Join(logpath, "node.output.log"), path.Join(logpath, "node.err.log"))
		node.RunNode(logger, viper.GetViper())
	},
}

func init() {
	RootCmd.AddCommand(runCmd)

	runCmd.Flags().StringP("chain_id", "", "", "specify the chain id when the node is joining it without genesis file")
	runCmd.Flags().BoolP("pprof", "", false, "start golang profile at port :6060")
	runCmd.Flags().BoolP("test", "", false, "run the node in test mode")

	viper.BindPFlag("chain_id", runCmd.Flag("chain_id"))
	viper.BindPFlag("pprof", runCmd.Flag("pprof"))
	viper.BindPFlag("test", runCmd.Flag("test"))
}
