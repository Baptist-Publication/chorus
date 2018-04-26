// Copyright 2017 Baptist-Publication Information Technology Services Co.,Ltd.
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

package p2p

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"math"
	"net"
	"runtime/debug"
	"strings"
	"sync/atomic"
	"time"

	"github.com/spf13/viper"
	"go.uber.org/zap"

	. "github.com/Baptist-Publication/chorus/module/lib/go-common"
	flow "github.com/Baptist-Publication/chorus/module/lib/go-flowrate/flowrate"
	"github.com/Baptist-Publication/chorus/module/xlib"
)

const (
	numBatchMsgPackets = 10
	minReadBufferSize  = 1024
	minWriteBufferSize = 65536
	idleTimeoutMinutes = 5
	updateStatsSeconds = 2
	pingTimeoutSeconds = 40
	flushThrottleMS    = 100

	defaultSendQueueCapacity   = 1
	defaultRecvBufferCapacity  = 4096
	defaultRecvMessageCapacity = 22020096 // 21MB
	defaultSendTimeoutSeconds  = 10
)

type receiveCbFunc func(chID byte, msgBytes []byte)
type errorCbFunc func(interface{})

/*
Each peer has one `MConnection` (multiplex connection) instance.

__multiplex__ *noun* a system or signal involving simultaneous transmission of
several messages along a single channel of communication.

Each `MConnection` handles message transmission on multiple abstract communication
`Channel`s.  Each channel has a globally unique byte id.
The byte id and the relative priorities of each `Channel` are configured upon
initialization of the connection.

There are two methods for sending messages:
	func (m MConnection) Send(chID byte, msg interface{}) bool {}
	func (m MConnection) TrySend(chID byte, msg interface{}) bool {}

`Send(chID, msg)` is a blocking call that waits until `msg` is successfully queued
for the channel with the given id byte `chID`, or until the request times out.
The message `msg` is serialized using the `tendermint/wire` submodule's
`WriteBinary()` reflection routine.

`TrySend(chID, msg)` is a nonblocking call that returns false if the channel's
queue is full.

Inbound message bytes are handled with an onReceive callback function.
*/
type MConnection struct {
	BaseService

	conn        net.Conn
	bufReader   *bufio.Reader
	bufWriter   *bufio.Writer
	sendMonitor *flow.Monitor
	recvMonitor *flow.Monitor
	sendRate    int64
	recvRate    int64
	send        chan struct{}
	pong        chan struct{}
	channels    []*Channel
	channelsIdx map[byte]*Channel
	onReceive   receiveCbFunc
	onError     errorCbFunc
	errored     uint32

	quit         chan struct{}
	flushTimer   *ThrottleTimer // flush writes as necessary but throttled.
	pingTimer    *RepeatTimer   // send pings periodically
	chStatsTimer *RepeatTimer   // update channel stats periodically

	LocalAddress  *NetAddress
	RemoteAddress *NetAddress

	connectResetWait time.Duration

	logger  *zap.Logger
	slogger *zap.SugaredLogger
}

func NewMConnection(logger *zap.Logger, config *viper.Viper, conn net.Conn, chDescs []*ChannelDescriptor, onReceive receiveCbFunc, onError errorCbFunc) *MConnection {
	mconn := &MConnection{
		conn:        conn,
		bufReader:   bufio.NewReaderSize(conn, minReadBufferSize),
		bufWriter:   bufio.NewWriterSize(conn, minWriteBufferSize),
		sendMonitor: flow.New(0, 0),
		recvMonitor: flow.New(0, 0),
		sendRate:    config.GetInt64(configKeySendRate),
		recvRate:    config.GetInt64(configKeyRecvRate),
		send:        make(chan struct{}, 1),
		pong:        make(chan struct{}),
		onReceive:   onReceive,
		onError:     onError,

		// Initialized in Start()
		quit:         nil,
		flushTimer:   nil,
		pingTimer:    nil,
		chStatsTimer: nil,

		LocalAddress:  NewNetAddress(logger, conn.LocalAddr()),
		RemoteAddress: NewNetAddress(logger, conn.RemoteAddr()),

		connectResetWait: time.Duration(config.GetInt("connection_reset_wait")) * time.Millisecond,

		logger:  logger,
		slogger: logger.Sugar(),
	}

	// Create channels
	var channelsIdx = map[byte]*Channel{}
	var channels = []*Channel{}

	for _, desc := range chDescs {
		descCopy := *desc // copy the desc else unsafe access across connections
		channel := newChannel(mconn, &descCopy)
		channelsIdx[channel.id] = channel
		channels = append(channels, channel)
	}
	mconn.channels = channels
	mconn.channelsIdx = channelsIdx

	mconn.BaseService = *NewBaseService(logger, "MConnection", mconn)

	return mconn
}

