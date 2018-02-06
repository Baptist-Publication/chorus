package node

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"math/big"
	"math/rand"
	"time"

	agtypes "github.com/Baptist-Publication/angine/types"
	"github.com/Baptist-Publication/chorus-module/lib/go-crypto"
	"github.com/Baptist-Publication/chorus-module/xlib/def"
	"github.com/Baptist-Publication/chorus/src/tools"
	cvtools "github.com/Baptist-Publication/chorus/src/tools"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

// ExecuteTx execute tx one by one in the loop, without lock, so should always be called between Lock() and Unlock() on the *stateDup
func (met *Metropolis) ExecuteTx(block *agtypes.BlockCache, bs []byte, txIndex int) (validtx []byte, err error) {
	if IsOrgRelatedTx(bs) {
		met.state.OrgHeights = append(met.state.OrgHeights, block.Header.Height)
	}

	txBytes := agtypes.UnwrapTx(bs)

	switch {
	case IsOrgCancelTx(bs):
		tx := &OrgCancelTx{}
		if err := cvtools.TxFromBytes(txBytes, tx); err != nil {
			return nil, errors.Wrap(err, "exec orgcancel failed")
		}
		if ok, err := cvtools.TxVerifySignature(tx); !ok {
			return nil, errors.Wrap(err, "org cancel fail to verify signature")
		}
		if err := met.checkOrgCancel(tx); err != nil {
			return nil, err
		}
		if err := met.executeOrgCancelTx(tx); err != nil {
			return nil, errors.Wrap(err, "exec orgcancel failed")
		}
	case IsOrgConfirmTx(bs):
		tx := &OrgConfirmTx{}
		if err = cvtools.TxFromBytes(txBytes, tx); err != nil {
			return nil, errors.Wrap(err, "exec orgconfirm failed")
		}
		if ok, err := cvtools.TxVerifySignature(tx); !ok {
			return nil, errors.Wrap(err, "org confirm fail to verify signature")
		}
		if err = met.checkOrgConfirm(tx); err != nil {
			return nil, err
		}
		if err := met.executeOrgConfirmTx(tx); err != nil {
			return nil, errors.Wrap(err, "exec orgconfirm failed")
		}
	case IsOrgTx(bs):
		tx := &OrgTx{}
		if err = cvtools.TxFromBytes(txBytes, tx); err != nil {
			return nil, errors.Wrap(err, "exec org failed")
		}
		if ok, err := cvtools.TxVerifySignature(tx); !ok {
			return nil, errors.Wrap(err, "org fail to verify signature")
		}
		if err := met.executeOrgTx(tx); err != nil {
			return nil, errors.Wrap(err, "exec org failed")
		}
	case IsEventRequestTx(bs):
		tx := &EventRequestTx{}
		if err := cvtools.TxFromBytes(txBytes, tx); err != nil {
			return nil, errors.Wrap(err, "exec eventrequest failed")
		}
		if ok, err := cvtools.TxVerifySignature(tx); !ok {
			return nil, errors.Wrap(err, "event request fail to verify signature")
		}
		if err := met.checkEventRequestTx(tx); err != nil {
			return nil, err
		}
		if err := met.executeEventRequestTx(tx, block.Header.Height, block.Header.Maker); err != nil {
			return nil, errors.Wrap(err, "exec eventrequest failed")
		}
	case IsEventSubscribeTx(bs):
		tx := &EventSubscribeTx{}
		if err := cvtools.TxFromBytes(txBytes, tx); err != nil {
			return nil, errors.Wrap(err, "exec eventsubscribe failed")
		}
		if ok, err := cvtools.TxVerifySignature(tx); !ok {
			return nil, errors.Wrap(err, "event subscribe fail to verify signature")
		}
		if err := met.checkEventSubscribeTx(tx); err != nil {
			return nil, errors.Wrap(err, "checkExecuteEventSubscribeTx failed in exec")
		}
		if err := met.executeEventSubscribeTx(tx, block); err != nil {
			return nil, errors.Wrap(err, "exec eventsubscribe failed")
		}
	case IsEventNotificationTx(bs):
		tx := &EventNotificationTx{}
		if err := cvtools.TxFromBytes(txBytes, tx); err != nil {
			return nil, errors.Wrap(err, "fail to load eventnotification")
		}
		if ok, err := cvtools.TxVerifySignature(tx); !ok {
			return nil, errors.Wrap(err, "event notification fail to verify signature")
		}
		if err := met.executeEventNotification(tx, block); err != nil {
			return nil, errors.Wrap(err, "exec notification failed")
		}
	case IsEventUnsubscribeTx(bs):
		tx := &EventUnsubscribeTx{}
		if err := cvtools.TxFromBytes(txBytes, tx); err != nil {
			return nil, errors.Wrap(err, "fail to load eventunsubscribe")
		}
		if ok, err := cvtools.TxVerifySignature(tx); !ok {
			return nil, errors.Wrap(err, "event unsubscribe fail to verify signature")
		}
		if err := met.executeEventUnsubscribeTx(tx, block); err != nil {
			return nil, errors.Wrap(err, "exec eventunsubscribe failed")
		}
	case IsEventUploadCodeTx(bs):
		tx := &EventUploadCodeTx{}
		if err := cvtools.TxFromBytes(txBytes, tx); err != nil {
			return nil, errors.Wrap(err, "fail to load eventuploadcode")
		}
		if ok, err := cvtools.TxVerifySignature(tx); !ok {
			return nil, errors.Wrap(err, "event upload code fail to verify signautre")
		}
		if err := met.checkEventUploadCodeTx(tx); err != nil {
			return nil, err
		}
		if err := met.executeEventUploadCodeTx(tx, block); err != nil {
			return nil, errors.Wrap(err, "exec eventuploadcode failed")
		}
	case IsCoSiInitTx(bs):
		tx := &CoSiInitTx{}
		if err := cvtools.TxFromBytes(txBytes, tx); err != nil {
			return nil, errors.Wrap(err, "exec cosi init failed")
		}
		if ok, err := cvtools.TxVerifySignature(tx); !ok {
			return nil, errors.Wrap(err, "cosi init fail to verify signature")
		}
		if err := met.checkCoSiInitTx(tx); err != nil {
			return nil, err
		}
		if err := met.executeCoSiInitTx(tx, block.Header.Height); err != nil {
			return nil, errors.Wrap(err, "exec cosi init failed")
		}
	case IsEcoTx(bs):
		if err := met.executeEcoTx(bs, block.Header.Height); err != nil {
			return nil, err
		}
	default:
		return nil, ErrUnknownTx
	}

	return bs, nil
}

