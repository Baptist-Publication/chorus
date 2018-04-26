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

package config

const CONFIGTPL = `# This is a TOML config file.
# For more information, see https://github.com/toml-lang/toml
environment = "production"                # log mode, e.g. "development"/"production"
p2p_laddr = "tcp://0.0.0.0:46656"               # p2p port that this node is listening
rpc_laddr = "tcp://0.0.0.0:46657"               # rpc port this node is exposing
event_laddr = "tcp://0.0.0.0:46658"             # chorus uses a exposed port for events function
log_path = ""                             #
seeds = ""                                # peers to connect when the node is starting
signbyCA = ""                             # you must require a signature from a valid CA if the blockchain is a permissioned blockchain
enable_incentive = true                   # to enable block coinbase transaction in angine

db_backend = "leveldb"
moniker = "__MONIKER__"

auth_by_ca = false
`
