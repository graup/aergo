/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package p2p

import (
	"bufio"
	"sync"

	inet "github.com/libp2p/go-libp2p-net"

	"github.com/aergoio/aergo/pkg/log"
	"github.com/aergoio/aergo/types"
	"github.com/multiformats/go-multicodec/protobuf"
)

// pattern: /protocol-name/request-or-response-message/version
const (
	pingRequest   = "/ping/pingreq/0.1"
	pingResponse  = "/ping/pingresp/0.1"
	statusRequest = "/ping/status/0.1"
	goAway        = "/ping/goaway/0.1"
)

// PingProtocol type
type PingProtocol struct {
	actorServ ActorService
	ps        PeerManager

	log      log.ILogger
	reqMutex sync.Mutex
}

// NewPingProtocol create ping subprotocol
func NewPingProtocol(logger log.ILogger) *PingProtocol {
	p := &PingProtocol{log: logger,
		reqMutex: sync.Mutex{}}
	return p
}

func (p *PingProtocol) initWith(p2pservice PeerManager) {
	p.ps = p2pservice
	p.ps.SetStreamHandler(pingRequest, p.onPingRequest)
	p.ps.SetStreamHandler(pingResponse, p.onPingResponse)
	p.ps.SetStreamHandler(statusRequest, p.onStatusRequest)
}

// remote peer requests handler
func (p *PingProtocol) onPingRequest(s inet.Stream) {
	peerID := s.Conn().RemotePeer()
	remotePeer, ok := p.ps.GetPeer(peerID)
	if !ok {
		warnLogUnknownPeer(p.log, s.Protocol(), peerID)
		return
	}

	remotePeer.readLock.Lock()
	defer remotePeer.readLock.Unlock()

	// get request data
	data := &types.Ping{}
	decoder := mc_pb.Multicodec(nil).Decoder(bufio.NewReader(s))
	err := decoder.Decode(data)
	if err != nil {
		p.log.Warnf("Failed to decode ping request. %s", err.Error())
		return
	}
	debugLogReceiveMsg(p.log, s.Protocol(), data.MessageData.Id, peerID, nil)

	// generate response message
	p.log.Debugf("Sending pong to %s. MSG ID %s.", s.Conn().RemotePeer().Pretty(), data.MessageData.Id)
	resp := &types.Pong{MessageData: &types.MessageData{}} // BestBlockHash: bestBlock.GetHash(),
	// BestHeight:    bestBlock.GetHeader().GetBlockNo(),

	remotePeer.sendMessage(newPbMsgResponseOrder(data.MessageData.Id, false, pingResponse, resp))
}

// remote ping response handler
func (p *PingProtocol) onPingResponse(s inet.Stream) {
	peerID := s.Conn().RemotePeer()
	remotePeer, ok := p.ps.GetPeer(peerID)
	if !ok {
		warnLogUnknownPeer(p.log, s.Protocol(), peerID)
		return
	}

	remotePeer.readLock.Lock()
	defer remotePeer.readLock.Unlock()

	data := &types.Pong{}
	decoder := mc_pb.Multicodec(nil).Decoder(bufio.NewReader(s))
	err := decoder.Decode(data)
	if err != nil {
		p.log.Warnf("Failed to decode pong request. %s", err.Error())
		return
	}
	debugLogReceiveMsg(p.log, s.Protocol(), data.MessageData.Id, peerID, nil)
	remotePeer.consumeRequest(data.MessageData.Id)
}

func (p *PingProtocol) onStatusRequest(s inet.Stream) {
	peerID := s.Conn().RemotePeer()
	remotePeer, ok := p.ps.LookupPeer(peerID)
	if !ok {
		warnLogUnknownPeer(p.log, s.Protocol(), peerID)
		return
	}

	remotePeer.readLock.Lock()
	defer remotePeer.readLock.Unlock()

	// get request data
	data := &types.Status{}
	decoder := mc_pb.Multicodec(nil).Decoder(bufio.NewReader(s))
	err := decoder.Decode(data)
	if err != nil {
		p.log.Warnf("Failed to decode ping request. %s", err.Error())
		return
	}
	debugLogReceiveMsg(p.log, s.Protocol(), data.MessageData.Id, peerID, nil)
	valid := p.ps.AuthenticateMessage(data, data.MessageData)
	if !valid {
		p.log.Warn("Failed to authenticate message")
		return
	}

	peerState := remotePeer.State()
	if peerState != types.STARTING && peerState != types.HANDSHAKING {
		remotePeer.goAwayMsg("Invalid status msg")
		return
	}

	remotePeer.updateMetaInfo(data)
	// TODO: check protocol version, blacklist, key authentication or etc.
	err = remotePeer.checkProtocolVersion()
	if err != nil {
		p.log.Infof("invalid protocol version of peer %v", remotePeer.meta.ID.Pretty())
		remotePeer.goAwayMsg("Handshake failed")
		return
	}

	// if state is han
	if peerState == types.STARTING {
		remotePeer.sendStatus()
	}
	p.log.Debug("starting handshake ")
	remotePeer.handshakePeer(data)
}