func (c *MConnection) OnStart() error {
	c.BaseService.OnStart()
	c.quit = make(chan struct{})
	c.flushTimer = NewThrottleTimer("flush", flushThrottleMS*time.Millisecond)
	c.pingTimer = NewRepeatTimer("ping", pingTimeoutSeconds*time.Second)
	c.chStatsTimer = NewRepeatTimer("chStats", updateStatsSeconds*time.Second)
	go c.sendRoutine()
	go c.recvRoutine()
	return nil
}

func (c *MConnection) OnStop() {
	c.BaseService.OnStop()
	c.flushTimer.Stop()
	c.pingTimer.Stop()
	c.chStatsTimer.Stop()
	if c.quit != nil {
		close(c.quit)
	}
	c.conn.Close()
	// We can't close pong safely here because
	// recvRoutine may write to it after we've stopped.
	// Though it doesn't need to get closed at all,
	// we close it @ recvRoutine.
	// close(c.pong)
}

func (c *MConnection) String() string {
	return fmt.Sprintf("MConn{%v}", c.conn.RemoteAddr())
}

func (c *MConnection) flush() {
	err := c.bufWriter.Flush()
	if err != nil {
		c.logger.Warn("MConnection flush failed", zap.Error(err))
	}
}

// Catch panics, usually caused by remote disconnects.
func (c *MConnection) _recover() {
	if r := recover(); r != nil {
		stack := debug.Stack()
		c.stopForError(StackError{r, stack})
	}
}

func (c *MConnection) stopForError(r interface{}) {
	c.Stop()
	if atomic.CompareAndSwapUint32(&c.errored, 0, 1) {
		if c.onError != nil {
			c.onError(r)
		}
	}
}

// Queues a message to be sent to channel.
// func (c *MConnection) Send(chID byte, msg interface{}) bool {
func (c *MConnection) Send(chID byte, msg []byte) bool {
	if !c.IsRunning() {
		return false
	}

	// Send message to channel.
	channel, ok := c.channelsIdx[chID]
	if !ok {
		c.logger.Error("Cannot send bytes, unknown channel", zap.ByteString("chID", []byte{chID}))
		return false
	}

	success := channel.sendBytes(msg)
	if success {
		// Wake up sendRoutine if necessary
		select {
		case c.send <- struct{}{}:
		default:
		}
	} else {
		c.slogger.Warnw("Send failed", "channel", chID, "conn", c, "msg", msg)
	}
	return success
}

// Queues a message to be sent to channel.
// Nonblocking, returns true if successful.
func (c *MConnection) TrySend(chID byte, msg []byte) bool {
	if !c.IsRunning() {
		return false
	}

	// Send message to channel.
	channel, ok := c.channelsIdx[chID]
	if !ok {
		c.logger.Error("Cannot send bytes, unknown channels", zap.ByteString("chID", []byte{chID}))
		return false
	}

	ok = channel.trySendBytes(msg)
	if ok {
		// Wake up sendRoutine if necessary
		select {
		case c.send <- struct{}{}:
		default:
		}
	}

	return ok
}

func (c *MConnection) CanSend(chID byte) bool {
	if !c.IsRunning() {
		return false
	}

	channel, ok := c.channelsIdx[chID]
	if !ok {
		c.slogger.Errorf("Unknown channel %X", chID)
		return false
	}
	return channel.canSend()
}

// sendRoutine polls for packets to send from channels.
func (c *MConnection) sendRoutine() {
	defer c._recover()

FOR_LOOP:
	for {
		var err error
		select {
		case <-c.flushTimer.Ch:
			// NOTE: flushTimer.Set() must be called every time
			// something is written to .bufWriter.
			c.flush()
		case <-c.chStatsTimer.Ch:
			for _, channel := range c.channels {
				channel.updateStats()
			}
		case <-c.pingTimer.Ch:
			//c.logger.Debug("Send Ping")
			xlib.BinWrite(c.bufWriter, packetTypePing)
			c.sendMonitor.Update(1)
			c.flush()
		case <-c.pong:
			//c.logger.Debug("Send Pong")
			xlib.BinWrite(c.bufWriter, packetTypePong)
			c.sendMonitor.Update(1)
			c.flush()
		case <-c.quit:
			break FOR_LOOP
		case <-c.send:
			// Send some msgPackets
			eof := c.sendSomeMsgPackets()
			if !eof {
				// Keep sendRoutine awake.
				select {
				case c.send <- struct{}{}:
				default:
				}
			}
		}

		if !c.IsRunning() {
			break FOR_LOOP
		}
		if err != nil {
			// the only way to identify "connection reset by peer" is to read the error humanly,
			// which is against go error handling principle

			if strings.Contains(err.Error(), "connection reset by peer") {
				//c.logger.Debug("Sleeping " + c.connectResetWait.String() + " because of 'connection reset by peer'")
				time.Sleep(c.connectResetWait)
			} else {
				c.slogger.Warnw("Connection failed @ sendRoutine", "conn", c, "error", err)
				c.stopForError(err)
				break FOR_LOOP
			}
		}
	}

	// Cleanup
}