// OnExecute would not care about Block.ExTxs
func (met *Metropolis) OnExecute(height, round def.INT, block *agtypes.BlockCache) (interface{}, error) {
	var (
		res agtypes.ExecuteResult
		err error
	)

	// reset fee accum
	met.FeeAccum = new(big.Int).SetUint64(0)

	if block.Header.Height == 1 { //register beginning validator info
		met.RegisterValidators(block.VSetCache())
	}

	bdata := block.Data
	for i := range bdata.Txs {
		vtx, err := met.ExecuteTx(block, bdata.Txs[i], i)
		if err == nil {
			res.ValidTxs = append(res.ValidTxs, vtx)
		} else {
			if err == ErrUnknownTx {
				met.logger.Info(err.Error(), zap.Binary("tx", bdata.Txs[i]))
			} else {
				met.logger.Error("[metro_onExec],exec error", zap.Error(err))
			}
			res.InvalidTxs = append(res.InvalidTxs, agtypes.ExecuteInvalidTx{Bytes: bdata.Txs[i], Error: err})
		}
	}

	if err = met.doCoinbaseTx(block); err != nil {
		res.Error = err
		return res, err
	}

	return res, err
}

func (met *Metropolis) doCoinbaseTx(block *agtypes.BlockCache) error {
	var pubkey crypto.PubKeyEd25519
	copy(pubkey[:], block.Header.CoinBase)

	rewards := calculateRewards(uint64(block.Header.Height))

	return met.accState.AddBalance(&pubkey, new(big.Int).Add(rewards, met.FeeAccum))
}

func calculateRewards(height uint64) *big.Int {
	startRewards := uint64(128 * 1000000000)
	declinePerBlocks := uint64(5000000)

	declineCount := height / declinePerBlocks
	if declineCount > 7 {
		declineCount = 7
	}
	declineBase := uint64(math.Pow(2, float64(declineCount)))
	rewards := startRewards / declineBase

	return new(big.Int).SetUint64(rewards)
}

func (met *Metropolis) executeOrgCancelTx(tx *OrgCancelTx) error {
	delete(met.PendingOrgTxs, string(tx.TxHash))
	return nil
}

