// Copyright 2022 The kubegems.io Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package tunnel

import (
	"errors"
	"io"
	"net"
	"time"

	"golang.org/x/sync/errgroup"
	"kubegems.io/kubegems/pkg/log"
)

var ErrFullChannel = errors.New("channel full")

type TunnelConn struct {
	c *Connections

	channel *ConnectedTunnel
	rawConn net.Conn

	local              string
	localConnectionID  int64
	remote             string
	remoteConnectionID int64

	rdata  []byte
	closed bool
	ack    chan *connectData
}

func (c *TunnelConn) recv(remotecid int64, data []byte, err string) error {
	select {
	case c.ack <- &connectData{remoteID: remotecid, err: err, data: data}:
		return nil
	default:
		log.Error(ErrFullChannel, "drop packet",
			"cid", c.localConnectionID,
			"remote", c.channel.ID,
			"remote cid", c.remoteConnectionID,
		)
		return ErrFullChannel
	}
}

func (c *TunnelConn) opened(remotecid int64) {
	c.remoteConnectionID = remotecid
}

func (c *TunnelConn) accepted(conn net.Conn) {
	c.rawConn = conn
	eg := errgroup.Group{}
	eg.Go(func() error {
		_, err := io.Copy(c.rawConn, c)
		return err
	})
	eg.Go(func() error {
		_, err := io.Copy(c, c.rawConn)
		return err
	})
	eg.Wait()
}

type connectData struct {
	remoteID int64
	err      string
	data     []byte
}

func (c *TunnelConn) Read(b []byte) (n int, err error) {
	var data []byte
	if c.rdata != nil {
		data = c.rdata
	} else {
		ack := <-c.ack
		if ack == nil /*closed*/ {
			return 0, net.ErrClosed
		}
		if ack.err == "EOF" {
			return 0, io.EOF
		}
		if ack.err != "" {
			return 0, errors.New(ack.err)
		}
		data = ack.data
	}
	if data == nil {
		return 0, io.EOF
	}
	if len(data) > len(b) {
		copy(b, data[:len(b)])
		c.rdata = data[len(b):]
		return len(b), nil
	}
	c.rdata = nil
	copy(b, data)
	return len(data), nil
}

func (c *TunnelConn) Write(b []byte) (n int, err error) {
	if c.closed {
		return 0, net.ErrClosed
	}
	if err = c.sendData(b); err != nil {
		return 0, err
	}
	return len(b), err
}

// Close tunnel connection and close raw connection,remove self from connection manager
func (c *TunnelConn) Close() error {
	return c.sendClose(c.close())
}

func (c *TunnelConn) close() error {
	return c.c.close(c.remoteConnectionID)
}

func (c *TunnelConn) LocalAddr() net.Addr {
	return nil
}

func (c *TunnelConn) RemoteAddr() net.Addr {
	return nil
}

func (c *TunnelConn) SetDeadline(t time.Time) error {
	return nil
}

func (c *TunnelConn) SetReadDeadline(t time.Time) error {
	return nil
}

func (c *TunnelConn) SetWriteDeadline(t time.Time) error {
	return nil
}

func (c *TunnelConn) sendData(data []byte) error {
	return c.sendPkt(func(pkt *Packet) {
		pkt.Kind = PacketKindData
		pkt.Data = data
	})
}

func (c *TunnelConn) sendClose(err error) error {
	return c.sendPkt(func(pkt *Packet) {
		pkt.Kind = PacketKindClose
		if err != nil {
			pkt.Error = err.Error()
		}
	})
}

func (c *TunnelConn) sendOpen(data PacketDataOpen) error {
	return c.sendPkt(func(pkt *Packet) {
		pkt.Kind = PacketKindOpen
		pkt.Data = PacketEncode(data)
	})
}

func (c *TunnelConn) sendPkt(fun func(pkt *Packet)) error {
	pkt := &Packet{
		Kind:    PacketKindData,
		Src:     c.local,
		SrcCID:  c.localConnectionID,
		Dest:    c.remote,
		DestCID: c.remoteConnectionID,
	}
	fun(pkt)
	return c.channel.Send(pkt)
}
