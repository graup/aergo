/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package p2p

// MsgReader read stream and return message object
type MsgReader interface {
	// ReadMsg return types.MsgHeader as header, proto.Message as data
	// The header and/or data can be nil if error is not nil
	ReadMsg() (Message, error)
}

// MsgWriter write message to stream
type MsgWriter interface {
	WriteMsg(msg Message) error
}

type MsgReadWriter interface {
	MsgReader
	MsgWriter
}
