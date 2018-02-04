package node

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"math/big"
	"path"
	"strconv"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"github.com/tendermint/tmlibs/db"
	"go.uber.org/zap"

	//"github.com/Baptist-Publication/angine"
	ac "github.com/Baptist-Publication/angine/config"
	pbtypes "github.com/Baptist-Publication/angine/protos/types"
	agtypes "github.com/Baptist-Publication/angine/types"
	cmn "github.com/Baptist-Publication/chorus-module/lib/go-common"
	"github.com/Baptist-Publication/chorus-module/lib/go-crypto"
	"github.com/Baptist-Publication/chorus-module/lib/go-merkle"
	"github.com/Baptist-Publication/chorus-module/xlib"
	"github.com/Baptist-Publication/chorus-module/xlib/def"
	acfg "github.com/Baptist-Publication/chorus/src/chain/config"
	"github.com/Baptist-Publication/chorus/src/chain/log"
)

const (
	BASE_APP_NAME = "metropolis"

	StartRedeemHeight = 15000
)

type (
	// AppName just wraps around string, maybe we can use the new feat in Go1.9
	AppName string

	// ChainID just wraps around string
	ChainID string

	// MetropolisState abstracts the structure of the information
	// that we need to make the application consistency verifiable
	MetropolisState struct {
		OrgStateHash   []byte
		OrgHeights     []def.INT
		EventStateHash []byte
		AccStateHash   []byte
		PowerStateHash []byte
	}

	// Metropolis defines the application
	Metropolis struct {
		agtypes.BaseApplication

		mtx  sync.Mutex
		node *Node

		// core abstracts the orgnode within application
		core Core

		config      *viper.Viper
		angineHooks agtypes.Hooks
		logger      *zap.Logger

		// top level app state
		state *MetropolisState

		// Organization
		OrgStateDB db.DB
		// OrgState is a DL that keeps tracks of subchains' state
		OrgState *OrgState
		OrgApps  map[ChainID]AppName
		// Orgs contains all the subchains this node is effectively in
		Orgs map[ChainID]*OrgNode
		// TODO: persist onto disk
		PendingOrgTxs map[string]*OrgTx

		// Events
		EventStateDB   db.DB
		EventState     *EventState
		EventWarehouse *EventWarehouse
		EventCodeBase  db.DB
		// TODO: persist onto disk
		PendingEventRequestTxs map[string]*EventRequestTx

		accStateDB db.DB
		accState   *AccState

		powerStateDB db.DB
		powerState   *PowerState

		FeeAccum *big.Int
	}

	// LastBlockInfo is just a must for every angine-based application
	LastBlockInfo struct {
		Height def.INT `msgpack:"height"`
		// hash from the top level state
		Hash []byte `msgpack:"hash"`
	}
)

var (
	// ErrUnknownTx is exported because I am not sure if it will be needed outside of this pkg
	ErrUnknownTx = fmt.Errorf("please give me something that I actually know about")

	Big0 = new(big.Int).SetUint64(0)
)

// NewLastBlockInfo just a convience to generate an empty LastBlockInfo
func NewLastBlockInfo() *LastBlockInfo {
	return &LastBlockInfo{
		Height: 0,
		Hash:   make([]byte, 0),
	}
}

