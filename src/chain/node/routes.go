package node

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	pbtypes "github.com/Baptist-Publication/angine/protos/types"
	agtypes "github.com/Baptist-Publication/angine/types"
	"github.com/Baptist-Publication/chorus-module/lib/go-crypto"
	rpc "github.com/Baptist-Publication/chorus-module/lib/go-rpc/server"
	"github.com/Baptist-Publication/chorus-module/xlib/def"
	"github.com/Baptist-Publication/chorus/src/chain/version"
)

const ChainIDArg = "chainid"

// RPCNode define the node's abilities provided for rpc calls
type RPCNode interface {
	GetOrg(string) (*OrgNode, error)
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
		"organizations":   rpc.NewRPCFunc(h.Orgs, ""),
		"status":          rpc.NewRPCFunc(h.Status, argsWithChainID("")),
		"net_info":        rpc.NewRPCFunc(h.NetInfo, argsWithChainID("")),
		"block":           rpc.NewRPCFunc(h.Block, argsWithChainID("height")),
		"validators":      rpc.NewRPCFunc(h.Validators, argsWithChainID("")),
		"za_surveillance": rpc.NewRPCFunc(h.ZaSurveillance, argsWithChainID("")),

		// broadcast API
		"broadcast_tx_commit": rpc.NewRPCFunc(h.BroadcastTxCommit, argsWithChainID("tx")),
		"broadcast_tx_sync":   rpc.NewRPCFunc(h.BroadcastTx, argsWithChainID("tx")),

		// query API
		"query":      rpc.NewRPCFunc(h.Query, argsWithChainID("query")),
		"event_code": rpc.NewRPCFunc(h.EventCode, argsWithChainID("code_hash")), // TODO now id is base-chain's name

		// specialOP API
		"request_special_op": rpc.NewRPCFunc(h.RequestSpecialOP, argsWithChainID("tx")),
	}
}

func (h *rpcHandler) Orgs() (agtypes.RPCResult, error) {
	app := h.node.MainOrg.Application.(*Metropolis)
	app.Lock()
	defer app.Unlock()
	names := make([]string, 0, len(app.Orgs))
	for n := range app.Orgs {
		names = append(names, string(n))
	}
	return &agtypes.ResultOrgs{Names: names}, nil
}

func (h *rpcHandler) Status(chainID string) (agtypes.RPCResult, error) {
	org, err := h.getOrg(chainID)
	if err != nil {
		return nil, ErrInvalidChainID
	}
	var (
		latestBlockMeta *pbtypes.BlockMeta
		latestBlockHash []byte
		latestAppHash   []byte
		latestBlockTime int64
	)
	latestHeight := org.Angine.Height()
	if latestHeight != 0 {
		_, latestBlockMeta, err = org.Angine.GetBlock(latestHeight)
		if err != nil {
			return nil, err
		}
		latestBlockHash = latestBlockMeta.Hash
		latestAppHash = latestBlockMeta.Header.AppHash
		latestBlockTime = latestBlockMeta.Header.Time
	}

	return &agtypes.ResultStatus{
		NodeInfo:          org.Angine.GetNodeInfo(),
		PubKey:            org.Angine.PrivValidator().GetPubKey(),
		LatestBlockHash:   latestBlockHash,
		LatestAppHash:     latestAppHash,
		LatestBlockHeight: latestHeight,
		LatestBlockTime:   latestBlockTime}, nil
}

func (h *rpcHandler) Block(chainID string, height def.INT) (agtypes.RPCResult, error) {
	org, err := h.getOrg(chainID)
	if err != nil {
		return nil, ErrInvalidChainID
	}
	res := agtypes.ResultBlock{}
	var blockc *agtypes.BlockCache
	blockc, res.BlockMeta, err = org.Angine.GetBlock(height)
	res.Block = blockc.Block
	return &res, err
}

func (h *rpcHandler) BroadcastTx(chainID string, tx []byte) (agtypes.RPCResult, error) {
	org, err := h.getOrg(chainID)
	if err != nil {
		return nil, ErrInvalidChainID
	}
	if err := org.Application.CheckTx(tx); err != nil {
		return nil, err
	}
	if err := org.Angine.BroadcastTx(tx); err != nil {
		return nil, err
	}
	return &agtypes.ResultBroadcastTx{Code: 0}, nil
}

