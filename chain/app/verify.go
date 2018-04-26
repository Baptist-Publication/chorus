package app

import (
	"bytes"
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	agtypes "github.com/Baptist-Publication/chorus/angine/types"
	"github.com/Baptist-Publication/chorus/eth/rlp"
	"github.com/Baptist-Publication/chorus/types"
)

var (
	validateRoutineCount = runtime.NumCPU()
	// validateRoutineCount = 1
)

const (
	appTxStatusNone     int32 = 0 // new
	appTxStatusInit     int32 = 1 // decoded from bytes 2 tx
	appTxStatusChecking int32 = 2 // validating by one routine
	appTxStatusChecked  int32 = 3 // validated
	appTxStatusFailed   int32 = 4 // tx is invalid
)

type appTx struct {
	rawbytes agtypes.Tx
	tx       *types.BlockTx
	ready    sync.WaitGroup
	status   int32
	err      error
}

func exeWithCPUParallelVeirfy(txs [][]byte, quit chan struct{},
	whenExec func(index int, raw []byte, tx *types.BlockTx), whenError func(bs []byte, err error)) error {
	var exit int32
	go func() {
		if quit == nil {
			return
		}
		select {
		case <-quit:
			atomic.StoreInt32(&exit, 1)
		}
	}()

	appTxQ := make([]appTx, len(txs))
	go makeTxQueue(txs, appTxQ, &exit)

	if validateRoutineCount < 1 || validateRoutineCount > 16 {
		validateRoutineCount = 8
	}
	for i := 0; i < validateRoutineCount; i++ {
		go validateRoutine(appTxQ, &exit)
	}

	size := len(txs)
	for i := 0; i < size; i++ {
	INNERFOR:
		for {
			q := atomic.LoadInt32(&exit)
			if q == 1 {
				return errQuitExecute
			}

			status := atomic.LoadInt32(&appTxQ[i].status)
			switch status {
			case appTxStatusChecked:
				whenExec(i, appTxQ[i].rawbytes, appTxQ[i].tx)
				break INNERFOR
			case appTxStatusFailed:
				whenError(appTxQ[i].rawbytes, appTxQ[i].err)
				break INNERFOR
			default:
				appTxQ[i].ready.Wait()
			}
		}
	}

	return nil
}

func makeTxQueue(txs [][]byte, apptxQ []appTx, exit *int32) {
	for i, raw := range txs {
		q := atomic.LoadInt32(exit)
		if q == 1 {
			return
		}

		apptxQ[i].rawbytes = raw
		apptxQ[i].ready.Add(1)

		// decode bytes
		if bytes.HasPrefix(raw, types.TxTagApp) {
			apptxQ[i].tx = new(types.BlockTx)
			if err := rlp.DecodeBytes(agtypes.UnwrapTx(raw), apptxQ[i].tx); err != nil {
				apptxQ[i].err = err
			}
		} else {
			// non-app tx will pass directly
			atomic.StoreInt32(&apptxQ[i].status, appTxStatusChecked)
		}

		atomic.StoreInt32(&apptxQ[i].status, appTxStatusInit)
	}
}

func validateRoutine(appTxQ []appTx, exit *int32) {
	size := len(appTxQ)
	for i := 0; i < size; i++ {
	INNERFOR:
		for {
			q := atomic.LoadInt32(exit)
			if q == 1 {
				return
			}

			status := atomic.LoadInt32(&appTxQ[i].status)
			switch status {
			case appTxStatusNone: // we can do nothing but waiting
				time.Sleep(time.Microsecond)
			case appTxStatusInit: // try validating
				tryValidate(&appTxQ[i])
			default: // move to next
				break INNERFOR
			}
		}
	}
}

func tryValidate(tx *appTx) {
	swapped := atomic.CompareAndSwapInt32(&tx.status, appTxStatusInit, appTxStatusChecking)
	if !swapped {
		return
	}

	defer tx.ready.Done()

	//  when this tx exists errors
	if tx.err != nil {
		atomic.StoreInt32(&tx.status, appTxStatusFailed)
		return
	}

	// when this tx is not a evm-like tx
	if tx.tx == nil {
		atomic.StoreInt32(&tx.status, appTxStatusChecked)
		return
	}

	valid, err := tx.tx.VerifySignature()
	if err != nil {
		atomic.StoreInt32(&tx.status, appTxStatusFailed)
		tx.err = err
		return
	}
	if !valid {
		atomic.StoreInt32(&tx.status, appTxStatusFailed)
		tx.err = fmt.Errorf("tx verify failed")
		return
	}

	atomic.StoreInt32(&tx.status, appTxStatusChecked)
	return
}

func exeWithCPUSerialVeirfy(txs [][]byte, quit chan struct{},
	whenExec func(index int, raw []byte, tx *types.BlockTx), whenError func(raw []byte, err error)) error {
	for i, raw := range txs {
		tx := &types.BlockTx{}
		if bytes.HasPrefix(raw, types.TxTagApp) {
			if err := rlp.DecodeBytes(agtypes.UnwrapTx(raw), tx); err != nil {
				whenError(raw, err)
				continue
			}

			valid, err := tx.VerifySignature()
			if err != nil {
				whenError(raw, err)
				continue
			}
			if !valid {
				whenError(raw, fmt.Errorf("tx verify failed"))
			}
		}

		whenExec(i, raw, tx)
	}

	return nil
}
