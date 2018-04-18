package config

import (
	"path"

	"github.com/spf13/viper"

	anginecfg "github.com/Baptist-Publication/angine/config"
	"github.com/Baptist-Publication/chorus-module/lib/go-common"
)

const (
	ShardsDir  = "shards"
	ShardsData = "data"
)

// LoadDefaultShardConfig should only be called after initChorusRuntime
func LoadDefaultAngineConfig(datadir, chainID string, conf map[string]interface{}) (*viper.Viper, error) {
	shardPath := path.Join(datadir, ShardsDir, chainID)
	if err := common.EnsureDir(shardPath, 0700); err != nil {
		return nil, err
	}
	if err := common.EnsureDir(path.Join(shardPath, ShardsData), 0700); err != nil {
		return nil, err
	}

	c := anginecfg.SetDefaults(shardPath, viper.New())
	for k, v := range conf {
		c.Set(k, v)
	}
	c.Set("chain_id", chainID)

	loadDefaultSqlConfig(c)
	loadDefaultEVMConfig(c)
	return c, nil
}

func loadDefaultSqlConfig(c *viper.Viper) {
	c.Set("db_type", "sqlite3")
	c.Set("db_conn_str", "")
}

func loadDefaultEVMConfig(c *viper.Viper) {
	c.Set("block_gaslimit", 80000000)
}
