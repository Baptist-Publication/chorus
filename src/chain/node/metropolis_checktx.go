package node

import (
	"bytes"
	"encoding/hex"
	"fmt"

	cosiED "github.com/bford/golang-x-crypto/ed25519"
	"github.com/bford/golang-x-crypto/ed25519/cosi"
	"github.com/pkg/errors"
	anginetypes "github.com/Baptist-Publication/angine/types"
	cvtools "github.com/Baptist-Publication/chorus/src/tools"
)

// CheckTx just contains a big switch which multiplex different kinds of transactions supported
func (met *Metropolis) CheckTx(bs []byte) error {
	txBytes := anginetypes.UnwrapTx(bs)
	switch {
	case IsOrgCancelTx(bs):
		tx := &OrgCancelTx{}
		if err := cvtools.TxFromBytes(txBytes, tx); err != nil {
			return err
		}
		if ok, err := cvtools.TxVerifySignature(tx); !ok {
			return err
		}
		if err := met.checkOrgCancel(tx); err != nil {
			return err
		}
	case IsOrgConfirmTx(bs):
		tx := &OrgConfirmTx{}
		if err := cvtools.TxFromBytes(txBytes, tx); err != nil {
			return err
		}
		if ok, err := cvtools.TxVerifySignature(tx); !ok {
			return err
		}
		if err := met.checkOrgConfirm(tx); err != nil {
			return err
		}
	case IsOrgTx(bs):
		tx := &OrgTx{}
		if err := cvtools.TxFromBytes(txBytes, tx); err != nil {
			return err
		}
		if ok, err := cvtools.TxVerifySignature(tx); !ok {
			return err
		}
		if err := met.checkOrgs(tx); err != nil {
			return err
		}
		if err := met.checkOrgState(tx); err != nil {
			return err
		}
	case IsEventUploadCodeTx(bs):
		tx := &EventUploadCodeTx{}
		if err := cvtools.TxFromBytes(txBytes, tx); err != nil {
			return err
		}
		if ok, err := cvtools.TxVerifySignature(tx); !ok {
			return err
		}
		if err := met.checkEventUploadCodeTx(tx); err != nil {
			return err
		}
	case IsEventRequestTx(bs):
		tx := &EventRequestTx{}
		if err := cvtools.TxFromBytes(txBytes, tx); err != nil {
			return err
		}
		if ok, err := cvtools.TxVerifySignature(tx); !ok {
			return err
		}
		if err := met.checkEventRequestTx(tx); err != nil {
			return err
		}
	case IsEventSubscribeTx(bs):
		tx := &EventSubscribeTx{}
		if err := cvtools.TxFromBytes(txBytes, tx); err != nil {
			return errors.Wrap(err, "load EventSubscribeTx failed")
		}
		if ok, err := cvtools.TxVerifySignature(tx); !ok {
			return err
		}
		if err := met.checkEventSubscribeTx(tx); err != nil {
			return errors.Wrap(err, "check EventSubscribeTx failed")
		}
	case IsEventUnsubscribeTx(bs):
		tx := &EventUnsubscribeTx{}
		if err := cvtools.TxFromBytes(txBytes, tx); err != nil {
			return errors.Wrap(err, "load EventUnsubscribeTx failed")
		}
		if ok, err := cvtools.TxVerifySignature(tx); !ok {
			return err
		}
		if err := met.checkEventUnsubscribeTx(tx); err != nil {
			return errors.Wrap(err, "check EventUnsubscribeTx failed")
		}
	case IsCoSiInitTx(bs):
		tx := &CoSiInitTx{}
		if err := cvtools.TxFromBytes(txBytes, tx); err != nil {
			return err
		}
		if ok, err := cvtools.TxVerifySignature(tx); !ok {
			return err
		}
		if err := met.checkCoSiInitTx(tx); err != nil {
			return err
		}
	case IsEventNotificationTx(bs):
		tx := &EventNotificationTx{}
		if err := cvtools.TxFromBytes(txBytes, tx); err != nil {
			return err
		}
		if ok, err := cvtools.TxVerifySignature(tx); !ok {
			return err
		}
		if err := met.checkEventNotificationTx(tx); err != nil {
			return err
		}

	case IsEcoTx(bs):
		if err := met.checkEcoTx(bs); err != nil {
			return err
		}
	}

	return nil
}