// executeOrgConfirmTx defines those actions every node should do after verifying the confirm tx and the tx confirmed by it.
// so now it's basicly modifying the orgState
func (met *Metropolis) executeOrgConfirmTx(tx *OrgConfirmTx) error {
	var (
		accnt  *OrgAccount
		err    error
		pubkey crypto.PubKeyEd25519
	)

	orgTx, ok := met.PendingOrgTxs[string(tx.TxHash)]
	if !ok {
		return errors.Errorf("no such pending org tx: %X", tx.TxHash)
	}
	delete(met.PendingOrgTxs, string(tx.TxHash))

	copy(pubkey[:], tx.PubKey)
	attr := tx.Attributes
	switch orgTx.Act {
	case OrgCreate:
		if accnt, err = met.OrgState.CreateAccount(tx.ChainID, 0); err != nil {
			return err
		}
		if err = accnt.AddNode(&pubkey, attr); err != nil {
			return err
		}
		return nil
	case OrgJoin:
		if accnt, err = met.OrgState.GetAccount(tx.ChainID); err != nil {
			return err
		}
		if err = accnt.AddNode(&pubkey, attr); err != nil {
			return err
		}
		return nil
	case OrgLeave:
		if accnt, err = met.OrgState.GetAccount(tx.ChainID); err != nil {
			return err
		}
		if err = accnt.RemoveNode(&pubkey); err != nil {
			return err
		}
		return nil
	}

	return fmt.Errorf("unknown act: %v", orgTx)
}

func (met *Metropolis) executeOrgTx(tx *OrgTx) (err error) {
	var (
		node          *OrgNode
		cancelOnError bool

		txHash, _ = cvtools.TxHash(tx)
		pubkey, _ = met.core.GetPublicKey()
	)

	defer func() {
		if cancelOnError && err != nil {
			met.sendOrgCancel(tx, &pubkey, txHash)
		}
	}()

	met.PendingOrgTxs[string(txHash)] = tx

	if !bytes.Equal(tx.PubKey, pubkey[:]) {
		// short cut, and wait for confirmation
		return nil
	}

	cancelOnError = true

	if err = met.checkOrgs(tx); err != nil {
		return err
	}
	if err = met.checkOrgState(tx); err != nil {
		return err
	}

	switch tx.Act {
	case OrgCreate:
		if node, err = met.createOrgNode(tx); err != nil {
			return err
		}
		if node == nil {
			return errors.Wrap(errors.Errorf("nil org node is returned"), "[executeOrgTx OrgCreate]")
		}
		met.SetOrg(tx.ChainID, tx.App, node)
	case OrgJoin:
		if node, err = met.createOrgNode(tx); err != nil {
			return err
		}
		if node == nil {
			return errors.Wrap(errors.Errorf("nil org node is returned"), "[executeOrgTx OrgJoin]")
		}
		met.SetOrg(tx.ChainID, tx.App, node)
	case OrgLeave:
		if err = met.RemoveOrg(tx.ChainID); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unimplemented")
	}

	met.sendOrgConfirm(tx, &pubkey, txHash, node)
	return nil
}

func (met *Metropolis) sendOrgConfirm(tx *OrgTx, pubkey crypto.PubKey, txHash []byte, node *OrgNode) {
	// pubkeyEd := pubkey.(crypto.PubKeyEd25519)
	attrs := make(map[string]string)
	if node != nil {
		metroNode := node.Application.(IMetropolisApp)
		nodeAttrs := metroNode.GetAttributes()
		for k, v := range nodeAttrs {
			attrs[k] = v
		}
	}
	if met.config.IsSet("event_external_address") {
		attrs["event_external_address"] = met.config.GetString("event_external_address")
	}
	confirmTx := &OrgConfirmTx{
		ChainID:    tx.ChainID,
		Act:        tx.Act,
		Time:       time.Now(),
		TxHash:     txHash,
		Attributes: attrs,
	}

	if _, err := cvtools.TxSign(confirmTx, met.core.GetEngine().PrivValidator().GetPrivateKey()); err != nil {
		met.logger.Error("[sendOrgConfirm]", zap.Error(err))
		return
	}
	load, _ := cvtools.TxToBytes(confirmTx)
	if err := met.core.GetEngine().BroadcastTx(agtypes.WrapTx(OrgConfirmTag, load)); err != nil {
		met.logger.Error("fail to broadcast org confirm", zap.Error(err))
	}
}

