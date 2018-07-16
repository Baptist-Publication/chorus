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

package p2p

import (
	"github.com/spf13/viper"
)

const (
	// Switch config keys
	ConfigKeyDialTimeoutSeconds      = "dial_timeout_seconds"
	ConfigKeyHandshakeTimeoutSeconds = "handshake_timeout_seconds"
	ConfigKeyMaxNumPeers             = "max_num_peers"
	ConfigKeyInboundPeersRatio		 = "out_bound_ratio"
	ConfigKeyAuthEnc                 = "authenticated_encryption"

	// MConnection config keys
	ConfigKeySendRate = "send_rate"
	ConfigKeyRecvRate = "recv_rate"

	// Fuzz params
	ConfigFuzzEnable               = "fuzz_enable" // use the fuzz wrapped conn
	ConfigFuzzActive               = "fuzz_active" // toggle fuzzing
	ConfigFuzzMode                 = "fuzz_mode"   // eg. drop, delay
	ConfigFuzzMaxDelayMilliseconds = "fuzz_max_delay_milliseconds"
	ConfigFuzzProbDropRW           = "fuzz_prob_drop_rw"
	ConfigFuzzProbDropConn         = "fuzz_prob_drop_conn"
	ConfigFuzzProbSleep            = "fuzz_prob_sleep"
)

func setConfigDefaults(config *viper.Viper) {
	// Switch default config
	config.SetDefault(ConfigKeyDialTimeoutSeconds, 3)
	config.SetDefault(ConfigKeyHandshakeTimeoutSeconds, 20)
	config.SetDefault(ConfigKeyMaxNumPeers, 50)
	config.SetDefault(ConfigKeyInboundPeersRatio, 3)
	config.SetDefault(ConfigKeyAuthEnc, true)

	// MConnection default config
	config.SetDefault(ConfigKeySendRate, 5120000) // 5000KB/s
	config.SetDefault(ConfigKeyRecvRate, 5120000) // 5000KB/s

	// Fuzz defaults
	config.SetDefault(ConfigFuzzEnable, false)
	config.SetDefault(ConfigFuzzActive, false)
	config.SetDefault(ConfigFuzzMode, FuzzModeDrop)
	config.SetDefault(ConfigFuzzMaxDelayMilliseconds, 3000)
	config.SetDefault(ConfigFuzzProbDropRW, 0.2)
	config.SetDefault(ConfigFuzzProbDropConn, 0.00)
	config.SetDefault(ConfigFuzzProbSleep, 0.00)

	config.SetDefault("connection_reset_wait", 300)
}