func (met *Metropolis) checkOrgCancel(tx *OrgCancelTx) error {
	orgTx, ok := met.PendingOrgTxs[string(tx.TxHash)]
	if !ok {
		return fmt.Errorf("cancel for nothing")
	}
	if v, e := cvtools.TxVerifySignature(orgTx); !v || e != nil {
		return fmt.Errorf("org tx signature verification failed")
	}

	return nil
}

func (met *Metropolis) checkOrgConfirm(tx *OrgConfirmTx) error {
	orgTx, ok := met.PendingOrgTxs[string(tx.TxHash)]
	if !ok {
		return fmt.Errorf("confirm for something I don't know: %X", tx.TxHash)
	}
	if res, err := cvtools.TxVerifySignature(orgTx); !res || err != nil {
		return fmt.Errorf("org tx signature verification failed`")
	}

	if !bytes.Equal(tx.PubKey, orgTx.PubKey) {
		return fmt.Errorf("pubkey mismatch, org:%X, tx:%X ", orgTx.PubKey, tx.PubKey)
	}
	return nil
}

func (met *Metropolis) checkOrgs(tx *OrgTx) error {
	pubkey, _ := met.core.GetPublicKey()
	if !bytes.Equal(pubkey[:], tx.PubKey) {
		return nil // leave it to other nodes
	}
	switch tx.Act {
	case OrgCreate:
		met.mtx.Lock()
		if _, ok := met.Orgs[ChainID(tx.ChainID)]; ok {
			met.mtx.Unlock()
			return ErrOrgAlreadyIn
		}
		met.mtx.Unlock()
		return nil
	case OrgJoin:
		met.mtx.Lock()
		if _, ok := met.Orgs[ChainID(tx.ChainID)]; ok {
			met.mtx.Unlock()
			return fmt.Errorf("already joined: %s", tx.ChainID)
		}
		met.mtx.Unlock()
		return nil
	case OrgLeave:
		met.mtx.Lock()
		if _, ok := met.Orgs[ChainID(tx.ChainID)]; !ok {
			met.mtx.Unlock()
			return fmt.Errorf("not in organization: %s", tx.ChainID)
		}
		met.mtx.Unlock()
		return nil
	default:
		return fmt.Errorf("unimplemented act: %v", tx.Act)
	}
}

func (met *Metropolis) checkOrgState(tx *OrgTx) error {
	pubkey, _ := met.core.GetPublicKey()
	if !bytes.Equal(pubkey[:], tx.PubKey) {
		return nil // leave it to other nodes
	}
	switch tx.Act {
	case OrgCreate:
		if met.OrgState.ExistAccount(tx.ChainID) {
			return ErrOrgExistsAlready
		}
		return nil
	case OrgJoin:
		if !met.OrgState.ExistAccount(tx.ChainID) {
			return ErrOrgNotExists
		}
		return nil
	case OrgLeave:

		return nil
	default:
		return fmt.Errorf("unimplemented act: %v", tx.Act)
	}
}

func (met *Metropolis) checkEventUploadCodeTx(tx *EventUploadCodeTx) error {
	// TODO syntax check
	// types of vars in lua is resolved in real time
	// so can't do further checks here
	return nil
}

func (met *Metropolis) checkEventRequestTx(tx *EventRequestTx) error {
	_, errSource := met.GetOrg(tx.Source)
	_, errListener := met.GetOrg(tx.Listener)

	if errSource == nil {
		if codeBytes := met.EventCodeBase.Get(tx.SourceHash); codeBytes == nil || len(codeBytes) <= 0 {
			fmt.Printf("event source(%s) doesn't own the code specified", tx.Source)
			return errors.Errorf("event source(%s) doesn't own the code specified", tx.Source)
		}
		return nil
	}

	if errListener == nil {
		if codeBytes := met.EventCodeBase.Get(tx.ListenerHash); codeBytes == nil || len(codeBytes) <= 0 {
			fmt.Printf("event listener(%s) doesn't own the code specified", tx.Listener)
			return errors.Errorf("event listener(%s) doesn't own the code specified", tx.Listener)
		}
		return nil
	}

	return errors.Errorf("invalid tx receiver, related organizations:%s, %s", tx.Source, tx.Listener)
}