// Returns true if messages from channels were exhausted.
// Blocks in accordance to .sendMonitor throttling.
func (c *MConnection) sendSomeMsgPackets() bool {
	// Block until .sendMonitor says we can write.
	// Once we're ready we send more than we asked for,
	// but amortized it should even out.
	c.sendMonitor.Limit(maxMsgPacketTotalSize, atomic.LoadInt64(&c.sendRate), true)

	// Now send some msgPackets.
	for i := 0; i < numBatchMsgPackets; i++ {
		if c.sendMsgPacket() {
			return true
		}
	}
	return false
}

// Returns true if messages from channels were exhausted.
func (c *MConnection) sendMsgPacket() bool {
	// Choose a channel to create a msgPacket from.
	// The chosen channel will be the one whose recentlySent/priority is the least.
	var leastRatio float32 = math.MaxFloat32
	var leastChannel *Channel
	for _, channel := range c.channels {
		// If nothing to send, skip this channel
		if !channel.isSendPending() {
			continue
		}
		// Get ratio, and keep track of lowest ratio.
		ratio := float32(channel.recentlySent) / float32(channel.priority)
		if ratio < leastRatio {
			leastRatio = ratio
			leastChannel = channel
		}
	}

	// Nothing to send?
	if leastChannel == nil {
		return true
	}

	// Make & send a msgPacket from this channel
	n, err := leastChannel.writeMsgPacketTo(c.bufWriter)
	if err != nil {
		c.logger.Warn("Failed to write msgPacket", zap.Error(err))
		c.stopForError(err)
		return true
	}
	c.sendMonitor.Update(int(n))
	c.flushTimer.Set()
	return false
}

// recvRoutine reads msgPackets and reconstructs the message using the channels' "recving" buffer.
// After a whole message has been assembled, it's pushed to onReceive().
// Blocks depending on how the connection is throttled.
func (c *MConnection) recvRoutine() {
	defer c._recover()

FOR_LOOP:
	for {
		// Block until .recvMonitor says we can read.
		c.recvMonitor.Limit(maxMsgPacketTotalSize, atomic.LoadInt64(&c.recvRate), true)

		/*
			// Peek into bufReader for debugging
			if numBytes := c.bufReader.Buffered(); numBytes > 0 {
				log.Info("Peek connection buffer", "numBytes", numBytes, "bytes", log15.Lazy{func() []byte {
					bytes, err := c.bufReader.Peek(MinInt(numBytes, 100))
					if err == nil {
						return bytes
					} else {
						log.Warn("Error peeking connection buffer", "error", err)
						return nil
					}
				}})
			}
		*/

		// Read packet type
		var err error
		var pktType byte
		err = xlib.BinRead(c.bufReader, &pktType)
		c.recvMonitor.Update(1)
		if err != nil {
			if c.IsRunning() {
				c.slogger.Warnw("Connection failed @ recvRoutine (reading byte)", "conn", c, "error", err)
				c.stopForError(err)
			}
			break FOR_LOOP
		}

		// Read more depending on packet type.
		switch pktType {
		case packetTypePing:
			// TODO: prevent abuse, as they cause flush()'s.
			//c.logger.Debug("Receive Ping")
			c.pong <- struct{}{}
		case packetTypePong:
			// do nothing
			//c.logger.Debug("Receive Pong")
		case packetTypeMsg:
			pkt := &msgPacket{}
			n, err := pkt.ReadBytes(c.bufReader)
			c.recvMonitor.Update(n)
			if err != nil {
				if c.IsRunning() {
					c.slogger.Warnw("Connection failed @ recvRoutine", "conn", c, "error", err)
					c.stopForError(err)
				}
				break FOR_LOOP
			}
			channel, ok := c.channelsIdx[pkt.ChannelID]
			if !ok || channel == nil {
				c.logger.Error("Unknown channel", zap.ByteString("ChannelID", []byte{pkt.ChannelID}))
				continue FOR_LOOP
			}
			msgBytes, err := channel.recvMsgPacket(pkt)
			if err != nil {
				if c.IsRunning() {
					c.slogger.Warnw("Connection failed @ recvRoutine", "conn", c, "error", err)
					c.stopForError(err)
				}
				break FOR_LOOP
			}
			if msgBytes != nil {
				//c.logger.Debug("Received bytes", zap.Binary("chID", []byte{pkt.ChannelID}), zap.Binary("msgBytes", msgBytes))
				c.onReceive(pkt.ChannelID, msgBytes)
			}
		default:
			c.logger.Error("Unknown message type", zap.ByteString("message type", []byte{pktType}))
		}

		// TODO: shouldn't this go in the sendRoutine?
		// Better to send a ping packet when *we* haven't sent anything for a while.
		c.pingTimer.Reset()
	}

	// Cleanup
	close(c.pong)
	for _ = range c.pong {
		// Drain
	}
}