func (met *Metropolis) sendOrgCancel(tx *OrgTx, pubkey crypto.PubKey, txHash []byte) {
	// pubkeyEd := pubkey.(crypto.PubKeyEd25519)
	cancelTx := &OrgCancelTx{
		TxHash: txHash,
		Time:   time.Now(),
	}

	if _, err := cvtools.TxSign(cancelTx, met.core.GetEngine().PrivValidator().GetPrivateKey()); err != nil {
		met.logger.Error("[sendOrgCancel]", zap.Error(err))
		return
	}
	load, _ := cvtools.TxToBytes(cancelTx)
	if err := met.core.GetEngine().BroadcastTx(agtypes.WrapTx(OrgCancelTag, load)); err != nil {
		met.logger.Error("fail to broadcast org cancel", zap.Error(err))
	}
}

// executeEventRequestTx uses CoSiInitTx to deliver the decision call to the subchain's members
func (met *Metropolis) executeEventRequestTx(tx *EventRequestTx, height def.INT, blockMaker []byte) error {
	pubkey, _ := met.core.GetPublicKey()
	privkey, _ := met.core.GetPrivateKey()

	txHash, err := cvtools.TxHash(tx)
	if err != nil {
		return err
	}

	// pending txs keep the relation cross txs, it is in the node's memory, no one else can mess with it
	met.PendingEventRequestTxs[string(txHash)] = tx

	// by default, we set up the rule that
	// only the current block maker is chosen to delegate the group in the following actions
	// and then, he will choose, randomly from the target subchain's nodes, one to be the leader in the CoSi procedure
	if bytes.Equal(blockMaker, pubkey[:]) {
		sourceAccnt, err := met.OrgState.GetAccount(tx.Source)
		if err != nil {
			return err
		}
		sourceNodes := sourceAccnt.GetNodes()
		idx := rand.Intn(len(sourceNodes))
		i := 0
		selected := ""
		for pk := range sourceNodes {
			if i == idx {
				selected = pk
				break
			}
			i++
		}
		selectedPK, err := hex.DecodeString(selected)
		if err != nil {
			return err
		}

		ctx := &CoSiInitTx{
			Type:     "eventrequest",
			ChainID:  tx.Source,
			Height:   height,
			Receiver: selectedPK,
			TxHash:   txHash,
			Data:     tx.SourceHash,
		}

		if _, err := cvtools.TxSign(ctx, &privkey); err != nil {
			return errors.Wrap(err, "[executeEventRequestTx]")
		}
		txBytes, _ := cvtools.TxToBytes(ctx)

		// engine.BroadcastTx fails under rare conditions
		if err := met.core.GetEngine().BroadcastTx(agtypes.WrapTx(CoSiInitTag, txBytes)); err != nil {
			return err
		}
	}

	return nil
}

func (met *Metropolis) executeCoSiInitTx(tx *CoSiInitTx, height def.INT) error {
	orgNode, err := met.GetOrg(tx.ChainID)
	if err != nil {
		return err
	}

	pubkey, _ := met.core.GetPublicKey()
	privkey, _ := met.core.GetPrivateKey()

	if bytes.Equal(tx.Receiver, pubkey[:]) {
		etx, ok := met.PendingEventRequestTxs[string(tx.TxHash)]
		if !ok {
			return fmt.Errorf("so such event request: %X", tx.TxHash)
		}
		accnt, err := met.OrgState.GetAccount(tx.ChainID)
		if err != nil {
			return errors.Wrap(err, "cosiininttx fail to get node's own orgState")
		}
		orgnode, err := met.GetOrg(tx.ChainID)
		if err != nil {
			return errors.Wrap(err, "cosiininttx fail to get node's own orgNode")
		}
		if _, ok := orgnode.Application.(EventApp); !ok {
			return errors.Errorf("%s can't support event", orgnode.GetAppName())
		}

		_, orgVals := orgnode.GetEngine().GetValidators()
		_, v := orgVals.GetByIndex(rand.Intn(orgVals.Size()))
		valPubKey := v.GetPubKey().(*crypto.PubKeyEd25519)
		attr, ok := accnt.GetNodes()[valPubKey.KeyString()]
		if !ok {
			return errors.Errorf("orgstate incosistent: val(%s) is not in the orgstate account", valPubKey.KeyString())
		}
		cosiAddr := GetCoSiAddress(attr)
		if cosiAddr == "" {
			return fmt.Errorf("CoSi requires an extra listening port")
		}

		// leader who must be a validator gives out his cosiAddr to all the followers
		cositx := &CoSiTx{
			Type:       tx.Type,
			Leader:     valPubKey[:],
			LeaderAddr: cosiAddr,
			TxHash:     tx.TxHash,
			Height:     height,
			Data:       append(tx.Data, []byte("|<>|"+etx.Listener)...),
		}

		if _, err := cvtools.TxSign(cositx, &privkey); err != nil {
			return errors.Wrap(err, "[executeCoSiInitTx]")
		}
		txBytes, _ := cvtools.TxToBytes(cositx)
		// CoSiTx gets delivered into the target subchain to perform any CoSi action within only that subchain
		if err := orgNode.GetEngine().BroadcastTx(agtypes.WrapTx(CoSiTag, txBytes)); err != nil {
			return err
		}
	}

	return nil
}