// NewMetropolis initialize all the necessary parts of the application:
// 1. state of metropolis
// 2. init BaseApplication
// 3. open databases and generate orgstate, eventstate and so on
// 4. set up the angine hooks
func NewMetropolis(logger *zap.Logger, conf *viper.Viper) *Metropolis {
	datadir := conf.GetString("db_dir")
	met := Metropolis{
		config: conf,
		logger: logger,

		state: &MetropolisState{
			OrgHeights:     make([]def.INT, 0),
			OrgStateHash:   make([]byte, 0),
			EventStateHash: make([]byte, 0),
			AccStateHash:   make([]byte, 0),
			PowerStateHash: make([]byte, 0),
		},

		OrgApps: make(map[ChainID]AppName),
		Orgs:    make(map[ChainID]*OrgNode),

		PendingOrgTxs:          make(map[string]*OrgTx),
		PendingEventRequestTxs: make(map[string]*EventRequestTx),
	}

	var err error
	if err = met.BaseApplication.InitBaseApplication(BASE_APP_NAME, datadir); err != nil {
		cmn.PanicCrisis(err)
	}
	if met.OrgStateDB, err = db.NewGoLevelDB("orgStatus", datadir); err != nil {
		cmn.PanicCrisis(err)
	}
	met.OrgState = NewOrgState(met.OrgStateDB)
	if met.EventStateDB, err = db.NewGoLevelDB("eventstate", datadir); err != nil {
		cmn.PanicCrisis(err)
	}
	met.EventState = NewEventState(met.EventStateDB)
	ewhDB, err := db.NewGoLevelDB("eventwarehouse", datadir)
	if err != nil {
		cmn.PanicCrisis(err)
	}
	met.EventWarehouse = NewEventWarehouse(ewhDB)

	met.EventCodeBase, err = db.NewGoLevelDB("eventcodebase", datadir)
	if err != nil {
		cmn.PanicCrisis(err)
	}

	if met.accStateDB, err = db.NewGoLevelDB("accountstate", datadir); err != nil {
		cmn.PanicCrisis(err)
	}
	met.accState = NewAccState(met.accStateDB)

	if met.powerStateDB, err = db.NewGoLevelDB("powerstate", datadir); err != nil {
		cmn.PanicCrisis(err)
	}
	met.powerState = NewPowerState(met.powerStateDB)

	met.angineHooks = agtypes.Hooks{
		OnExecute: agtypes.NewHook(met.OnExecute),
		OnCommit:  agtypes.NewHook(met.OnCommit),
	}

	return &met
}

// Hash is merely a wrapper of MetropolisState.Hash
func (met *Metropolis) Hash() []byte {
	return met.state.Hash()
}

// Lock locks Metropolis universally
func (met *Metropolis) Lock() {
	met.mtx.Lock()
}

// Unlock undos what Lock does
func (met *Metropolis) Unlock() {
	met.mtx.Unlock()
}

// SetNode
func (met *Metropolis) SetNode(n *Node) {
	met.node = n
}

// SetCore gives Metropolis a way to make some calls to the node running underneath
// it is more abstractive than SetNode, which I am planning to deprecate.
func (met *Metropolis) SetCore(c Core) {
	met.core = c
}

// BroadcastTx just passes the transaction into core
func (met *Metropolis) BroadcastTx(tx []byte) error {
	return met.core.GetEngine().BroadcastTx(tx)
}

// GetNodePubKey gets our universal public key
func (met *Metropolis) GetNodePubKey() crypto.PubKey {
	return met.core.GetEngine().PrivValidator().GetPubKey()
}

// Stop stops all still running organization
func (met *Metropolis) Stop() {
	for i := range met.Orgs {
		met.Orgs[i].Stop()
	}
}

// Load gets all those ledgers have been persisted back alive by a metropolisstate hash
func (met *Metropolis) Load(lb *LastBlockInfo) error {
	met.Lock()
	defer met.Unlock()

	state := &MetropolisState{}
	if lb == nil {
		return nil
	}

	bs := met.Database.Get(lb.Hash)
	if err := state.FromBytes(bs); err != nil {
		return errors.Wrap(err, "fail to restore metropolis state")
	}
	met.state = state
	met.OrgState.Load(met.state.OrgStateHash)
	met.EventState.Load(met.state.EventStateHash)
	if met.config.GetBool("enable_incentive") {
		met.accState.Load(met.state.AccStateHash)
		met.powerState.Load(met.state.PowerStateHash)
	}
	return nil
}

// Start will restore organization according to orgtx history
func (met *Metropolis) Start() error {
	if err := met.spawnOffchainEventListener(); err != nil {
		met.logger.Error("fail to start event server", zap.Error(err))
		return err
	}

	lastBlock := NewLastBlockInfo()
	if res, err := met.LoadLastBlock(lastBlock); err == nil {
		lastBlock = res.(*LastBlockInfo)
	}
	if lastBlock.Hash == nil || len(lastBlock.Hash) == 0 {
		return nil
	}
	if err := met.Load(lastBlock); err != nil {
		met.logger.Error("fail to load metropolis state", zap.Error(err))
		return err
	}

	for _, height := range met.state.OrgHeights {
		block, _, err := met.core.GetEngine().GetBlock(def.INT(height))
		if err != nil {
			return err
		}
		for _, tx := range block.Data.Txs {
			txBytes := agtypes.UnwrapTx(tx)
			switch {
			case IsOrgCancelTx(tx):
				met.restoreOrgCancel(txBytes)
			case IsOrgConfirmTx(tx):
				met.restoreOrgConfirm(txBytes)
			case IsOrgTx(tx):
				met.restoreOrg(txBytes)
			}
		}
	}

	return nil
}

