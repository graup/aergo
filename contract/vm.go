/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package contract

/*
#cgo CFLAGS: -I${SRCDIR}/../libtool/include/luajit-2.0
#cgo LDFLAGS: -g ${SRCDIR}/../libtool/lib/libluajit-5.1.a -lm

#include <stdlib.h>
#include <string.h>
#include "vm.h"
*/
import "C"
import (
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"sync"

	"github.com/aergoio/aergo-lib/db"
	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/types"
	"unsafe"

	"github.com/aergoio/aergo/state"
)

const DbName = "contracts.db"
const constructorName = "constructor"

var (
	ctrLog      *log.Logger
	DB          db.DB
	contractMap stateMap
)

type Contract struct {
	code    []byte
	address []byte
}

type CallState struct {
	ctrState  *state.ContractState
	prevState *types.State
	curState  *types.State
}

type StateSet struct {
	contract  *state.ContractState
	bs        *types.BlockState
	sdb       *state.ChainStateDB
	callState map[string]*CallState
}

type stateMap struct {
	states map[string]*StateSet
	mu     sync.Mutex
}

type LState = C.struct_lua_State
type LBlockchainCtx = C.struct_blockchain_ctx

type Executor struct {
	L             *LState
	contract      *Contract
	err           error
	blockchainCtx *LBlockchainCtx
	jsonRet       string
}

func init() {
	ctrLog = log.NewLogger("contract")
	contractMap.init()
}

func NewContext(sdb *state.ChainStateDB, blockState *types.BlockState, contractState *state.ContractState, Sender, txHash []byte, blockHeight uint64,
	timestamp int64, node string, confirmed bool, contractID []byte, query bool) *LBlockchainCtx {

	var iConfirmed, isQuery int
	if confirmed {
		iConfirmed = 1
	}
	if query {
		isQuery = 1
	}
	enContractId := types.EncodeAddress(contractID)
	enTxHash := hex.EncodeToString(txHash)

	stateKey := fmt.Sprintf("%s%s", enContractId, enTxHash)
	contractMap.register(stateKey, &StateSet{contract: contractState, bs: blockState, sdb: sdb})

	return &LBlockchainCtx{
		stateKey:    C.CString(stateKey),
		sender:      C.CString(types.EncodeAddress(Sender)),
		txHash:      C.CString(enTxHash),
		blockHeight: C.ulonglong(blockHeight),
		timestamp:   C.longlong(timestamp),
		node:        C.CString(node),
		confirmed:   C.int(iConfirmed),
		contractId:  C.CString(enContractId),
		isQuery:     C.int(isQuery),
	}
}

func newLState() *LState {
	return C.vm_newstate()
}

func (L *LState) Close() {
	if L != nil {
		C.lua_close(L)
	}
}

func newExecutor(contract *Contract, bcCtx *LBlockchainCtx) *Executor {
	address := C.CString(types.EncodeAddress(contract.address))
	defer C.free(unsafe.Pointer(address))

	ce := &Executor{
		contract: contract,
		L:        newLState(),
	}
	if ce.L == nil {
		ctrLog.Error().Str("error", "Failed: create lua state")
		ce.err = errors.New("Failed: create lua state")
		return ce
	}
	if cErrMsg := C.vm_loadbuff(
		ce.L,
		(*C.char)(unsafe.Pointer(&contract.code[0])),
		C.size_t(len(contract.code)),
		address,
		bcCtx,
	); cErrMsg != nil {
		errMsg := C.GoString(cErrMsg)
		C.free(unsafe.Pointer(cErrMsg))
		ctrLog.Error().Str("error", errMsg)
		ce.err = errors.New(errMsg)
	}
	return ce
}

func (ce *Executor) processArgs(ci *types.CallInfo) {
	for _, v := range ci.Args {
		switch arg := v.(type) {
		case string:
			argC := C.CString(arg)
			C.lua_pushstring(ce.L, argC)
			C.free(unsafe.Pointer(argC))
		case int:
			C.lua_pushinteger(ce.L, C.lua_Integer(arg))
		case float64:
			C.lua_pushnumber(ce.L, C.double(arg))
		case bool:
			var b int
			if arg {
				b = 1
			}
			C.lua_pushboolean(ce.L, C.int(b))
		default:
			fmt.Println("unsupported type:" + reflect.TypeOf(v).Name())
			ce.err = errors.New("unsupported type:" + reflect.TypeOf(v).Name())
			return
		}
	}
}