func (met *Metropolis) executeEventSubscribeTx(tx *EventSubscribeTx, block *agtypes.BlockCache) error {
	rtx, ok := met.PendingEventRequestTxs[string(tx.TxHash)]
	if !ok {
		met.logger.Debug("there is no such pending request")
		return fmt.Errorf("there is no such pending request: %v", tx.TxHash)
	}

	var err error
	var listener, source *EventAccount

	if met.EventState.ExistAccount(rtx.Listener) {
		listener, err = met.EventState.GetAccount(rtx.Listener)
	} else {
		listener, err = met.EventState.CreateAccount(rtx.Listener)
	}
	if err != nil {
		met.logger.Error("event state account error", zap.Error(err))
		return err
	}

	if met.EventState.ExistAccount(rtx.Source) {
		source, err = met.EventState.GetAccount(rtx.Source)
	} else {
		source, err = met.EventState.CreateAccount(rtx.Source)
	}
	if err != nil {
		met.logger.Error("event state account error", zap.Error(err))
		return err
	}

	listener.AddPublisher(rtx.Source, rtx.ListenerHash)
	source.AddSubscriber(rtx.Listener, rtx.SourceHash)

	fmt.Printf("%s is listening on %s\n", rtx.Listener, rtx.Source)
	met.logger.Info(fmt.Sprintf("%s is listening on %s", rtx.Listener, rtx.Source))

	return nil
}

func (met *Metropolis) executeEventNotification(tx *EventNotificationTx, block *agtypes.BlockCache) error {
	subOrg, err := met.GetOrg(tx.Listener)
	if err != nil {
		met.logger.Debug("this event notification is not for this node", zap.String("listener", tx.Listener))
		return nil
	}

	// perform a lottery to win the chance to fetch event data
	// the lottery goes as:
	// Min(Hash(pubkey + eventID)) wins
	eventID := fmt.Sprintf("%s,%s,%d", tx.Listener, tx.Source, tx.Height)
	pked, _ := met.core.GetPublicKey()
	nodePK := pked.KeyString()
	lottery := new(big.Int)
	competitor := new(big.Int)
	lotteryByte32 := sha256.Sum256(append([]byte(nodePK), []byte(eventID)...))
	lottery.SetBytes(lotteryByte32[:])

	subAccount, err := met.OrgState.GetAccount(tx.Listener)
	if err != nil {
		return err
	}

	for key := range subAccount.Nodes {
		if key == nodePK {
			continue
		}
		byte32 := sha256.Sum256(append([]byte(key), []byte(eventID)...))
		competitor.SetBytes(byte32[:])
		if -1 == competitor.Cmp(lottery) {
			met.logger.Debug("execute event notification", zap.String("return", "not winning the lottery"))
			return nil
		}
	}

	fmt.Println("my turn to fetch event")

	// Now this node is qualified to fetch event
	suber, ok := subOrg.Application.(EventSubscriber)
	if !ok {
		return fmt.Errorf("%s doesn't implement EventSubscriber", tx.Listener)
	}

	// EventSubscriber.IncomingEvent return false if the app descides not to fetch the event on its arrival
	if !suber.IncomingEvent(tx.Source, tx.Height) {
		return nil
	}

	accnt, err := met.OrgState.GetAccount(tx.Source)
	if err != nil {
		return err
	}
	nodes := accnt.GetNodes()
	var i int
	var pkStr string
	l := rand.Intn(len(nodes))
	for pkStr = range nodes {
		if i == l {
			break
		}
		i++
	}

	// get remote event address from orgstate attributes
	if eventAddr, ok := nodes[pkStr]["event_external_address"]; ok {
		go func() {
			event, err := met.fetchEvent(eventAddr, eventID, tx.DataHash)
			if err != nil {
				met.logger.Error("[fetch event error]", zap.Error(err))
				return
			}

			eventData, err := DecodeEventData(event)
			if err != nil {
				met.logger.Error("[decode event error]", zap.Error(err))
				return
			}

			if err := suber.ConfirmEvent(tx); err != nil {
				met.logger.Error("[confirm event error]", zap.Error(err))
			}

			listenerAccnt, _ := met.EventState.GetAccount(tx.Listener)
			codeHash := listenerAccnt.GetPublishers()[tx.Source]
			codeBytes := met.EventCodeBase.Get(codeHash)

			var refinedData EventData
			// event lua on the receiver side is optional
			if len(codeBytes) > 0 {
				luas := cvtools.NewLuaState()
				defer luas.Close()
				var err error
				refinedData, err = cvtools.ExecLuaWithParam(luas, string(codeBytes), eventData)
				if err != nil {
					met.logger.Warn("[publish_event], exec lua code err", zap.Error(err))
					return
				} else if len(refinedData) == 0 {
					return
				}
			}

			suber.HandleEvent(refinedData, tx)
		}()
	}
	return nil
}

