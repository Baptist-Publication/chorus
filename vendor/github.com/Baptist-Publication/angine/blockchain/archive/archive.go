// Copyright 2017 ZhongAn Information Technology Services Co.,Ltd.
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

package archive

import (
	"strconv"

	agtypes "github.com/Baptist-Publication/angine/types"
	dbm "github.com/Baptist-Publication/chorus-module/lib/go-db"
)

type Archive struct {
	db        dbm.DB
	Threshold agtypes.INT
}

var dbName = "archive"

func NewArchive(dbBackend, dbDir string, threshold agtypes.INT) *Archive {
	archiveDB := dbm.NewDB(dbName, dbBackend, dbDir)
	return &Archive{archiveDB, threshold}
}

func (ar *Archive) QueryFileHash(height agtypes.INT) (ret []byte) {
	origin := (height-1)/ar.Threshold*ar.Threshold + 1
	key := strconv.FormatInt(origin, 10) + "_" + strconv.FormatInt(origin-1+ar.Threshold, 10)
	ret = ar.db.Get([]byte(key))
	return
}

func (ar *Archive) AddItem(key, value string) {
	ar.db.SetSync([]byte(key), []byte(value))
}

func (ar *Archive) Close() {
	ar.db.Close()
}
