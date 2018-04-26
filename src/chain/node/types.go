package node

import (
	"github.com/spf13/viper"
	"go.uber.org/zap"

	pbtypes "github.com/Baptist-Publication/angine/protos/types"
	agtypes "github.com/Baptist-Publication/angine/types"
	"github.com/Baptist-Publication/chorus-module/lib/go-crypto"
	"github.com/Baptist-Publication/chorus-module/xlib/def"
)

type (
	// Application embeds types.Application, defines application interface in chorus
	Application interface {
		agtypes.Application
		SetCore(Core)
		GetAttributes() AppAttributes
	}

	// Core defines the interface at which an application sees its containing organization
	Core interface {
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
		Broadcaster
	}

	// Broadcaster means we can deliver tx in application
	Broadcaster interface {
		BroadcastTx([]byte) error
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