func (h *rpcHandler) BroadcastTxCommit(chainID string, tx []byte) (agtypes.RPCResult, error) {
	org, err := h.getOrg(chainID)
	if err != nil {
		return nil, ErrInvalidChainID
	}
	if err := org.Application.CheckTx(tx); err != nil {
		return nil, err
	}
	if err := org.Angine.BroadcastTxCommit(tx); err != nil {
		return nil, err
	}

	return &agtypes.ResultBroadcastTxCommit{Code: 0}, nil
}

func (h *rpcHandler) Query(chainID string, query []byte) (agtypes.RPCResult, error) {
	org, err := h.getOrg(chainID)
	if err != nil {
		return nil, ErrInvalidChainID
	}
	return &agtypes.ResultQuery{Result: org.Application.Query(query)}, nil
}

func (h *rpcHandler) EventCode(chainID string, codeHash []byte) (agtypes.RPCResult, error) {
	if len(codeHash) == 0 {
		return nil, ErrMissingParams
	}
	app := h.node.MainOrg.Application.(*Metropolis)
	ret := app.EventCodeBase.Get(codeHash)
	return &agtypes.ResultQuery{
		Result: agtypes.NewResultOK(ret, ""),
	}, nil
}

func (h *rpcHandler) Validators(chainID string) (agtypes.RPCResult, error) {
	org, err := h.getOrg(chainID)
	if err != nil {
		return nil, ErrInvalidChainID
	}

	_, vs := org.Angine.GetValidators()
	return &agtypes.ResultValidators{
		Validators:  vs.Validators,
		BlockHeight: org.Angine.Height(),
	}, nil
}

func (h *rpcHandler) ZaSurveillance(chainID string) (agtypes.RPCResult, error) {
	org, err := h.getOrg(chainID)
	if err != nil {
		return nil, ErrInvalidChainID
	}
	bcHeight := org.Angine.Height()

	var totalNumTxs, txAvg int64
	if bcHeight >= 2 {
		startHeight := bcHeight - 200
		if startHeight < 1 {
			startHeight = 1
		}
		eBlock, _, err := org.Angine.GetBlock(bcHeight)
		if err != nil {
			return nil, err
		}
		endTime := agtypes.NanoToTime(eBlock.Header.Time)
		sBlock, _, err := org.Angine.GetBlock(startHeight)
		if err != nil {
			return nil, err
		}
		startTime := agtypes.NanoToTime(sBlock.Header.Time)
		totalNumTxs += int64(sBlock.Header.NumTxs)
		dura := endTime.Sub(startTime)
		for h := startHeight + 1; h < bcHeight; h++ {
			block, _, err := org.Angine.GetBlock(h)
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

	_, vals := org.Angine.GetValidators()

	res := agtypes.ResultSurveillance{
		Height:        bcHeight,
		NanoSecsPerTx: time.Duration(txAvg),
		Addr:          h.node.NodeInfo().RemoteAddr,
		IsValidator:   org.Angine.IsNodeValidator(&(h.node.NodeInfo().PubKey)),
		NumValidators: vals.Size(),
		NumPeers:      org.Angine.GetNumPeers(),
		RunningTime:   runningTime,
		PubKey:        h.node.NodeInfo().PubKey.KeyString(),
	}
	return &res, nil
}

func (h *rpcHandler) NetInfo(chainID string) (agtypes.RPCResult, error) {
	org, err := h.getOrg(chainID)
	if err != nil {
		return nil, ErrInvalidChainID
	}
	res := agtypes.ResultNetInfo{}
	res.Listening, res.Listeners, res.Peers = org.Angine.GetP2PNetInfo()
	return &res, nil
}

func (h *rpcHandler) RequestSpecialOP(chainID string, tx []byte) (agtypes.RPCResult, error) {
	org, err := h.getOrg(chainID)
	if err != nil {
		return nil, ErrInvalidChainID
	}

	if err := org.Angine.ProcessSpecialOP(tx); err != nil {
		res := &agtypes.ResultRequestSpecialOP{
			Code: pbtypes.CodeType_InternalError,
			Log:  err.Error(),
		}
		return res, err
	}

	return &agtypes.ResultRequestSpecialOP{
		Code: pbtypes.CodeType_OK,
	}, nil
}

func argsWithChainID(args string) string {
	if args == "" {
		return ChainIDArg
	}
	return ChainIDArg + "," + args
}

func (h *rpcHandler) getOrg(chainID string) (*OrgNode, error) {
	var org *OrgNode
	var err error
	if chainID == h.node.MainChainID {
		org = h.node.MainOrg
	} else {
		met := h.node.MainOrg.Application.(*Metropolis)
		org, err = met.GetOrg(chainID)
		if err != nil {
			return nil, ErrInvalidChainID
		}
	}

	return org, nil
}
