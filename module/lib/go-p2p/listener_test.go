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
	"bytes"
	"fmt"
	"os"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var logger *zap.Logger

type (
	infoWithDebug struct{}
)

func (l infoWithDebug) Enabled(lv zapcore.Level) bool {
	return lv == zapcore.InfoLevel || lv == zapcore.DebugLevel
}

func init() {
	encoderCfg := zap.NewProductionEncoderConfig()
	zcore := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderCfg),
		zapcore.NewMultiWriteSyncer(os.Stdout),
		infoWithDebug{},
	)

	logger = zap.New(zcore)
}

func TestListener(t *testing.T) {
	// Create a listener
	l, err := NewDefaultListener(logger, "tcp", ":8001", true)

	// Dial the listener
	lAddr := l.ExternalAddress()
	connOut, err := lAddr.Dial()
	if err != nil {
		t.Fatalf("Could not connect to listener address %v", lAddr)
	} else {
		t.Logf("Created a connection to listener address %v", lAddr)
	}
	connIn, ok := <-l.Connections()
	if !ok {
		t.Fatalf("Could not get inbound connection from listener")
	}

	msg := []byte("hi!")
	go connIn.Write(msg)
	b := make([]byte, 32)
	n, err := connOut.Read(b)
	if err != nil {
		t.Fatalf("Error reading off connection: %v", err)
	}

	b = b[:n]
	if !bytes.Equal(msg, b) {
		t.Fatalf("Got %s, expected %s", b, msg)
	}

	// Close the server, no longer needed.
	l.Stop()
}

func TestNewDefaultListener(t *testing.T) {
	l, err := NewDefaultListener(logger, "tcp", ":8001", true)
	panicErr(err)
	fmt.Println("For [:8001] :", l.ExternalAddress().String())

	l, err = NewDefaultListener(logger, "tcp", "0.0.0.0:9001", true)
	panicErr(err)
	fmt.Println("For [0.0.0.0:9001] :", l.ExternalAddress().String())

	l, err = NewDefaultListener(logger, "tcp", "192.168.27.116:7001", true)
	panicErr(err)
	fmt.Println("For [192.168.27.116:7001] :", l.ExternalAddress().String())

	l, err = NewDefaultListener(logger, "tcp", "172.16.7.162:5001", true)
	panicErr(err)
	fmt.Println("For [172.16.7.162:5001] :", l.ExternalAddress().String())
}

func panicErr(err error) {
	if err != nil {
		panic(err)
	}
}