func (ce *Executor) call(ci *types.CallInfo, target *LState) {
	if ce.err != nil {
		return
	}
	abiStr := C.CString("abi")
	callStr := C.CString("call")
	abiName := C.CString(ci.Name)

	defer C.free(unsafe.Pointer(abiStr))
	defer C.free(unsafe.Pointer(callStr))
	defer C.free(unsafe.Pointer(abiName))

	C.vm_getfield(ce.L, abiStr)
	C.lua_getfield(ce.L, -1, callStr)
	C.lua_pushstring(ce.L, abiName)

	ce.processArgs(ci)
	nret := C.int(0)
	if cErrMsg := C.vm_pcall(ce.L, C.int(len(ci.Args)+1), &nret); cErrMsg != nil {
		errMsg := C.GoString(cErrMsg)
		C.free(unsafe.Pointer(cErrMsg))
		ctrLog.Warn().Str("error", errMsg).Msgf("contract %s", types.EncodeAddress(ce.contract.address))
		ce.err = errors.New(errMsg)
		return
	}

	if target == nil {
		ce.jsonRet = C.GoString(C.vm_get_json_ret(ce.L, nret))
	} else {
		C.vm_copy_result(ce.L, target, nret)
	}
}

func (ce *Executor) constructCall(ci *types.CallInfo) {
	if ce.err != nil {
		return
	}
	initName := C.CString(constructorName)
	defer C.free(unsafe.Pointer(initName))

	C.vm_getfield(ce.L, initName)
	ce.processArgs(ci)

	if ce.err != nil {
		return
	}
	nret := C.int(0)
	if cErrMsg := C.vm_pcall(ce.L, C.int(len(ci.Args)), &nret); cErrMsg != nil {
		errMsg := C.GoString(cErrMsg)
		C.free(unsafe.Pointer(cErrMsg))
		ctrLog.Warn().Str("error", errMsg).Msgf("contract %s", types.EncodeAddress(ce.contract.address))
		ce.err = errors.New(errMsg)
		return
	}

	ce.jsonRet = C.GoString(C.vm_get_json_ret(ce.L, nret))
}

func (ce *Executor) commitCalledContract() error {
	if ce.blockchainCtx == nil {
		return nil
	}
	stateKey := C.GoString(ce.blockchainCtx.stateKey)
	stateSet := contractMap.lookup(stateKey)

	if stateSet == nil || stateSet.callState == nil {
		return nil
	}

	sdb := stateSet.sdb
	bs := stateSet.bs

	var err error
	for k, v := range stateSet.callState {
		err = sdb.CommitContractState(v.ctrState)
		if err != nil {
			return err
		}
		aid, _ := types.DecodeAddress(k)
		bs.PutAccount(types.ToAccountID(aid), v.prevState, v.curState)
	}
	return nil
}

func (ce *Executor) close() {
	if ce != nil {
		ce.L.Close()
		if ce.blockchainCtx != nil {
			context := ce.blockchainCtx
			contractMap.unregister(C.GoString(context.stateKey))

			C.free(unsafe.Pointer(context.sender))
			C.free(unsafe.Pointer(context.txHash))
			C.free(unsafe.Pointer(context.node))
			C.free(unsafe.Pointer(context.stateKey))
			C.free(unsafe.Pointer(context))
		}
	}
}

func Call(contractState *state.ContractState, code, contractAddress, txHash []byte, bcCtx *LBlockchainCtx, dbTx db.Transaction) error {
	var err error
	var ci types.CallInfo
	contract := getContract(contractState, contractAddress, nil)
	if contract != nil {
		err = json.Unmarshal(code, &ci)
		if err != nil {
			ctrLog.Warn().AnErr("error", err).Msgf("contract %s", types.EncodeAddress(contractAddress))
		}
	} else {
		err = fmt.Errorf("cannot find contract %s", string(contractAddress))
		ctrLog.Warn().AnErr("err", err)
	}
	var ce *Executor
	if err == nil {
		ctrLog.Debug().Str("abi", string(code)).Msgf("contract %s", types.EncodeAddress(contractAddress))
		ce = newExecutor(contract, bcCtx)
		defer ce.close()
		ce.call(&ci, nil)
		err = ce.err
		if err == nil {
			err = ce.commitCalledContract()
		}
	}
	var receipt types.Receipt
	if err == nil {
		receipt = types.NewReceipt(contractAddress, "SUCCESS", ce.jsonRet)
	} else {
		receipt = types.NewReceipt(contractAddress, err.Error(), "")
	}
	dbTx.Set(txHash, receipt.Bytes())
	return err
}