// GetAngineHooks returns the hooks we defined for angine to grasp
func (met *Metropolis) GetAngineHooks() agtypes.Hooks {
	return met.angineHooks
}

func (met *Metropolis) GetAttributes() map[string]string {
	return nil
}

// CompatibleWithAngine just exists to satisfy agtypes.Application
func (met *Metropolis) CompatibleWithAngine() {}

// OnCommit persists state that we define to be consistent in a cross-block way
func (met *Metropolis) OnCommit(height, round def.INT, block *agtypes.BlockCache) (interface{}, error) {
	var err error

	if met.state.OrgStateHash, err = met.OrgState.Commit(); err != nil {
		met.logger.Error("fail to commit orgState", zap.Error(err))
		return nil, err
	}
	if met.state.EventStateHash, err = met.EventState.Commit(); err != nil {
		met.logger.Error("fail to commit eventState", zap.Error(err))
		return nil, err
	}

	if met.config.GetBool("enable_incentive") {
		if met.state.AccStateHash, err = met.accState.Commit(); err != nil {
			met.logger.Error("fail to commit accState", zap.Error(err))
			return nil, err
		}
		if met.state.PowerStateHash, err = met.powerState.Commit(); err != nil {
			met.logger.Error("fail to commit powerState", zap.Error(err))
			return nil, err
		}
	}

	met.state.OrgHeights = sterilizeOrgHeights(met.state.OrgHeights)

	lastBlock := LastBlockInfo{Height: height, Hash: met.Hash()}
	met.Database.SetSync(lastBlock.Hash, met.state.ToBytes())
	met.SaveLastBlock(lastBlock)

	defer func() {
		met.OrgState.Reload(met.state.OrgStateHash)
		met.EventState.Reload(met.state.EventStateHash)
		if met.config.GetBool("enable_incentive") {
			met.accState.Reload(met.state.AccStateHash)
			met.powerState.Reload(met.state.PowerStateHash)
		}
	}()

	fmt.Println("height:", height, "fee:", met.FeeAccum.String(), "account size:", met.accState.Size(), "power size:", met.powerState.Size())

	return agtypes.CommitResult{AppHash: lastBlock.Hash}, nil
}

// IsTxKnown is a fast way to identify txs unknown
func (met *Metropolis) IsTxKnown(bs []byte) bool {
	return IsOrgRelatedTx(bs)
}

// CheckTx returns nil if sees an unknown tx

// Info gives information about the application in general
func (met *Metropolis) Info() (resInfo agtypes.ResultInfo) {
	lb := NewLastBlockInfo()
	if res, err := met.LoadLastBlock(lb); err == nil {
		lb = res.(*LastBlockInfo)
	}

	resInfo.LastBlockAppHash = lb.Hash
	resInfo.LastBlockHeight = lb.Height
	resInfo.Version = "alpha 0.2"
	resInfo.Data = "metropolis with organizations and inter-organization communication"
	return
}

// SetOption can dynamicly change some options of the application
func (met *Metropolis) SetOption() {}

