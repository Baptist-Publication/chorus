package node

import (
	"errors"
	"strconv"
	"strings"
	"time"

	pbtypes "github.com/Baptist-Publication/chorus/angine/protos/types"
	agtypes "github.com/Baptist-Publication/chorus/angine/types"
	"github.com/Baptist-Publication/chorus/module/lib/go-crypto"
	rpc "github.com/Baptist-Publication/chorus/module/lib/go-rpc/server"
	"github.com/Baptist-Publication/chorus/module/xlib/def"
)

const ChainIDArg = "chainid"

// RPCNode define the node's abilities provided for rpc calls
type RPCNode interface {
	Height() def.INT
	GetBlock(height def.INT) (*agtypes.BlockCache, *pbtypes.BlockMeta)
	BroadcastTx(tx []byte) error
	BroadcastTxCommit(tx []byte) error
	FlushMempool()
	GetValidators() (def.INT, []*agtypes.Validator)
	GetP2PNetInfo() (bool, []string, []*agtypes.Peer)
	GetNumPeers() int
	GetConsensusStateInfo() (string, []string)
	GetNumUnconfirmedTxs() int
	GetUnconfirmedTxs() []agtypes.Tx
	IsNodeValidator(pub crypto.PubKey) bool
	GetBlacklist() []string
}

type rpcHandler struct {
	node *Node
}

var (
	ErrInvalidChainID = errors.New("no such chain id")
	ErrMissingParams  = errors.New("missing params")
)

func newRPCHandler(n *Node) *rpcHandler {
	return &rpcHandler{node: n}
}

func (n *Node) rpcRoutes() map[string]*rpc.RPCFunc {
	h := newRPCHandler(n)
	return map[string]*rpc.RPCFunc{
		// info API
		// "organizations":   rpc.NewRPCFunc(h.Orgs, ""),
		"status":          rpc.NewRPCFunc(h.Status, ""),
		"net_info":        rpc.NewRPCFunc(h.NetInfo, ""),
		"block":           rpc.NewRPCFunc(h.Block, "height"),
		"validators":      rpc.NewRPCFunc(h.Validators, ""),
		"is_validator":    rpc.NewRPCFunc(h.Is_Validator, "pubkey"),
		"za_surveillance": rpc.NewRPCFunc(h.ZaSurveillance, ""),

		// broadcast API
		"broadcast_tx_commit": rpc.NewRPCFunc(h.BroadcastTxCommit, "tx"),
		"broadcast_tx_sync":   rpc.NewRPCFunc(h.BroadcastTx, "tx"),

		// query API
		"query": rpc.NewRPCFunc(h.Query, "query"),
	}
}

func (h *rpcHandler) Status(chainID string) (agtypes.RPCResult, error) {
	var (
		err             error
		latestBlockMeta *pbtypes.BlockMeta
		latestBlockHash []byte
		latestAppHash   []byte
		latestBlockTime int64
	)
	latestHeight := h.node.Angine.Height()
	if latestHeight != 0 {
		_, latestBlockMeta, err = h.node.Angine.GetBlock(latestHeight)
		if err != nil {
			return nil, err
		}
		latestBlockHash = latestBlockMeta.Hash
		latestAppHash = latestBlockMeta.Header.AppHash
		latestBlockTime = latestBlockMeta.Header.Time
	}

	return &agtypes.ResultStatus{
		NodeInfo:          h.node.Angine.GetNodeInfo(),
		PubKey:            h.node.Angine.PrivValidator().GetPubKey(),
		LatestBlockHash:   latestBlockHash,
		LatestAppHash:     latestAppHash,
		LatestBlockHeight: latestHeight,
		LatestBlockTime:   latestBlockTime}, nil
}

func (h *rpcHandler) Block(chainID string, height def.INT) (agtypes.RPCResult, error) {
	var err error
	res := agtypes.ResultBlock{}
	var blockc *agtypes.BlockCache
	blockc, res.BlockMeta, err = h.node.Angine.GetBlock(height)
	res.Block = blockc.Block
	return &res, err
}

