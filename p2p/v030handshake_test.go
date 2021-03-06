/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package p2p

import (
	"fmt"
	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/types"
	"github.com/stretchr/testify/mock"
	"reflect"
	"testing"
)

func TestV030StatusHS_handshakeOutboundPeer(t *testing.T) {
	logger = log.NewLogger("test")
	mockActor := new(MockActorService)
	mockCA := new(MockChainAccessor)
	mockPM := new(MockPeerManager)

	dummyMeta := PeerMeta{ID: dummyPeerID}
	mockPM.On("SelfMeta").Return(dummyMeta)
	dummyBlock := &types.Block{Hash: dummyBlockHash, Header: &types.BlockHeader{BlockNo: dummyBlockHeight}}
	mockActor.On("GetChainAccessor").Return(mockCA)
	mockCA.On("GetBestBlock").Return(dummyBlock, nil)

	dummyStatusMsg := &types.Status{}
	statusBytes, _ := MarshalMessage(dummyStatusMsg)
	tests := []struct {
		name       string
		readReturn *types.Status
		readError  error
		writeError error
		want       *types.Status
		wantErr    bool
	}{
		{"TSuccess", dummyStatusMsg, nil, nil, dummyStatusMsg, false},
		{"TUnexpMsg", nil, nil, nil, nil, true},
		{"TRFail", dummyStatusMsg, fmt.Errorf("failed"), nil, nil, true},
		{"TWFail", dummyStatusMsg, nil, fmt.Errorf("failed"), nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dummyReader := new(MockReader)
			dummyWriter := new(MockWriter)
			mockRW := new(MockMsgReadWriter)

			containerMsg := &V030Message{payload:statusBytes}
			if tt.readReturn != nil {
				containerMsg.subProtocol = StatusRequest
			} else {
				containerMsg.subProtocol = AddressesRequest
			}
			mockRW.On("ReadMsg").Return(containerMsg, tt.readError)
			mockRW.On("WriteMsg", mock.Anything).Return(tt.writeError)

			h := newV030StateHS(mockPM, mockActor, logger, samplePeerID, dummyReader, dummyWriter)
			h.msgRW = mockRW
			got, err := h.doForOutbound()
			if (err != nil) != tt.wantErr {
				t.Errorf("PeerHandshaker.handshakeOutboundPeer() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PeerHandshaker.handshakeOutboundPeer() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestV030StatusHS_handshakeInboundPeer(t *testing.T) {
	// t.SkipNow()
	logger = log.NewLogger("test")
	mockActor := new(MockActorService)
	mockCA := new(MockChainAccessor)
	mockPM := new(MockPeerManager)

	dummyMeta := PeerMeta{ID: dummyPeerID}
	mockPM.On("SelfMeta").Return(dummyMeta)
	dummyBlock := &types.Block{Hash: dummyBlockHash, Header: &types.BlockHeader{BlockNo: dummyBlockHeight}}
	//dummyBlkRsp := message.GetBestBlockRsp{Block: dummyBlock}
	mockActor.On("GetChainAccessor").Return(mockCA)
	mockCA.On("GetBestBlock").Return(dummyBlock, nil)

	dummyStatusMsg := &types.Status{}
	statusBytes, _ := MarshalMessage(dummyStatusMsg)
	tests := []struct {
		name       string
		readReturn *types.Status
		readError  error
		writeError error
		want       *types.Status
		wantErr    bool
	}{
		{"TSuccess", dummyStatusMsg, nil, nil, dummyStatusMsg, false},
		{"TUnexpMsg", nil, nil, nil, nil, true},
		{"TRFail", dummyStatusMsg, fmt.Errorf("failed"), nil, nil, true},
		{"TWFail", dummyStatusMsg, nil, fmt.Errorf("failed"), nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dummyReader := new(MockReader)
			dummyWriter := new(MockWriter)
			mockRW := new(MockMsgReadWriter)
			containerMsg := &V030Message{payload:statusBytes}
			if tt.readReturn != nil {
				containerMsg.subProtocol = StatusRequest
			} else {
				containerMsg.subProtocol = AddressesRequest
			}

			mockRW.On("ReadMsg").Return(containerMsg, tt.readError)
			mockRW.On("WriteMsg", mock.Anything).Return(tt.writeError)

			h := newV030StateHS(mockPM, mockActor, logger, samplePeerID, dummyReader, dummyWriter)
			h.msgRW = mockRW
			got, err := h.doForInbound()
			if (err != nil) != tt.wantErr {
				t.Errorf("PeerHandshaker.handshakeOutboundPeer() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PeerHandshaker.handshakeOutboundPeer() = %v, want %v", got, tt.want)
			}
		})
	}
}