// Query now gives the ability to query transaction execution result with the transaction hash
func (met *Metropolis) Query(query []byte) agtypes.Result {
	if len(query) == 0 {
		return agtypes.NewResultOK([]byte{}, "Empty query")
	}
	var res agtypes.Result

	switch query[0] {
	case agtypes.QueryTxExecution:
		qryRes, err := met.core.GetEngine().Query(query[0], query[1:])
		if err != nil {
			return agtypes.NewError(pbtypes.CodeType_InternalError, err.Error())
		}
		info, ok := qryRes.(*agtypes.TxExecutionResult)
		if !ok {
			return agtypes.NewError(pbtypes.CodeType_InternalError, err.Error())
		}
		res.Code = pbtypes.CodeType_OK
		res.Data, _ = info.ToBytes()
		return res
	case QueryEvents:
		keys := make([]string, 0)
		for k := range met.EventWarehouse.state.state {
			keys = append(keys, k)
		}

		buffers := &bytes.Buffer{}
		encoder := gob.NewEncoder(buffers)
		if err := encoder.Encode(keys); err != nil {
			res.Code = pbtypes.CodeType_InternalError
			res.Log = err.Error()
		} else {
			res.Code = pbtypes.CodeType_OK
			res.Data = buffers.Bytes()
		}
	case QueryNonce:
		var pubkey crypto.PubKeyEd25519
		copy(pubkey[:], query[1:])
		nonce := met.accState.QueryNonce(&pubkey)
		res.Code = pbtypes.CodeType_OK
		res.Data = make([]byte, 8)
		binary.LittleEndian.PutUint64(res.Data, nonce)
	case QueryBalance:
		var pubkey crypto.PubKeyEd25519
		copy(pubkey[:], query[1:])
		balance := met.accState.QueryBalance(&pubkey)
		res.Code = pbtypes.CodeType_OK
		res.Data = []byte(balance.String())
	case QueryPower:
		var pubkey crypto.PubKeyEd25519
		copy(pubkey[:], query[1:])

		power, mHeight := met.powerState.QueryPower(&pubkey)
		bs, err := json.Marshal([]string{power.String(), strconv.Itoa(int(mHeight))})
		if err != nil {
			res.Code = pbtypes.CodeType_InternalError
		} else {
			res.Code = pbtypes.CodeType_OK
			res.Data = bs
		}
	}

	return res
}

// SetOrg wraps some atomic operations that need to be done when this node create/join an organization
func (met *Metropolis) SetOrg(chainID, appname string, o *OrgNode) {
	met.mtx.Lock()
	met.Orgs[ChainID(chainID)] = o
	met.OrgApps[ChainID(chainID)] = AppName(appname)
	met.mtx.Unlock()
}

// GetOrg gets the orgnode associated to the chainid, or gives an error if we are not connected to the chain
func (met *Metropolis) GetOrg(id string) (*OrgNode, error) {
	met.mtx.Lock()
	defer met.mtx.Unlock()
	if o, ok := met.Orgs[ChainID(id)]; ok {
		return o, nil
	}
	return nil, fmt.Errorf("no such org: %s", id)
}

// RemoveOrg wraps atomic operations needed to be done when node gets removed from an organization
func (met *Metropolis) RemoveOrg(id string) error {
	met.mtx.Lock()
	orgNode := met.Orgs[ChainID(id)]
	orgNode.Stop()
	if orgNode.IsRunning() && !orgNode.Stop() {
		met.mtx.Unlock()
		return fmt.Errorf("node is still running, error cowordly")
	}
	delete(met.Orgs, ChainID(id))
	delete(met.OrgApps, ChainID(id))
	met.mtx.Unlock()
	return nil
}

