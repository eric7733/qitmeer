// Copyright (c) 2017-2018 The qitmeer developers
// Copyright (c) 2013-2016 The btcsuite developers
// Copyright (c) 2015-2018 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.
package peerserver

import (
	"errors"
	"github.com/satori/go.uuid"
)

type getConnCountMsg struct {
	reply chan int32
}

type getPeersMsg struct {
	reply chan []*serverPeer
}

type getOutboundGroup struct {
	key   string
	reply chan int
}

type getAddedNodesMsg struct {
	reply chan []*serverPeer
}

type disconnectNodeMsg struct {
	cmp   func(*serverPeer) bool
	reply chan error
}

type connectNodeMsg struct {
	addr      string
	permanent bool
	reply     chan error
}

type removeNodeMsg struct {
	cmp   func(*serverPeer) bool
	reply chan error
}

type getPeerMsg struct {
	uuid  uuid.UUID
	reply chan bool
}

// handleQuery is the central handler for all queries and commands from other
// goroutines related to peer state.
func (s *PeerServer) handleQuery(state *peerState, querymsg interface{}) {
	switch msg := querymsg.(type) {
	case getConnCountMsg:
		nconnected := int32(0)
		state.forAllPeers(func(sp *serverPeer) {
			if sp.Connected() {
				nconnected++
			}
		})
		msg.reply <- nconnected

	case getPeersMsg:
		peers := make([]*serverPeer, 0, state.Count())
		state.forAllPeers(func(sp *serverPeer) {
			if !sp.Connected() {
				return
			}
			peers = append(peers, sp)
		})
		msg.reply <- peers

	case connectNodeMsg:
		msg.reply <- errors.New("not support")
	case removeNodeMsg:
		msg.reply <- errors.New("not support")
	case getOutboundGroup:
		count, ok := state.outboundGroups[msg.key]
		if ok {
			msg.reply <- count
		} else {
			msg.reply <- 0
		}
	case getAddedNodesMsg:
		peers := make([]*serverPeer, 0)
		msg.reply <- peers
	case disconnectNodeMsg:
		msg.reply <- errors.New("not support")
		
	case getPeerMsg:
		has:=false
		state.forAllPeers(func(sp *serverPeer) {
			if !sp.Connected() || has {
				return
			}
			if uuid.Equal(sp.UUID(),msg.uuid) {
				has=true
				return
			}
		})
		msg.reply <- has
	}
}
