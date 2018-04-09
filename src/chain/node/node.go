package node

import (
	"fmt"
	"net"
	"net/http"
	_ "net/http/pprof"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/viper"
	"go.uber.org/zap"

	"github.com/Baptist-Publication/angine"
	ac "github.com/Baptist-Publication/angine/config"
	"github.com/Baptist-Publication/angine/types"
	cmn "github.com/Baptist-Publication/chorus-module/lib/go-common"
	"github.com/Baptist-Publication/chorus-module/lib/go-crypto"
	"github.com/Baptist-Publication/chorus-module/lib/go-p2p"
	"github.com/Baptist-Publication/chorus-module/lib/go-rpc/server"
	"github.com/Baptist-Publication/chorus-module/lib/go-wire"
	"github.com/Baptist-Publication/chorus/src/chain/app/evm"
	"github.com/Baptist-Publication/chorus/src/chain/version"
)

const (
	ReceiptsPrefix  = "receipts-"
	OfficialAddress = "0x7752b42608a0f1943c19fc5802cb027e60b4c911"
)

var Apps = make(map[string]AppMaker)

type Node struct {
	MainChainID string
	MainOrg     *OrgNode

	config        *viper.Viper
	privValidator *types.PrivValidator
	nodeInfo      *p2p.NodeInfo

	logger *zap.Logger
}

func AppExists(name string) (yes bool) {
	_, yes = Apps[name]
	return
}

func NewNode(logger *zap.Logger, conf *viper.Viper) *Node {
	aConf := ac.GetConfig(conf.GetString("runtime"))
	for k, v := range conf.AllSettings() {
		aConf.Set(k, v)
	}

	evmapp, _ := evm.NewEVMApp(logger, aConf)
	evmAngine := angine.NewAngine(logger, &angine.Tunes{Conf: aConf})
	tune := evmAngine.Tune
	if err := evmAngine.ConnectApp(evmapp); err != nil {
		cmn.PanicCrisis(err)
	}

	chainID := ""
	if evmAngine.Genesis() != nil {
		chainID = evmAngine.Genesis().ChainID
	}
	node := &Node{
		MainChainID: chainID,
		MainOrg: &OrgNode{
			Application: evmapp,
			Angine:      evmAngine,
			AngineTune:  tune,
			GenesisDoc:  evmAngine.Genesis(),
		},

		nodeInfo:      makeNodeInfo(aConf, evmAngine.PrivValidator().GetPubKey().(*crypto.PubKeyEd25519), evmAngine.P2PHost(), evmAngine.P2PPort()),
		config:        aConf,
		privValidator: evmAngine.PrivValidator(),
		logger:        logger,
	}

	evmAngine.RegisterNodeInfo(node.nodeInfo)

	// TODO reorg before runing online
	//register validator info to metro statedb
	//if conf.GetBool("enable_incentive") {
	//metropolis.RegisterValidators(metroAngine)
	//}

	return node
}

func RunNode(logger *zap.Logger, config *viper.Viper) {
	node := NewNode(logger, config)
	if err := node.Start(); err != nil {
		cmn.Exit(cmn.Fmt("Failed to start node: %v", err))
	}
	if node.GetConf().GetString("rpc_laddr") != "" {
		if _, err := node.StartRPC(); err != nil {
			cmn.PanicCrisis(err)
		}
	}
	if config.GetBool("pprof") {
		go func() {
			http.ListenAndServe(":6060", nil)
		}()
	}

	fmt.Printf("node (%s) is running on %s:%d ......\n", node.MainChainID, node.NodeInfo().ListenHost(), node.NodeInfo().ListenPort())

	cmn.TrapSignal(func() {
		node.Stop()
	})
}

// Call Start() after adding the listeners.
func (n *Node) Start() error {
	if err := n.MainOrg.Start(); err != nil {
		return fmt.Errorf("fail to start, error: %v", err)
	}

	n.MainOrg.GenesisDoc = n.MainOrg.Angine.Genesis()
	n.MainChainID = n.MainOrg.GenesisDoc.ChainID

	return nil
}

func (n *Node) Stop() {
	n.logger.Info("Stopping Node")
	n.MainOrg.Stop()
}

func makeNodeInfo(config *viper.Viper, pubkey *crypto.PubKeyEd25519, p2pHost string, p2pPort uint16) *p2p.NodeInfo {
	nodeInfo := &p2p.NodeInfo{
		PubKey:      *pubkey,
		Moniker:     config.GetString("moniker"),
		Network:     config.GetString("chain_id"),
		SigndPubKey: config.GetString("signbyCA"),
		Version:     version.GetVersion(),
		Other: []string{
			cmn.Fmt("wire_version=%v", wire.Version),
			cmn.Fmt("p2p_version=%v", p2p.Version),
			// Fmt("consensus_version=%v", n.StateMachine.Version()),
			// Fmt("rpc_version=%v/%v", rpc.Version, rpccore.Version),
			cmn.Fmt("node_start_at=%s", strconv.FormatInt(time.Now().Unix(), 10)),
			cmn.Fmt("revision=%s", version.GetCommitVersion()),
		},
		RemoteAddr: config.GetString("rpc_laddr"),
		ListenAddr: cmn.Fmt("%v:%v", p2pHost, p2pPort),
	}

	// We assume that the rpcListener has the same ExternalAddress.
	// This is probably true because both P2P and RPC listeners use UPnP,
	// except of course if the rpc is only bound to localhost

	return nodeInfo
}

func (n *Node) NodeInfo() *p2p.NodeInfo {
	return n.nodeInfo
}

func (n *Node) StartRPC() ([]net.Listener, error) {
	listenAddrs := strings.Split(n.config.GetString("rpc_laddr"), ",")
	listeners := make([]net.Listener, len(listenAddrs))

	for i, listenAddr := range listenAddrs {
		mux := http.NewServeMux()
		// wm := rpcserver.NewWebsocketManager(rpcRoutes, n.evsw)
		// mux.HandleFunc("/websocket", wm.WebsocketHandler)
		rpcserver.RegisterRPCFuncs(n.logger, mux, n.rpcRoutes())
		listener, err := rpcserver.StartHTTPServer(n.logger, listenAddr, mux)
		if err != nil {
			return nil, err
		}
		listeners[i] = listener
	}

	return listeners, nil
}

func (n *Node) PrivValidator() *types.PrivValidator {
	return n.privValidator
}

func (n *Node) GetConf() *viper.Viper {
	return n.config
}