// checkEventSubscribeTx verifies ED25519 CoSignature
func (met *Metropolis) checkEventSubscribeTx(tx *EventSubscribeTx) error {
	if _, ok := met.PendingEventRequestTxs[string(tx.TxHash)]; !ok {
		return errors.Wrap(errors.Errorf("no such event request: %X", tx.TxHash), "checkEventSubscribeTx")
	}
	accnt, err := met.OrgState.GetAccount(tx.Source)
	if err != nil {
		return errors.Wrap(err, "checkEventSubscribeTx")
	}
	var pks []cosiED.PublicKey
	for k := range accnt.GetNodes() {
		pk, err := hex.DecodeString(k)
		if err != nil {
			return err
		}
		pks = append(pks, cosiED.PublicKey(pk))
	}

	policy := cosi.ThresholdPolicy(tx.Threshold)

	if !cosi.Verify(pks, policy, tx.SignData, tx.Signature) {
		return errors.Wrap(err, "checkEventSubscribeTx")
	}
	return nil
}

func (met *Metropolis) checkEventUnsubscribeTx(tx *EventUnsubscribeTx) error {
	sourceAccount, err := met.EventState.GetAccount(tx.Source)
	if err != nil {
		return err
	}
	listenerAccount, err := met.EventState.GetAccount(tx.Listener)
	if err != nil {
		return err
	}
	if _, ok := sourceAccount.GetSubscribers()[tx.Listener]; !ok {
		return errors.Errorf("illegal relation: %s is not a subscriber of %s", tx.Listener, tx.Source)
	}
	if _, ok := listenerAccount.GetPublishers()[tx.Source]; !ok {
		return errors.Errorf("illegal relation: %s is not a publisher of %s", tx.Source, tx.Listener)
	}

	// verify the provided proof here
	_ = tx.Proof

	return nil
}

// checkCoSiInitTx checks the cosi leader is indeed the maker of the block at that height
// so, in a cosi round, the leader doesn't have to be a member of the subchain. Just a validator on the metropolis is enough.
func (met *Metropolis) checkCoSiInitTx(tx *CoSiInitTx) error {
	block, _, err := met.core.GetEngine().GetBlock(tx.Height)
	if err != nil {
		return err
	}
	if !bytes.Equal(block.Header.Maker, tx.PubKey) {
		return fmt.Errorf("illegal tx pubkey: %X, expected: %X", tx.PubKey, block.Header.Maker)
	}

	return nil
}

func (met *Metropolis) checkEventNotificationTx(tx *EventNotificationTx) error {
	sourceAccnt, err := met.EventState.GetAccount(tx.Source)
	if err != nil {
		return err
	}
	listenerAccnt, err := met.EventState.GetAccount(tx.Listener)
	if err != nil {
		return err
	}
	sourceSubscribers := sourceAccnt.GetSubscribers()
	if _, ok := sourceSubscribers[tx.Listener]; !ok {
		return errors.Errorf("%s is not one of the subscribers of %s", tx.Listener, tx.Source)
	}
	listenerPublishers := listenerAccnt.GetPublishers()
	if _, ok := listenerPublishers[tx.Source]; !ok {
		return errors.Errorf("%s is not in the publisher list of %s", tx.Source, tx.Listener)
	}

	return nil
}

func (met *Metropolis) checkEcoTx(bs []byte) error {
	// parse tx
	cvTx, err := met.GetEcoCivilTx(bs)
	if err != nil {
		return err
	}

	// verify sig
	if ok, err := cvtools.TxVerifySignature(cvTx); !ok {
		return errors.Wrap(err, "eco tx fail to verify signature")
	}

	// check nonce
	acc, _ := met.accState.GetAccount(cvTx.GetPubKey())
	if acc != nil && acc.Nonce > cvTx.GetNonce() {
		return errors.New("Invalid nonce")
	}

	return nil
}