func (met *Metropolis) executeEventUnsubscribeTx(tx *EventUnsubscribeTx, block *agtypes.BlockCache) error {
	if err := met.checkEventUnsubscribeTx(tx); err != nil {
		return err
	}

	listener, _ := met.EventState.GetAccount(tx.Listener)
	source, _ := met.EventState.GetAccount(tx.Source)

	// remove event code
	met.EventCodeBase.DeleteSync(listener.MyPubers[tx.Source])
	met.EventCodeBase.DeleteSync(listener.MySubers[tx.Listener])

	listener.RemovePublisher(tx.Source)
	source.RemoveSubscriber(tx.Listener)

	met.logger.Info(fmt.Sprintf("%s is no longer listening on %s\n", tx.Listener, tx.Source))
	return nil
}

func (met *Metropolis) executeEventUploadCodeTx(tx *EventUploadCodeTx, block *agtypes.BlockCache) error {
	if _, err := met.GetOrg(tx.Owner); err == nil {
		codehash, _ := tools.HashRipemd160([]byte(tx.Code))
		met.EventCodeBase.SetSync(codehash, []byte(tx.Code))
	}

	return nil
}

func (met *Metropolis) executeEcoTx(bs []byte, height def.INT) error {
	txBytes := agtypes.UnwrapTx(bs)

	switch {
	case IsEcoMortgageTx(bs):
		tx := &EcoMortgageTx{}
		if err := cvtools.TxFromBytes(txBytes, tx); err != nil {
			return errors.Wrap(err, "parse eco mortgage failed")
		}
		if ok, err := cvtools.TxVerifySignature(tx); !ok {
			return errors.Wrap(err, "eco mortgage fail to verify signature")
		}
		if err := met.executeEcoMortgageTx(tx, height); err != nil {
			return errors.Wrap(err, "exec eco mortgage failed")
		}
	case IsEcoRedemptionTx(bs):
		tx := &EcoRedemptionTx{}
		if err := cvtools.TxFromBytes(txBytes, tx); err != nil {
			return errors.Wrap(err, "parse eco redemption failed")
		}
		if ok, err := cvtools.TxVerifySignature(tx); !ok {
			return errors.Wrap(err, "eco redemption fail to verify signature")
		}
		if err := met.executeEcoRedemptionTx(tx, height); err != nil {
			return errors.Wrap(err, "exec eco redemption failed")
		}
	case IsEcoTransferTx(bs):
		tx := &EcoTransferTx{}
		if err := cvtools.TxFromBytes(txBytes, tx); err != nil {
			return errors.Wrap(err, "parse eco transfer failed")
		}
		if ok, err := cvtools.TxVerifySignature(tx); !ok {
			return errors.Wrap(err, "eco transfer fail to verify signature")
		}
		if err := met.executeEcoTransferTx(tx, height); err != nil {
			return errors.Wrap(err, "exec transfer redemption failed")
		}
	case IsEcoInitAllocTx(bs):
		if height != 1 {
			return fmt.Errorf("Invalid init alloc tx")
		}
		tx := &EcoInitAllocTx{}
		if err := json.Unmarshal(txBytes, tx); err != nil {
			return errors.Wrap(err, "parse eco init alloc failed")
		}
		if err := met.executeEcoInitAllocTx(tx); err != nil {
			return errors.Wrap(err, "exec eco init alloc failed")
		}
	}

	return nil
}
