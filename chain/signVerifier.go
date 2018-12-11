package chain

import (
	"errors"
	"github.com/aergoio/aergo/account/key"
	"github.com/aergoio/aergo/internal/enc"
	"github.com/aergoio/aergo/message"
	"github.com/aergoio/aergo/pkg/component"
	"github.com/aergoio/aergo/types"
	"math"
	"time"
)

type SignVerifier struct {
	comm component.IComponentRequester

	workerCnt int
	workChs   []chan verifyWork
	doneCh    chan *VerifyResult

	checkMempool bool

	hitMempool int
}

type verifyWork struct {
	idx int
	txs []*types.Tx
}

type VerifyResult struct {
	work    *verifyWork
	errors  []error
	hit     int
	success bool
	elapsed time.Duration
}

var (
	ErrTxFormatInvalid = errors.New("tx invalid format")
	dfltUseMempool     = false
	//logger = log.NewLogger("signverifier")
)

func NewSignVerifier(comm component.IComponentRequester, workerCnt int, checkMempool bool) *SignVerifier {
	workChs := make([]chan verifyWork, workerCnt)
	for i, _ := range workChs {
		workChs[i] = make(chan verifyWork, 3)
	}

	sv := &SignVerifier{
		comm:         comm,
		workerCnt:    workerCnt,
		workChs:      workChs,
		doneCh:       make(chan *VerifyResult, workerCnt),
		checkMempool: checkMempool,
	}

	for i := 0; i < workerCnt; i++ {
		go sv.verifyTxLoop(i, sv.workChs[i])
	}

	return sv
}

func (sv *SignVerifier) Stop() {
	for _, workC := range sv.workChs {
		close(workC)
	}
	close(sv.doneCh)
}

func (sv *SignVerifier) verifyTxLoop(workerNo int, workC chan verifyWork) {
	logger.Debug().Int("worker", workerNo).Msg("verify worker run")

	for txWork := range workC {
		errs := make([]error, len(txWork.txs))
		start := time.Now()

		//logger.Debug().Int("worker", workerNo).Int("from", txWork.idx).Int("len", len(txWork.txs)).Msg("get work to verify tx")
		totalHit := 0
		success := true

		for i, tx := range txWork.txs {
			hit, err := sv.verifyTx(sv.comm, tx)
			if err != nil {
				logger.Error().Int("worker", workerNo).Int("idx", txWork.idx+i).Str("hash", enc.ToString(tx.GetHash())).
					Err(err).Msg("error verify tx")
				success = false
			}

			if hit {
				totalHit++
			}
			errs[i] = err
		}

		t := time.Now()
		elapsed := t.Sub(start)

		sv.doneCh <- &VerifyResult{work: &txWork, errors: errs, hit: totalHit, success: success, elapsed: elapsed}
	}

	logger.Debug().Int("worker", workerNo).Msg("verify worker stop")

}

func (sv *SignVerifier) verifyTx(comm component.IComponentRequester, tx *types.Tx) (hit bool, err error) {
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
		return true, nil
	}

	errors := make([]error, txLen, txLen)

	//logger.Debug().Int("txlen", txLen).Msg("verify tx start")

	//split txlists to worker count
	workSize := int(math.Ceil(float64(txLen) / float64(sv.workerCnt)))
	workCount := int(math.Ceil(float64(txLen) / float64(workSize)))
	logger.Debug().Int("worksize", workSize).Int("workcount", workCount).Int("txlen", txLen).Msg("push tx start")

	go func() {
		from := 0
		to := 0
		for i := 0; i < sv.workerCnt && from < txLen; i++ {
			if from+workSize <= txLen {
				to = from + workSize
			} else {
				to = txLen
			}

			logger.Debug().Int("from", from).Int("to", to).Msg("push tx start")
			sv.workChs[i] <- verifyWork{idx: i, txs: txs[from:to]}

			from = to
		}
	}()

	var doneCnt = 0
	success := true
	sv.hitMempool = 0
LOOP:
	for {
		select {
		case result := <-sv.doneCh:
			doneCnt++
			//logger.Debug().Int("donecnt", doneCnt).Msg("verify tx done")

			if result.work.idx < 0 || len(result.work.txs) == 0 || len(result.work.txs) != len(result.errors) {
				logger.Error().Int("idx", result.work.idx).Msg("Invalid Verify Result Index")
				success = false
				break LOOP
			}

			copy(errors[result.work.idx:len(result.work.txs)], result.errors)

			if !result.success {
				success = false
			}

			sv.hitMempool += result.hit

			logger.Debug().Int("worker", result.work.idx).Int64("elap", result.elapsed.Nanoseconds()).Msg("worker done")
			if doneCnt == workCount {
				break LOOP
			}
		}
	}

	logger.Debug().Int("mempool", sv.hitMempool).Msg("verify tx done")
	return success, errors
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
