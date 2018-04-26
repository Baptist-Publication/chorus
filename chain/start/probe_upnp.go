package start

import (
	"encoding/json"
	"fmt"

	"go.uber.org/zap"

	"github.com/Baptist-Publication/chorus/module/lib/go-p2p/upnp"
)

func Probe_upnp(logger *zap.Logger) {

	capabilities, err := upnp.Probe(logger)
	if err != nil {
		fmt.Printf("Probe failed: %v\n", err)
	} else {
		fmt.Println("Probe success!")
		jsonBytes, err := json.Marshal(capabilities)
		if err != nil {
			panic(err)
		}
		fmt.Println(string(jsonBytes))
	}

}