func Create(contractState *state.ContractState, code, contractAddress, txHash []byte, bcCtx *LBlockchainCtx, dbTx db.Transaction) error {
	ctrLog.Debug().Str("contractAddress", types.EncodeAddress(contractAddress)).Msg("new contract is deployed")
	codeLen := binary.LittleEndian.Uint32(code[0:])
	sCode := code[4:codeLen]

	err := contractState.SetCode(sCode)
	if err != nil {
		return err
	}
	contract := getContract(contractState, contractAddress, sCode)
	if contract == nil {
		err = fmt.Errorf("cannot deploy contract %s", types.EncodeAddress(contractAddress))
		ctrLog.Warn().AnErr("err", err)
		return err
	}
	var ce *Executor
	ctrLog.Debug().Str("deploy code", string(code)).Msgf("contract deploy %s", types.EncodeAddress(contractAddress))
	ce = newExecutor(contract, bcCtx)
	defer ce.close()

	var ci types.CallInfo
	if len(code) != int(codeLen) {
		err = json.Unmarshal(code[codeLen:], &ci.Args)
	}
	var errMsg string
	if err != nil {
		errMsg = err.Error()
	}
	ce.constructCall(&ci)
	if ce.err != nil {
		errMsg += "\", \"constructor call error:" + ce.err.Error()
	}
	receipt := types.NewReceipt(contractAddress, "CREATED", "{\""+errMsg+"\"}")
	dbTx.Set(txHash, receipt.Bytes())

	return nil
}

func Query(contractAddress []byte, contractState *state.ContractState, queryInfo []byte) ([]byte, error) {
	var ci types.CallInfo
	var err error
	contract := getContract(contractState, contractAddress, nil)
	if contract != nil {
		err = json.Unmarshal(queryInfo, &ci)
		if err != nil {
			ctrLog.Warn().AnErr("error", err).Msgf("contract %s", types.EncodeAddress(contractAddress))
		}
	} else {
		err = fmt.Errorf("cannot find contract %s", types.EncodeAddress(contractAddress))
		ctrLog.Warn().AnErr("err", err)
	}
	if err != nil {
		return nil, err
	}
	var ce *Executor

	bcCtx := NewContext(nil, nil, contractState, contractAddress, nil,
		0, 0, "", false, contractAddress, true)
	ctrLog.Debug().Str("abi", string(queryInfo)).Msgf("contract %s", types.EncodeAddress(contractAddress))
	ce = newExecutor(contract, bcCtx)
	defer ce.close()
	ce.call(&ci, nil)
	err = ce.err

	return []byte(ce.jsonRet), err
}

func getContract(contractState *state.ContractState, contractAddress []byte, code []byte) *Contract {
	var val []byte
	val = code
	if val == nil {
		var err error
		val, err = contractState.GetCode()

		if err != nil {
			return nil
		}
	}
	if len(val) > 0 {
		l := binary.LittleEndian.Uint32(val[0:])
		return &Contract{
			code:    val[4 : 4+l],
			address: contractAddress[:],
		}
	}
	return nil
}

func GetReceipt(txHash []byte) (*types.Receipt, error) {
	val := DB.Get(txHash)
	if len(val) == 0 {
		return nil, errors.New("cannot find a receipt")
	}
	return types.NewReceiptFromBytes(val), nil
}

func GetABI(contractState *state.ContractState, contractAddress []byte) (*types.ABI, error) {
	val, err := contractState.GetCode()
	if err != nil {
		return nil, err
	}
	if len(val) == 0 {
		return nil, errors.New("cannot find contract")
	}
	l := binary.LittleEndian.Uint32(val[0:])
	abi := new(types.ABI)
	if err := json.Unmarshal(val[4+l:], abi); err != nil {
		return nil, err
	}
	return abi, nil
}

func (sm *stateMap) init() {
	sm.states = make(map[string]*StateSet)
}

func (sm *stateMap) register(key string, item *StateSet) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if sm.states[key] != nil {
		err := fmt.Errorf("already exists contract state: %s", key)
		ctrLog.Warn().AnErr("err", err)
	}
	sm.states[key] = item
}

