package chain

import (
	"errors"
	"github.com/aergoio/aergo/account/key"
	"github.com/aergoio/aergo/internal/enc"
	"github.com/aergoio/aergo/message"
	"github.com/aergoio/aergo/pkg/component"
	"github.com/aergoio/aergo/types"
	"time"
)

type SignVerifier struct {
	comm component.IComponentRequester

	workerCnt int
	workCh    chan verifyWork
	doneCh    chan VerifyResult

	checkMempool bool

	hitMempool int
}

type verifyWork struct {
	idx int
	tx  *types.Tx
}

type VerifyResult struct {
	work *verifyWork
	err  error
	hit  bool
}

var (
	ErrTxFormatInvalid = errors.New("tx invalid format")
	dfltUseMempool     = true
	//logger = log.NewLogger("signverifier")
)

func NewSignVerifier(comm component.IComponentRequester, workerCnt int, checkMempool bool) *SignVerifier {
	sv := &SignVerifier{
		comm:         comm,
		workerCnt:    workerCnt,
		workCh:       make(chan verifyWork, workerCnt),
		doneCh:       make(chan VerifyResult, workerCnt),
		checkMempool: checkMempool,
	}

	for i := 0; i < workerCnt; i++ {
		go sv.verifyTxLoop(i)
	}

	return sv
}

func (sv *SignVerifier) Stop() {
	close(sv.workCh)
	close(sv.doneCh)
}

func (sv *SignVerifier) verifyTxLoop(workerNo int) {
	logger.Debug().Int("worker", workerNo).Msg("verify worker run")

	for txWork := range sv.workCh {
		//logger.Debug().Int("worker", workerNo).Int("idx", txWork.idx).Msg("get work to verify tx")
		hit, err := sv.verifyTx(sv.comm, txWork.tx)

		if err != nil {
			logger.Error().Int("worker", workerNo).Bool("hitmempoo", hit).Str("hash", enc.ToString(txWork.tx.GetHash())).
				Err(err).Msg("error verify tx")
		}

		sv.doneCh <- VerifyResult{work: &txWork, err: err, hit: hit}
	}

	logger.Debug().Int("worker", workerNo).Msg("verify worker stop")

}

func (sv *SignVerifier) verifyTx(comm component.IComponentRequester, tx *types.Tx) (hitMempool bool, err error) {
	if sv.checkMempool {
		result, err := comm.RequestToFutureResult(message.MemPoolSvc, &message.MemPoolExist{Hash: tx.GetHash()}, time.Second,
			"chain/signverifier/verifytx")
		if err != nil {
			logger.Error().Err(err).Msg("failed to get verify from mempool")
			return false, err
		}

		msg := result.(*message.MemPoolExistRsp)
		if msg.Tx != nil {
			return true, nil
		}
	}

	account := tx.GetBody().GetAccount()
	if account == nil {
		return false, ErrTxFormatInvalid
	}

	err = key.VerifyTx(tx)

	if err != nil {
		return false, err
	}

	return false, nil
}

func (sv *SignVerifier) VerifyTxs(txlist *types.TxList) (bool, []error) {
	txs := txlist.GetTxs()
	txLen := len(txs)

	if txLen == 0 {
		return false, nil
	}

	errors := make([]error, txLen, txLen)

	//logger.Debug().Int("txlen", txLen).Msg("verify tx start")

	go func() {
		for i, tx := range txs {
			//logger.Debug().Int("idx", i).Msg("push tx start")
			sv.workCh <- verifyWork{idx: i, tx: tx}
		}
	}()

	var doneCnt = 0
	failed := false
	sv.hitMempool = 0

LOOP:
	for {
		select {
		case result := <-sv.doneCh:
			doneCnt++
			//logger.Debug().Int("donecnt", doneCnt).Msg("verify tx done")

			if result.work.idx < 0 || result.work.idx >= txLen {
				logger.Error().Int("idx", result.work.idx).Msg("Invalid Verify Result Index")
				continue
			}

			errors[result.work.idx] = result.err

			if result.err != nil {
				logger.Error().Err(result.err).Int("txno", result.work.idx).
					Msg("verifing tx failed")
				failed = true
			} else {
				sv.hitMempool++
			}

			if doneCnt == txLen {
				break LOOP
			}
		}
	}

	logger.Debug().Int("mempool", sv.hitMempool).Msg("verify tx done")
	return failed, errors
}

func (bv *SignVerifier) verifyTxsInplace(txlist *types.TxList) (bool, []error) {
	txs := txlist.GetTxs()
	txLen := len(txs)
	errors := make([]error, txLen, txLen)
	failed := false
	var hit bool

	logger.Debug().Int("txlen", txLen).Msg("verify tx inplace start")

	for i, tx := range txs {
		hit, errors[i] = bv.verifyTx(bv.comm, tx)
		failed = true

		if hit {
			bv.hitMempool++
		}
	}

	logger.Debug().Int("mempool", bv.hitMempool).Msg("verify tx inplace done")
	return failed, errors
}