// createOrgNode start a new org node in goroutine, so the app could use runtime.Goexit when error occurs
func (met *Metropolis) createOrgNode(tx *OrgTx) (*OrgNode, error) {
	if !AppExists(tx.App) {
		return nil, fmt.Errorf("no such app: %s", tx.App)
	}
	runtime := ac.RuntimeDir(met.config.GetString("runtime"))
	conf, err := acfg.LoadDefaultAngineConfig(runtime, tx.ChainID, tx.Config)
	if err != nil {
		return nil, err
	}

	if !cmn.FileExists(conf.GetString("priv_validator_file")) {
		privVal := met.core.GetEngine().PrivValidator().CopyReset()
		ppv := &privVal
		ppv.SetFile(conf.GetString("priv_validator_file"))
		ppv.Save()
	}

	if tx.Genesis.ChainID != "" && !cmn.FileExists(conf.GetString("genesis_file")) {
		if err := tx.Genesis.SaveAs(conf.GetString("genesis_file")); err != nil {
			met.logger.Error("fail to save org's genesis", zap.String("chainid", tx.ChainID), zap.String("path", conf.GetString("genesis_file")))
			return nil, fmt.Errorf("fail to save org's genesis")
		}
	}

	resChan := make(chan *OrgNode, 1)
	go func(c chan<- *OrgNode) {
		applog := log.Initialize(met.config.GetString("environment"), path.Join(conf.GetString("log_path"), tx.ChainID, "output.log"), path.Join(conf.GetString("log_path"), tx.ChainID, "err.log"))
		n := NewOrgNode(applog, tx.App, conf, met)
		if n == nil {
			met.logger.Error("startOrgNode failed", zap.Error(err))
			c <- nil
			return
		}
		if err := n.Start(); err != nil {
			met.logger.Error("startOrgNode failed", zap.Error(err))
			c <- nil
			return
		}
		c <- n
	}(resChan)

	var node *OrgNode
	select {
	case node = <-resChan:
	case <-time.After(10 * time.Second):
	}

	if node != nil {
		fmt.Printf("organization %s is running\n", node.GetChainID())
		return node, nil
	}

	return nil, fmt.Errorf("fail to start org node, check the log")
}

func (met *Metropolis) RegisterValidators(validatorset *agtypes.ValidatorSet) {
	for _, validator := range validatorset.Validators {
		if pub, ok := validator.GetPubKey().(*crypto.PubKeyEd25519); ok {
			met.accState.CreateAccount(pub[:], Big0)
			met.powerState.CreatePower(pub[:], new(big.Int).SetUint64(uint64(validator.VotingPower)), 1)
		}
	}
}

func (met *Metropolis) SuspectValidator(pubkey []byte, reason string) {
	pwr, err := met.powerState.GetPower(pubkey)
	if err != nil {
		met.logger.Error("Query accstate failed: " + err.Error())
		return
	}

	var pk crypto.PubKeyEd25519
	copy(pk[:], pubkey)
	if err := met.powerState.SubVTPower(&pk, new(big.Int).Div(pwr.VTPower, new(big.Int).SetUint64(10)), pwr.MHeight); err != nil {
		met.logger.Error("SubVTPower failed: " + err.Error())
		return
	}

	met.logger.Info(fmt.Sprintf("Do punish account: %X\n", pubkey))
	fmt.Printf("Do punish[%s] account: %X\n", reason, pubkey)
}

type comparablePower Power

func (ca *comparablePower) Less(o interface{}) bool {
	return ca.VTPower.Cmp(o.(*comparablePower).VTPower) < 0
}

// n is the number of participants
func CalcBase(n int) int {
	if n <= 6 {
		// avg, do not use triangle
		return 0
	}
	for i := 5; ; i++ {
		if n <= i*(i+1)/2 {
			return i
		}
	}
}

func CalcVP(base int, position int) uint64 {
	for level := 1; ; level++ {
		if position <= level*(level+1)/2 {
			return uint64((base - level + 1) * (base - level + 1))
		}
	}
}

func (met *Metropolis) ValSetLoader() agtypes.ValSetLoaderFunc {
	return func(height, round def.INT, size int) *agtypes.ValidatorSet {
		vals := met.fakeRandomVals(height, round, size)

		return agtypes.NewValidatorSet(vals)
	}
}