func (h *rpcHandler) BroadcastTx(chainID string, tx []byte) (agtypes.RPCResult, error) {
	if err := h.node.Application.CheckTx(tx); err != nil {
		return nil, err
	}
	if err := h.node.Angine.BroadcastTx(tx); err != nil {
		return nil, err
	}
	return &agtypes.ResultBroadcastTx{Code: 0}, nil
}

func (h *rpcHandler) BroadcastTxCommit(chainID string, tx []byte) (agtypes.RPCResult, error) {
	if err := h.node.Application.CheckTx(tx); err != nil {
		return nil, err
	}
	if err := h.node.Angine.BroadcastTxCommit(tx); err != nil {
		return nil, err
	}

	return &agtypes.ResultBroadcastTxCommit{Code: 0}, nil
}

func (h *rpcHandler) Query(chainID string, query []byte) (agtypes.RPCResult, error) {
	return &agtypes.ResultQuery{Result: h.node.Application.Query(query)}, nil
}

func (h *rpcHandler) Validators(chainID string) (agtypes.RPCResult, error) {

	_, vs := h.node.Angine.GetValidators()
	return &agtypes.ResultValidators{
		Validators:  vs.Validators,
		BlockHeight: h.node.Angine.Height(),
	}, nil
}

func (h *rpcHandler) Is_Validator(chainID, pubkey string) (agtypes.RPCResult, error) {

	_, vs := h.node.Angine.GetValidators()
	for _, val := range vs.Validators {
		if pubkey == val.PubKey.KeyString() {
			return &agtypes.ResultQuery{
				Result: agtypes.NewResultOK([]byte{1}, ""),
			}, nil
		}
	}

	return &agtypes.ResultQuery{
		Result: agtypes.NewResultOK([]byte{0}, "account not is validator"),
	}, nil
}

func (h *rpcHandler) ZaSurveillance(chainID string) (agtypes.RPCResult, error) {
	bcHeight := h.node.Angine.Height()

	var totalNumTxs, txAvg int64
	if bcHeight >= 2 {
		startHeight := bcHeight - 200
		if startHeight < 1 {
			startHeight = 1
		}
		eBlock, _, err := h.node.Angine.GetBlock(bcHeight)
		if err != nil {
			return nil, err
		}
		endTime := agtypes.NanoToTime(eBlock.Header.Time)
		sBlock, _, err := h.node.Angine.GetBlock(startHeight)
		if err != nil {
			return nil, err
		}
		startTime := agtypes.NanoToTime(sBlock.Header.Time)
		totalNumTxs += int64(sBlock.Header.NumTxs)
		dura := endTime.Sub(startTime)
		for i := startHeight + 1; i < bcHeight; i++ {
			block, _, err := h.node.Angine.GetBlock(i)
			if err != nil {
				return nil, err
			}
			totalNumTxs += int64(block.Header.NumTxs)
		}
		if totalNumTxs > 0 {
			txAvg = int64(dura) / totalNumTxs
		}
	}

	var runningTime time.Duration
	for _, oth := range h.node.NodeInfo().Other {
		if strings.HasPrefix(oth, "node_start_at") {
			ts, err := strconv.ParseInt(string(oth[14:]), 10, 64)
			if err != nil {
				return -1, err
			}
			runningTime = time.Duration(time.Now().Unix() - ts)
		}
	}

	_, vals := h.node.Angine.GetValidators()

	res := agtypes.ResultSurveillance{
		Height:        bcHeight,
		NanoSecsPerTx: time.Duration(txAvg),
		Addr:          h.node.NodeInfo().RemoteAddr,
		IsValidator:   h.node.Angine.IsNodeValidator(&(h.node.NodeInfo().PubKey)),
		NumValidators: vals.Size(),
		NumPeers:      h.node.Angine.GetNumPeers(),
		RunningTime:   runningTime,
		PubKey:        h.node.NodeInfo().PubKey.KeyString(),
	}
	return &res, nil
}

func (h *rpcHandler) NetInfo(chainID string) (agtypes.RPCResult, error) {
	res := agtypes.ResultNetInfo{}
	res.Listening, res.Listeners, res.Peers = h.node.Angine.GetP2PNetInfo()
	return &res, nil
}