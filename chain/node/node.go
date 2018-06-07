package node

import (
	"fmt"
	"net"
	"net/http"
	_ "net/http/pprof"
	"strconv"
	"strings"
	"time"

	"github.com/Baptist-Publication/chorus/angine"
	agtypes "github.com/Baptist-Publication/chorus/angine/types"
	"github.com/Baptist-Publication/chorus/chain/app"
	"github.com/Baptist-Publication/chorus/chain/version"
	"github.com/Baptist-Publication/chorus/config"
	cmn "github.com/Baptist-Publication/chorus/module/lib/go-common"
	"github.com/Baptist-Publication/chorus/module/lib/go-crypto"
	"github.com/Baptist-Publication/chorus/module/lib/go-p2p"
	"github.com/Baptist-Publication/chorus/module/lib/go-rpc/server"
	"github.com/Baptist-Publication/chorus/module/lib/go-wire"
	"github.com/Baptist-Publication/chorus/test/testdb"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

const (
	ReceiptsPrefix = "receipts-"
)

var Apps = make(map[string]AppMaker)

type Node struct {
	config        *viper.Viper
	privValidator *agtypes.PrivValidator
	nodeInfo      *p2p.NodeInfo

	logger *zap.Logger

	running     int64
	Superior    Superior
	Angine      *angine.Angine
	Application agtypes.Application
	GenesisDoc  *agtypes.GenesisDoc
}

func AppExists(name string) (yes bool) {
	_, yes = Apps[name]
	return
}

func NewNode(logger *zap.Logger, c *viper.Viper) *Node {
	conf := config.GetConfig(c.GetString("runtime"))
	for k, v := range c.AllSettings() {
		conf.Set(k, v)
	}

	App, _ := app.NewApp(logger, conf)
	evmAngine := angine.NewAngine(logger, conf)
	if evmAngine == nil {
		cmn.Exit("fail to start, please check your angine config")
	}
	if err := evmAngine.ConnectApp(App); err != nil {
		cmn.PanicCrisis(err)
	}
	App.AngineRef = evmAngine

	node := &Node{
		Application: App,
		Angine:      evmAngine,
		GenesisDoc:  evmAngine.Genesis(),

		nodeInfo:      makeNodeInfo(conf, evmAngine.PrivValidator().GetPubKey().(*crypto.PubKeyEd25519), evmAngine.P2PHost(), evmAngine.P2PPort()),
		config:        conf,
		privValidator: evmAngine.PrivValidator(),
		logger:        logger,
	}

	evmAngine.RegisterNodeInfo(node.nodeInfo)

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

	fmt.Printf("node is running on %s:%d ......\n", node.NodeInfo().ListenHost(), node.NodeInfo().ListenPort())

	cmn.TrapSignal(func() {
		node.Stop()
	})
}

// Call Start() after adding the listeners.
func (n *Node) Start() error {
	err := testdb.InitDB()
	if err != nil {
		n.Angine.Stop()
		return err
	}
	if err := n.Application.Start(); err != nil {
		n.Angine.Stop()
		return err
	}
	if err := n.Angine.Start(); err != nil {
		n.Application.Stop()
		n.Angine.Stop()
		return err
	}
	for n.Angine.Genesis() == nil {
		time.Sleep(500 * time.Millisecond)
	}

	n.GenesisDoc = n.Angine.Genesis()

	return nil
}

func (n *Node) Stop() {
	n.logger.Info("Stopping Node")

	n.Angine.Stop()
	n.Application.Stop()
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

func (n *Node) PrivValidator() *agtypes.PrivValidator {
	return n.privValidator
}

func (n *Node) GetConf() *viper.Viper {
	return n.config
}