type ConnectionStatus struct {
	SendMonitor flow.Status
	RecvMonitor flow.Status
	Channels    []ChannelStatus
}

type ChannelStatus struct {
	ID                byte
	SendQueueCapacity int
	SendQueueSize     int
	Priority          int
	RecentlySent      int64
}

func (c *MConnection) Status() ConnectionStatus {
	var status ConnectionStatus
	status.SendMonitor = c.sendMonitor.Status()
	status.RecvMonitor = c.recvMonitor.Status()
	status.Channels = make([]ChannelStatus, len(c.channels))
	for i, channel := range c.channels {
		status.Channels[i] = ChannelStatus{
			ID:                channel.id,
			SendQueueCapacity: cap(channel.sendQueue),
			SendQueueSize:     int(channel.sendQueueSize), // TODO use atomic
			Priority:          channel.priority,
			RecentlySent:      channel.recentlySent,
		}
	}
	return status
}

//-----------------------------------------------------------------------------

type ChannelDescriptor struct {
	ID                  byte
	Priority            int
	SendQueueCapacity   int
	RecvBufferCapacity  int
	RecvMessageCapacity int
}

func (chDesc *ChannelDescriptor) FillDefaults() {
	if chDesc.SendQueueCapacity == 0 {
		chDesc.SendQueueCapacity = defaultSendQueueCapacity
	}
	if chDesc.RecvBufferCapacity == 0 {
		chDesc.RecvBufferCapacity = defaultRecvBufferCapacity
	}
	if chDesc.RecvMessageCapacity == 0 {
		chDesc.RecvMessageCapacity = defaultRecvMessageCapacity
	}
}

// TODO: lowercase.
// NOTE: not goroutine-safe.
type Channel struct {
	conn          *MConnection
	desc          *ChannelDescriptor
	id            byte
	sendQueue     chan []byte
	sendQueueSize int32 // atomic.
	recving       []byte
	sending       []byte
	priority      int
	recentlySent  int64 // exponential moving average

	slogger *zap.SugaredLogger
}

func newChannel(conn *MConnection, desc *ChannelDescriptor) *Channel {
	desc.FillDefaults()
	if desc.Priority <= 0 {
		PanicSanity("Channel default priority must be a postive integer")
	}
	return &Channel{
		conn:      conn,
		desc:      desc,
		id:        desc.ID,
		sendQueue: make(chan []byte, desc.SendQueueCapacity),
		recving:   make([]byte, 0, desc.RecvBufferCapacity),
		priority:  desc.Priority,
	}
}

// Queues message to send to this channel.
// Goroutine-safe
// Times out (and returns false) after defaultSendTimeoutSeconds
func (ch *Channel) sendBytes(bytes []byte) bool {
	timeout := time.NewTimer(defaultSendTimeoutSeconds * time.Second)
	select {
	case <-timeout.C:
		// timeout
		return false
	case ch.sendQueue <- bytes:
		atomic.AddInt32(&ch.sendQueueSize, 1)
		return true
	}
}

// Queues message to send to this channel.
// Nonblocking, returns true if successful.
// Goroutine-safe
func (ch *Channel) trySendBytes(bytes []byte) bool {
	select {
	case ch.sendQueue <- bytes:
		atomic.AddInt32(&ch.sendQueueSize, 1)
		return true
	default:
		return false
	}
}

// Goroutine-safe
func (ch *Channel) loadSendQueueSize() (size int) {
	return int(atomic.LoadInt32(&ch.sendQueueSize))
}

// Goroutine-safe
// Use only as a heuristic.
func (ch *Channel) canSend() bool {
	return ch.loadSendQueueSize() < defaultSendQueueCapacity
}

