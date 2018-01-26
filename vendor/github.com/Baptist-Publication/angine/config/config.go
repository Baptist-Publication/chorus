// Copyright 2017 Baptist-Publication Information Technology Services Co.,Ltd.
//
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

package config

import (
	"os"
	"path"
	"strings"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"
	"github.com/spf13/viper"

	cmn "github.com/Baptist-Publication/chorus-module/lib/go-common"
)

const (
	// RUNTIME_ENV defines the name of the environment variable for runtime path
	RUNTIME_ENV = "ANGINE_RUNTIME"
	// DEFAULT_RUNTIME defines the default path for runtime path relative to $HOME
	DEFAULT_RUNTIME = ".angine"
	// DATADIR is the data dir in the runtime, basically you don't change this never
	DATADIR    = "data"
	ARCHIVEDIR = "archive"
	// CONFIGFILE is the name of the configuration file name in the runtime path for angine
	CONFIGFILE = "config.toml"
)

var runtimePath string

func parseConfigTpl(moniker string, root string) (conf string) {
	conf = strings.Replace(CONFIGTPL, "__MONIKER__", moniker, -1)
	// conf = strings.Replace(conf, "__CONFROOT__", root, -1)
	return
}

// RuntimeDir determines where the runtime directory is : param > environment > default
func RuntimeDir(root string) string {
	if root != "" {
		return root
	}
	if len(runtimePath) == 0 {
		runtimePath = os.Getenv(RUNTIME_ENV)
		if len(runtimePath) == 0 {
			runtimePath, _ = homedir.Dir()
		}
		runtimePath = path.Join(runtimePath, DEFAULT_RUNTIME)
	}
	return runtimePath
}

// InitRuntime makes all the necessary directorys for angine's runtime
// and generate the config template for you if it is not there already
func InitRuntime(root string) {
	cmn.EnsureDir(root, 0700)
	cmn.EnsureDir(path.Join(root, DATADIR), 0700)
	configFilePath := path.Join(root, CONFIGFILE)
	if !cmn.FileExists(configFilePath) {
		cmn.MustWriteFile(configFilePath, []byte(parseConfigTpl("anonymous", root)), 0644)
	}
}

// GetConfig returns a ready-to-go config instance with all defaults filled in
func GetConfig(root string) (conf *viper.Viper) {
	var err error
	runtime := RuntimeDir(root)
	InitRuntime(runtime)

	conf = viper.New()
	conf.SetEnvPrefix("ANGINE")
	conf.SetConfigFile(path.Join(runtime, CONFIGFILE))
	if err = conf.ReadInConfig(); err != nil {
		cmn.Exit(errors.Wrap(err, "angine configuration").Error())
	}
	if conf.IsSet("chain_id") {
		cmn.Exit("Cannot set 'chain_id' via config.toml")
	}
	if conf.IsSet("revision_file") {
		cmn.Exit("Cannot set 'revision_file' via config.toml. It must match what's in the Makefile")
	}

	SetDefaults(runtime, conf)

	return
}

// SetDefaults sets all the default configs for angine
func SetDefaults(runtime string, conf *viper.Viper) *viper.Viper {
	conf.SetDefault("environment", "development")
	conf.SetDefault("runtime", runtime)
	conf.SetDefault("genesis_file", path.Join(runtime, "genesis.json"))
	conf.SetDefault("moniker", "anonymous")
	conf.SetDefault("p2p_laddr", "tcp://0.0.0.0:46656")
	conf.SetDefault("seeds", "")
	conf.SetDefault("auth_by_ca", false)              // auth by ca general switch
	conf.SetDefault("non_validator_auth_by_ca", true) // whether non-validator nodes need auth by ca, only effective when auth_by_ca is true
	conf.SetDefault("fast_sync", true)
	conf.SetDefault("skip_upnp", true)
	conf.SetDefault("addrbook_file", path.Join(runtime, "addrbook.json"))
	conf.SetDefault("addrbook_strict", false) // disable to allow connections locally
	conf.SetDefault("pex_reactor", true)      // enable for peer exchange
	conf.SetDefault("priv_validator_file", path.Join(runtime, "priv_validator.json"))
	conf.SetDefault("db_backend", "leveldb")
	conf.SetDefault("db_dir", path.Join(runtime, DATADIR))
	conf.SetDefault("db_archive_dir", path.Join(runtime, ARCHIVEDIR))
	conf.SetDefault("revision_file", path.Join(runtime, "revision"))
	conf.SetDefault("filter_peers", false)

	conf.SetDefault("signbyCA", "") // auth signature from CA
	conf.SetDefault("log_path", "")
	conf.SetDefault("threshold_blocks", 900)

	conf.SetDefault("enable_incentive", false)
	setMempoolDefaults(conf)
	setConsensusDefaults(conf)

	return conf
}

func setMempoolDefaults(conf *viper.Viper) {
	conf.SetDefault("mempool_broadcast", true)
	conf.SetDefault("mempool_wal_dir", path.Join(conf.GetString("runtime"), DATADIR, "mempool.wal"))
	conf.SetDefault("mempool_recheck", false)
	conf.SetDefault("mempool_recheck_empty", false)
	conf.SetDefault("mempool_enable_txs_limits", false)
}

func setConsensusDefaults(conf *viper.Viper) {
	conf.SetDefault("cs_wal_dir", path.Join(conf.GetString("runtime"), DATADIR, "cs.wal"))
	conf.SetDefault("cs_wal_light", false)
	conf.SetDefault("block_max_txs", 5000)         // max number of txs
	conf.SetDefault("block_max_size", 2*1024*1024) // max size of block(just for txs)
	conf.SetDefault("block_part_size", 65536)      // part size 64K
	conf.SetDefault("disable_data_hash", false)
	conf.SetDefault("timeout_propose", 5000)
	conf.SetDefault("timeout_propose_delta", 500)
	conf.SetDefault("timeout_prevote", 3000)
	conf.SetDefault("timeout_prevote_delta", 500)
	conf.SetDefault("timeout_precommit", 3000)
	conf.SetDefault("timeout_precommit_delta", 500)
	conf.SetDefault("timeout_commit", 10000)
	conf.SetDefault("skip_timeout_commit", false)

	conf.SetDefault("tracerouter_msg_ttl", 5) // seconds
}