func (sm *stateMap) unregister(key string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if sm.states[key] == nil {
		err := fmt.Errorf("cannot find contract state: %s", key)
		ctrLog.Warn().AnErr("err", err)
	}

	delete(sm.states, key)
}

func (sm *stateMap) lookup(key string) *StateSet {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	return sm.states[key]
}

func luaPushStr(L *LState, str string) {
	cStr := C.CString(str)
	C.lua_pushstring(L, cStr)
	C.free(unsafe.Pointer(cStr))
}

//export LuaSetDB
func LuaSetDB(L *LState, stateKey *C.char, key *C.char, value *C.char) C.int {
	stateKeyString := C.GoString(stateKey)
	keyString := C.GoString(key)
	valueString := C.GoString(value)

	stateSet := contractMap.lookup(stateKeyString)
	if stateSet == nil {
		luaPushStr(L, "[System.LuaSetDB]not found contract state")
		return -1
	}

	err := stateSet.contract.SetData([]byte(keyString), []byte(valueString))
	if err != nil {
		luaPushStr(L, err.Error())
		return -1
	}
	return 0
}

//export LuaGetDB
func LuaGetDB(L *LState, stateKey *C.char, key *C.char) C.int {
	stateKeyString := C.GoString(stateKey)
	keyString := C.GoString(key)

	stateSet := contractMap.lookup(stateKeyString)
	if stateSet == nil {
		luaPushStr(L, "[System.LuaGetDB]not found contract state")
		return -1
	}

	data, err := stateSet.contract.GetData([]byte(keyString))
	if err != nil {
		luaPushStr(L, err.Error())
		return -1
	}

	if data == nil {
		return 0
	}
	luaPushStr(L, string(data))
	return 1
}

//export LuaCallContract
func LuaCallContract(L *LState, bcCtx *LBlockchainCtx, contractId *C.char, fname *C.char, args *C.char) C.int {
	stateKeyStr := C.GoString(bcCtx.stateKey)
	contractIdStr := C.GoString(contractId)
	fnameStr := C.GoString(fname)
	argsStr := C.GoString(args)

	cid, err := types.DecodeAddress(contractIdStr)
	if err != nil {
		luaPushStr(L, "[System.LuaGetContract]invalid contractId :"+err.Error())
		return -1
	}

	stateSet := contractMap.lookup(stateKeyStr)
	if stateSet == nil {
		luaPushStr(L, "[System.LuaCallContract]not found contract state")
		return -1
	}

	if stateSet.callState == nil {
		stateSet.callState = make(map[string]*CallState)
	}
	callState := stateSet.callState[contractIdStr]
	if callState == nil {
		sdb := stateSet.sdb
		bs := stateSet.bs

		prevState, err := sdb.GetBlockAccountClone(bs, types.ToAccountID(cid))
		if err != nil {
			luaPushStr(L, "[System.LuaGetContract]getAccount Error :"+err.Error())
			return -1
		}

		curState := types.Clone(*prevState).(types.State)
		contractState, err := sdb.OpenContractState(&curState)
		if err != nil {
			luaPushStr(L, "[System.LuaGetContract]getAccount Error :"+err.Error())
			return -1
		}
		callState =
			&CallState{ctrState: contractState, prevState: prevState, curState: &curState}
		stateSet.callState[contractIdStr] = callState
	}
	contract := getContract(callState.ctrState, cid, nil)
	if contract == nil {
		luaPushStr(L, "[System.LuaGetContract]cannot find contract "+string(contractIdStr))
		return -1
	}
	newBcCtx := NewContext(nil, nil, callState.ctrState,
		C.GoBytes(unsafe.Pointer(bcCtx.sender), C.int(C.strlen(bcCtx.sender))),
		[]byte(""), uint64(bcCtx.blockHeight), int64(bcCtx.timestamp), "", false, cid, false)
	ce := newExecutor(contract, newBcCtx)
	defer ce.close()

	if ce.err != nil {
		luaPushStr(L, "[System.LuaGetContract]newExecutor Error :"+ce.err.Error())
		return -1
	}

	var ci types.CallInfo
	ci.Name = fnameStr
	err = json.Unmarshal([]byte(argsStr), &ci.Args)
	if err != nil {
		luaPushStr(L, "[System.LuaCallContract] invalid args:"+err.Error())
		return -1
	}
	ce.call(&ci, L)
	if ce.err != nil {
		luaPushStr(L, "[System.LuaCallContract] call err:"+ce.err.Error())
		return -1
	}

	return 0
}