// Returns true if any msgPackets are pending to be sent.
// Call before calling nextMsgPacket()
// Goroutine-safe
func (ch *Channel) isSendPending() bool {
	if len(ch.sending) == 0 {
		if len(ch.sendQueue) == 0 {
			return false
		}
		ch.sending = <-ch.sendQueue
	}
	return true
}

// Creates a new msgPacket to send.
// Not goroutine-safe
func (ch *Channel) nextMsgPacket() msgPacket {
	packet := msgPacket{}
	packet.ChannelID = byte(ch.id)
	packet.Bytes = ch.sending[:MinInt(maxMsgPacketPayloadSize, len(ch.sending))]
	if len(ch.sending) <= maxMsgPacketPayloadSize {
		packet.EOF = byte(0x01)
		ch.sending = nil
		atomic.AddInt32(&ch.sendQueueSize, -1) // decrement sendQueueSize
	} else {
		packet.EOF = byte(0x00)
		ch.sending = ch.sending[MinInt(maxMsgPacketPayloadSize, len(ch.sending)):]
	}
	return packet
}

// Writes next msgPacket to w.
// Not goroutine-safe
func (ch *Channel) writeMsgPacketTo(w io.Writer) (n int, err error) {
	packet := ch.nextMsgPacket()
	//ch.conn.slogger.Debugw("Write Msg Packet", "conn", ch.conn, "packet", packet)
	xlib.BinWrite(w, packetTypeMsg)
	n, err = packet.WriteBytes(w)
	if err == nil {
		ch.recentlySent += int64(n)
	}
	return
}

var ErrBinaryReadOverflow = errors.New("Error:bin binary read overflow")

// Handles incoming msgPackets. Returns a msg bytes if msg is complete.
// Not goroutine-safe
func (ch *Channel) recvMsgPacket(packet *msgPacket) ([]byte, error) {
	// log.Debug("Read Msg Packet", "conn", ch.conn, "packet", packet)
	if ch.desc.RecvMessageCapacity < len(ch.recving)+len(packet.Bytes) {
		return nil, ErrBinaryReadOverflow
	}
	ch.recving = append(ch.recving, packet.Bytes...)
	if packet.EOF == byte(0x01) {
		msgBytes := ch.recving
		// clear the slice without re-allocating.
		// http://stackoverflow.com/questions/16971741/how-do-you-clear-a-slice-in-go
		//   suggests this could be a memory leak, but we might as well keep the memory for the channel until it closes,
		//	at which point the recving slice stops being used and should be garbage collected
		ch.recving = ch.recving[:0] // make([]byte, 0, ch.desc.RecvBufferCapacity)
		return msgBytes, nil
	}
	return nil, nil
}

// Call this periodically to update stats for throttling purposes.
// Not goroutine-safe
func (ch *Channel) updateStats() {
	// Exponential decay of stats.
	// TODO: optimize.
	ch.recentlySent = int64(float64(ch.recentlySent) * 0.8)
}

//-----------------------------------------------------------------------------

const (
	maxMsgPacketPayloadSize  = 1024
	maxMsgPacketOverheadSize = 10 // It's actually lower but good enough
	maxMsgPacketTotalSize    = maxMsgPacketPayloadSize + maxMsgPacketOverheadSize
	packetTypePing           = byte(0x01)
	packetTypePong           = byte(0x02)
	packetTypeMsg            = byte(0x03)
)

// Messages in channels are chopped into smaller msgPackets for multiplexing.
type msgPacket struct {
	ChannelID byte
	EOF       byte // 1 means message ends here.
	Bytes     []byte
}

func (mp *msgPacket) WriteBytes(writer io.Writer) (n int, err error) {
	if err = xlib.BinWrite(writer, mp.ChannelID); err != nil {
		return
	}
	n += 1
	if err = xlib.BinWrite(writer, mp.EOF); err != nil {
		return
	}
	n += 1
	if err = xlib.WriteBytes(writer, mp.Bytes); err != nil {
		return
	}
	n += len(mp.Bytes)
	return
}

func (mp *msgPacket) ReadBytes(reader io.Reader) (n int, err error) {
	if err = xlib.BinRead(reader, &mp.ChannelID); err != nil {
		return
	}
	n += 1
	if err = xlib.BinRead(reader, &mp.EOF); err != nil {
		return
	}
	n += 1
	if mp.Bytes, err = xlib.ReadBytes(reader); err != nil {
		return
	}
	n += len(mp.Bytes)
	return
}

func (p msgPacket) String() string {
	return fmt.Sprintf("MsgPacket{%X:%X T:%X}", p.ChannelID, p.Bytes, p.EOF)
}
