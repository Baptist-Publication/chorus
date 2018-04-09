package node

import (
	"encoding/json"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"sort"

	pbtypes "github.com/Baptist-Publication/angine/protos/types"
	agtypes "github.com/Baptist-Publication/angine/types"
	"github.com/Baptist-Publication/chorus-module/lib/go-crypto"
	"github.com/Baptist-Publication/chorus-module/xlib/def"
)

type (
	EventData map[string]interface{}

	eventItem struct {
		Name  string
		Value interface{}
	}

	event []eventItem
)

func newEventItem(n string, v interface{}) eventItem {
	return eventItem{
		Name:  n,
		Value: v,
	}
}

func EncodeEventData(data EventData) ([]byte, error) {
	// buf := new(bytes.Buffer)
	// encoder := gob.NewEncoder(buf)
	// if err := encoder.Encode(data); err != nil {
	// 	return nil, errors.Wrap(err, "[encodeEventData]")
	// }

	// return buf.Bytes(), nil

	sortedK := make([]string, 0)
	for k := range data {
		sortedK = append(sortedK, k)
	}
	sort.Strings(sortedK)

	ev := make(event, len(sortedK))
	for i, k := range sortedK {
		ev[i] = newEventItem(k, data[k])
	}

	return json.Marshal(ev)
}

func DecodeEventData(bs []byte) (EventData, error) {
	// data := make(map[string]interface{})
	// if err := gob.NewDecoder(bytes.NewReader(bs)).Decode(&data); err != nil {
	// 	return nil, errors.Wrap(err, "[decodeEventData]")
	// }
	// return data, nil
	ev := make(event, 0)
	err := json.Unmarshal(bs, &ev)
	if err != nil {
		return nil, err
	}

	data := make(EventData)
	for _, v := range ev {
		data[v.Name] = v.Value
	}

	return data, err
}

type (
	// Application embeds types.Application, defines application interface in chorus
	Application interface {
		agtypes.Application
		SetCore(Core)
		GetAttributes() AppAttributes
	}

	// Core defines the interface at which an application sees its containing organization
	Core interface {
		Publisher

		IsValidator() bool
		GetPublicKey() (crypto.PubKeyEd25519, bool)
		GetPrivateKey() (crypto.PrivKeyEd25519, bool)
		GetChainID() string
		GetEngine() Engine
		BroadcastTxSuperior([]byte) error
	}

	// Engine defines the consensus engine
	Engine interface {
		GetBlock(def.INT) (*agtypes.BlockCache, *pbtypes.BlockMeta, error)
		GetBlockMeta(def.INT) (*pbtypes.BlockMeta, error)
		GetValidators() (def.INT, *agtypes.ValidatorSet)
		PrivValidator() *agtypes.PrivValidator
		BroadcastTx([]byte) error
		Query(byte, []byte) (interface{}, error)
	}

	// Superior defines the application on the upper level, e.g. Metropolis
	Superior interface {
		Publisher
		Broadcaster
	}

	// Broadcaster means we can deliver tx in application
	Broadcaster interface {
		BroadcastTx([]byte) error
	}

	// Publisher means that we can publish events
	Publisher interface {
		// PublishEvent
		// if data is neither tx nor []tx, the related tx hash should be given accordingly
		PublishEvent(from string, block *agtypes.BlockCache, data []EventData, txhash []byte) error
		CodeExists([]byte) bool
	}

	// Serializable transforms to bytes
	Serializable interface {
		ToBytes() ([]byte, error)
	}

	// Unserializable transforms from bytes
	Unserializable interface {
		FromBytes(bs []byte)
	}

	// Hashable aliases Serializable
	Hashable interface {
		Serializable
	}

	// AppMaker is the signature for functions which take charge of create new instance of applications
	AppMaker func(*zap.Logger, *viper.Viper, crypto.PrivKey) (Application, error)
)

// AppAttributes is just a type alias
type AppAttributes = map[string]string

type IMetropolisApp interface {
	GetAttribute(string) (string, bool)
	GetAttributes() AppAttributes
	SetAttributes(AppAttributes)
	PushAttribute(string, string)
	AttributeExists(string) bool
}
