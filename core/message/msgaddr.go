// Copyright (c) 2017-2018 The qitmeer developers
// Copyright (c) 2013-2015 The btcsuite developers
// Copyright (c) 2015-2016 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package message

import (
	"fmt"
	"github.com/Qitmeer/qitmeer-lib/core/types"
	s "github.com/Qitmeer/qitmeer/core/serialization"
	"io"
)

// MaxAddrPerMsg is the maximum number of addresses that can be in a single
// bitcoin addr message (MsgAddr).
const MaxAddrPerMsg = 1000

// MsgAddr implements the Message interface and represents a bitcoin
// addr message.  It is used to provide a list of known active peers on the
// network.  An active peer is considered one that has transmitted a message
// within the last 3 hours.  Nodes which have not transmitted in that time
// frame should be forgotten.  Each message is limited to a maximum number of
// addresses, which is currently 1000.  As a result, multiple messages must
// be used to relay the full list.
//
// Use the AddAddress function to build up the list of known addresses when
// sending an addr message to another peer.
type MsgAddr struct {
	AddrList []*types.NetAddress
}

// AddAddress adds a known active peer to the message.
func (msg *MsgAddr) AddAddress(na *types.NetAddress) error {
	if len(msg.AddrList)+1 > MaxAddrPerMsg {
		str := fmt.Sprintf("too many addresses in message [max %v]",
			MaxAddrPerMsg)
		return messageError("MsgAddr.AddAddress", str)
	}

	msg.AddrList = append(msg.AddrList, na)
	return nil
}

// AddAddresses adds multiple known active peers to the message.
func (msg *MsgAddr) AddAddresses(netAddrs ...*types.NetAddress) error {
	for _, na := range netAddrs {
		err := msg.AddAddress(na)
		if err != nil {
			return err
		}
	}
	return nil
}

// ClearAddresses removes all addresses from the message.
func (msg *MsgAddr) ClearAddresses() {
	msg.AddrList = []*types.NetAddress{}
}

// Decode decodes r into the receiver.
// This is part of the Message interface implementation.
func (msg *MsgAddr) Decode(r io.Reader, pver uint32) error {
	count, err := s.ReadVarInt(r, pver)
	if err != nil {
		return err
	}

	// Limit to max addresses per message.
	if count > MaxAddrPerMsg {
		str := fmt.Sprintf("too many addresses for message "+
			"[count %v, max %v]", count, MaxAddrPerMsg)
		return messageError("MsgAddr.Decode", str)
	}

	addrList := make([]types.NetAddress, count)
	msg.AddrList = make([]*types.NetAddress, 0, count)
	for i := uint64(0); i < count; i++ {
		na := &addrList[i]
		err := types.ReadNetAddress(r, pver, na, true)
		if err != nil {
			return err
		}
		msg.AddAddress(na)
	}
	return nil
}

// ncode encodes the receiver to w.
// This is part of the Message interface implementation.
func (msg *MsgAddr) Encode(w io.Writer, pver uint32) error {
	// Protocol versions before MultipleAddressVersion only allowed 1 address
	// per message.
	count := len(msg.AddrList)
	if count > MaxAddrPerMsg {
		str := fmt.Sprintf("too many addresses for message "+
			"[count %v, max %v]", count, MaxAddrPerMsg)
		return messageError("MsgAddr.BtcEncode", str)
	}

	err := s.WriteVarInt(w, pver, uint64(count))
	if err != nil {
		return err
	}

	for _, na := range msg.AddrList {
		err = types.WriteNetAddress(w, pver, na, true)
		if err != nil {
			return err
		}
	}

	return nil
}

// Command returns the protocol command string for the message.  This is part
// of the Message interface implementation.
func (msg *MsgAddr) Command() string {
	return CmdAddr
}

// MaxPayloadLength returns the maximum length the payload can be for the
// receiver.  This is part of the Message interface implementation.
func (msg *MsgAddr) MaxPayloadLength(pver uint32) uint32 {
	// Num addresses (varInt) + max allowed addresses.
	return s.MaxVarIntPayload + (MaxAddrPerMsg * types.MaxNetAddressPayload(pver))
}

// NewMsgAddr returns a new bitcoin addr message that conforms to the
// Message interface.  See MsgAddr for details.
func NewMsgAddr() *MsgAddr {
	return &MsgAddr{
		AddrList: make([]*types.NetAddress, 0, MaxAddrPerMsg),
	}
}
