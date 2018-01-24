package start

import (
	"github.com/spf13/viper"

	"github.com/Baptist-Publication/angine"
)

func Initfiles(conf *viper.Viper) {
	angine.Initialize(&angine.Tunes{Conf: conf},"")
}