func (met *Metropolis) fakeRandomVals(height, round def.INT, size int) []*agtypes.Validator {
	pwrs := make([]*Power, 0, 21)

	// Iterate power list of world state
	// to get all power account than joining current election
	accList := make([]*Power, 0, met.powerState.Size())
	vsetHeap := cmn.NewHeap() // max heap of length 15
	met.powerState.Iterate(func(pwr *Power) bool {
		if pwr.MHeight == -1 { // indicate he is not in
			return false
		}

		accList = append(accList, pwr)
		vsetHeap.Push(pwr, (*comparablePower)(pwr))
		if vsetHeap.Len() > 15 {
			vsetHeap.Pop()
		}
		return false
	})

	// Pick all and share the same power
	// if the length of partners is less than 6
	if len(accList) <= 6 {
		vals := make([]*agtypes.Validator, len(accList))
		for i, v := range accList {
			var pk crypto.PubKeyEd25519
			copy(pk[:], v.Pubkey)
			vals[i] = &agtypes.Validator{PubKey: crypto.StPubKey{&pk}, Address: pk.Address(), VotingPower: 100}
		}
		return vals
	}

	// Calculate the number of power account that need to select 'randomly' from the rest account
	// n = len(accList)
	//  if      n <= 15 	then 0
	//  if 15 < n <= 21 	then 15 - n
	//  if 21 < n			then 6
	numLuckyGuys := int(len(accList)) - 15
	if numLuckyGuys < 0 {
		numLuckyGuys = 0
	} else if numLuckyGuys > 6 {
		numLuckyGuys = 6
	}

	// Pick the rich-guys
	// max(rich-guys) = 15
	// min(rich-guys) = 7
	exists := make(map[*Power]struct{})
	i := int(vsetHeap.Len())
	for vsetHeap.Len() > 0 {
		pwr := vsetHeap.Pop().(*Power)
		pwr.VTPower = new(big.Int).SetUint64(CalcVP(6, i)) // re-calculate power
		pwrs = append(pwrs, pwr)
		exists[pwr] = struct{}{}
		i--
	}

	// Pick lucky-guys 'randomly' from the rest of power account
	// we use a map(means exists) to identify the elected guys
	retry := 1
	bigbang := new(big.Int).SetBytes(met.accState.Hash())
	luckyguys := make([]*Power, 0, numLuckyGuys)
	for len(luckyguys) < numLuckyGuys {
		guy, err := fakeRandomAccount(accList, exists, height, round, bigbang, &retry)
		if err != nil {
			fmt.Println("error in fakeRandomAccount:", err.Error())
			return nil
		}
		guy.VTPower = new(big.Int).SetUint64(1)
		luckyguys = append(luckyguys, guy)
	}

	// combine ...
	pwrs = append(pwrs, luckyguys...)

	// make validators according to the power accounts elected above
	vals := make([]*agtypes.Validator, len(pwrs))
	for i, v := range pwrs {
		var pk crypto.PubKeyEd25519
		copy(pk[:], v.Pubkey)
		vals[i] = &agtypes.Validator{
			PubKey:      crypto.StPubKey{&pk},
			Address:     pk.Address(),
			VotingPower: v.VTPower.Int64(),
		}
	}

	return vals
}

func fakeRandomAccount(accs []*Power, exists map[*Power]struct{}, height, round def.INT, bigbang *big.Int, retry *int) (*Power, error) {
	if len(accs) == len(exists) {
		return nil, fmt.Errorf("No account can be picked any more")
	}

	base := new(big.Int).Add(bigbang, new(big.Int).SetUint64(uint64(height+1)*uint64(*retry+1)))
	index := new(big.Int).Mod(base, new(big.Int).SetUint64(uint64(len(accs)))).Int64()
	for {
		(*retry)++
		tryPick := accs[index]
		if _, yes := exists[tryPick]; !yes {
			exists[tryPick] = struct{}{}
			return tryPick, nil
		}
		if index == int64(len(accs))-1 {
			index = 0
		} else {
			index++
		}
	}
}

func (ms *MetropolisState) Hash() []byte {
	return merkle.SimpleHashFromBinary(ms)
}

func (ms *MetropolisState) ToBytes() []byte {
	var bs []byte
	buf := bytes.NewBuffer(bs)
	gec := gob.NewEncoder(buf)
	if err := gec.Encode(ms); err != nil {
		panic(err)
	}
	return buf.Bytes()
}

func (ms *MetropolisState) FromBytes(bs []byte) error {
	bsReader := bytes.NewReader(bs)
	gdc := gob.NewDecoder(bsReader)
	return gdc.Decode(ms)
}

// remove duplicates and sort
func sterilizeOrgHeights(orgH []def.INT) []def.INT {
	hm := make(map[def.INT]struct{})
	for _, h := range orgH {
		hm[h] = struct{}{}
	}
	uniqueHeights := make([]def.INT, 0)
	for k := range hm {
		uniqueHeights = append(uniqueHeights, def.INT(k))
	}
	xlib.Int64Slice(uniqueHeights).Sort()
	return uniqueHeights
}
