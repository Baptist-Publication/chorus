package utils

import (
	"sync"

	"github.com/Baptist-Publication/chorus/module/xlib/def"
)

var data sync.Map

type AngineData struct {
	Height def.INT // LastBlockHeight of state
}

func UpdateHeight(chainID string, height def.INT) {
	d, _ := Load(chainID)
	d.Height = height
	Store(chainID, &d)
}

func LoadHeight(chainID string) def.INT {
	d, _ := Load(chainID)
	return d.Height
}

func Store(chainID string, d *AngineData) {
	data.Store(chainID, d)
}

func Load(chainID string) (AngineData, bool) {
	if pd, has := data.Load(chainID); has {
		return *(pd.(*AngineData)), true
	}
	return AngineData{}, false
}

func Delete(chainID string) {
	data.Delete(chainID)
}

func Has(chainID string) bool {
	_, has := data.Load(chainID)
	return has
}
