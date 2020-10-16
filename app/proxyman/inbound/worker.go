package inbound

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"v2ray.com/core/app/proxyman"
	"v2ray.com/core/common"
	"v2ray.com/core/common/buf"
	"v2ray.com/core/common/net"
	"v2ray.com/core/common/serial"
	"v2ray.com/core/common/session"
	"v2ray.com/core/common/signal/done"
	"v2ray.com/core/common/task"
	"v2ray.com/core/features/routing"
	"v2ray.com/core/features/stats"
	"v2ray.com/core/proxy"
	"v2ray.com/core/transport/internet"
	"v2ray.com/core/transport/internet/tcp"
	"v2ray.com/core/transport/internet/udp"
	"v2ray.com/core/transport/pipe"
)

type worker interface {
	Start() error
	Close() error
	Port() net.Port
	Proxy() proxy.Inbound
}

type tcpWorker struct {
	address         net.Address
	port            net.Port
	proxy           proxy.Inbound
	stream          *internet.MemoryStreamConfig
	recvOrigDest    bool
	tag             string
	dispatcher      routing.Dispatcher
	sniffingConfig  *proxyman.SniffingConfig
	uplinkCounter   stats.Counter
	downlinkCounter stats.Counter

	hub internet.Listener

	ctx context.Context
}

func getTProxyType(s *internet.MemoryStreamConfig) internet.SocketConfig_TProxyMode {
	if s == nil || s.SocketSettings == nil {
		return internet.SocketConfig_Off
	}
	return s.SocketSettings.Tproxy
}

func (w *tcpWorker) callback(conn internet.Connection) {
	ctx, cancel := context.WithCancel(w.ctx)
	sid := session.NewID()
	ctx = session.ContextWithID(ctx, sid)

	if w.recvOrigDest {
		var dest net.Destination
		switch getTProxyType(w.stream) {
		case internet.SocketConfig_Redirect:
			d, err := tcp.GetOriginalDestination(conn)
			if err != nil {
				newError("failed to get original destination").Base(err).WriteToLog(session.ExportIDToError(ctx))
			} else {
				dest = d
			}
		case internet.SocketConfig_TProxy:
			dest = net.DestinationFromAddr(conn.LocalAddr())
		}
		if dest.IsValid() {
			ctx = session.ContextWithOutbound(ctx, &session.Outbound{
				Target: dest,
			})
		}
	}
	ctx = session.ContextWithInbound(ctx, &session.Inbound{
		Source:  net.DestinationFromAddr(conn.RemoteAddr()),
		Gateway: net.TCPDestination(w.address, w.port),
		Tag:     w.tag,
	})
	content := new(session.Content)
	if w.sniffingConfig != nil {
		content.SniffingRequest.Enabled = w.sniffingConfig.Enabled
		content.SniffingRequest.OverrideDestinationForProtocol = w.sniffingConfig.DestinationOverride
	}
	ctx = session.ContextWithContent(ctx, content)
	if w.uplinkCounter != nil || w.downlinkCounter != nil {
		conn = &internet.StatCouterConnection{
			Connection:   conn,
			ReadCounter:  w.uplinkCounter,
			WriteCounter: w.downlinkCounter,
		}
	}
	if err := w.proxy.Process(ctx, net.Network_TCP, conn, w.dispatcher); err != nil {
		newError("connection ends").Base(err).WriteToLog(session.ExportIDToError(ctx))
	}
	cancel()
	if err := conn.Close(); err != nil {
		newError("failed to close connection").Base(err).WriteToLog(session.ExportIDToError(ctx))
	}
}

func (w *tcpWorker) Proxy() proxy.Inbound {
	return w.proxy
}

func (w *tcpWorker) Start() error {
	ctx := context.Background()
	hub, err := internet.ListenTCP(ctx, w.address, w.port, w.stream, func(conn internet.Connection) {
		go w.callback(conn)
	})
	if err != nil {
		return newError("failed to listen TCP on ", w.port).AtWarning().Base(err)
	}
	w.hub = hub
	return nil
}

func (w *tcpWorker) Close() error {
	var errors []interface{}
	if w.hub != nil {
		if err := common.Close(w.hub); err != nil {
			errors = append(errors, err)
		}
		if err := common.Close(w.proxy); err != nil {
			errors = append(errors, err)
		}
	}
	if len(errors) > 0 {
		return newError("failed to close all resources").Base(newError(serial.Concat(errors...)))
	}

	return nil
}

func (w *tcpWorker) Port() net.Port {
	return w.port
}

type udpConn struct {
	lastActivityTime int64 // in seconds
	reader           buf.Reader
	writer           buf.Writer
	output           func([]byte) (int, error)
	remote           net.Addr
	local            net.Addr
	done             *done.Instance
	uplink           stats.Counter
	downlink         stats.Counter
}

func (c *udpConn) updateActivity() {
	atomic.StoreInt64(&c.lastActivityTime, time.Now().Unix())
}

// ReadMultiBuffer implements buf.Reader
func (c *udpConn) ReadMultiBuffer() (buf.MultiBuffer, error) {
	mb, err := c.reader.ReadMultiBuffer()
	if err != nil {
		return nil, err
	}
	c.updateActivity()

	if c.uplink != nil {
		c.uplink.Add(int64(mb.Len()))
	}

	return mb, nil
}

func (c *udpConn) Read(buf []byte) (int, error) {
	panic("not implemented")
}

// Write implements io.Writer.
func (c *udpConn) Write(buf []byte) (int, error) {
	n, err := c.output(buf)
	if c.downlink != nil {
		c.downlink.Add(int64(n))
	}
	if err == nil {
		c.updateActivity()
	}
	return n, err
}

func (c *udpConn) Close() error {
	common.Must(c.done.Close())
	common.Must(common.Close(c.writer))
	return nil
}

func (c *udpConn) RemoteAddr() net.Addr {
	return c.remote
}

func (c *udpConn) LocalAddr() net.Addr {
	return c.local
}

func (*udpConn) SetDeadline(time.Time) error {
	return nil
}

func (*udpConn) SetReadDeadline(time.Time) error {
	return nil
}

func (*udpConn) SetWriteDeadline(time.Time) error {
	return nil
}

type connID struct {
	src  net.Destination
	dest net.Destination
}

type udpWorker struct {
	sync.RWMutex

	proxy           proxy.Inbound
	hub             *udp.Hub
	address         net.Address
	port            net.Port
	tag             string
	stream          *internet.MemoryStreamConfig
	dispatcher      routing.Dispatcher
	uplinkCounter   stats.Counter
	downlinkCounter stats.Counter

	checker    *task.Periodic
	activeConn map[connID]*udpConn
}

func (w *udpWorker) getConnection(id connID) (*udpConn, bool) {
	w.Lock()
	defer w.Unlock()

	if conn, found := w.activeConn[id]; found && !conn.done.Done() {
		return conn, true
	}

	pReader, pWriter := pipe.New(pipe.DiscardOverflow(), pipe.WithSizeLimit(16*1024))
	conn := &udpConn{
		reader: pReader,
		writer: pWriter,
		output: func(b []byte) (int, error) {
			return w.hub.WriteTo(b, id.src)
		},
		remote: &net.UDPAddr{
			IP:   id.src.Address.IP(),
			Port: int(id.src.Port),
		},
		local: &net.UDPAddr{
			IP:   w.address.IP(),
			Port: int(w.port),
		},
		done:     done.New(),
		uplink:   w.uplinkCounter,
		downlink: w.downlinkCounter,
	}
	w.activeConn[id] = conn

	conn.updateActivity()
	return conn, false
}

func (w *udpWorker) callback(b *buf.Buffer, source net.Destination, originalDest net.Destination) {
	id := connID{
		src: source,
	}
	if originalDest.IsValid() {
		id.dest = originalDest
	}
	conn, existing := w.getConnection(id)

	// payload will be discarded in pipe is full.
	conn.writer.WriteMultiBuffer(buf.MultiBuffer{b})

	if !existing {
		common.Must(w.checker.Start())

		go func() {
			ctx := context.Background()
			sid := session.NewID()
			ctx = session.ContextWithID(ctx, sid)

			if originalDest.IsValid() {
				ctx = session.ContextWithOutbound(ctx, &session.Outbound{
					Target: originalDest,
				})
			}
			ctx = session.ContextWithInbound(ctx, &session.Inbound{
				Source:  source,
				Gateway: net.UDPDestination(w.address, w.port),
				Tag:     w.tag,
			})
			if err := w.proxy.Process(ctx, net.Network_UDP, conn, w.dispatcher); err != nil {
				newError("connection ends").Base(err).WriteToLog(session.ExportIDToError(ctx))
			}
			conn.Close()
			w.removeConn(id)
		}()
	}
}

func (w *udpWorker) removeConn(id connID) {
	w.Lock()
	delete(w.activeConn, id)
	w.Unlock()
}

func (w *udpWorker) handlePackets() {
	receive := w.hub.Receive()
	for payload := range receive {
		w.callback(payload.Payload, payload.Source, payload.Target)
	}
}

func (w *udpWorker) clean() error {
	nowSec := time.Now().Unix()
	w.Lock()
	defer w.Unlock()

	if len(w.activeConn) == 0 {
		return newError("no more connections. stopping...")
	}

	for addr, conn := range w.activeConn {
		if nowSec-atomic.LoadInt64(&conn.lastActivityTime) > 8 { // TODO Timeout too small
			delete(w.activeConn, addr)
			conn.Close()
		}
	}

	if len(w.activeConn) == 0 {
		w.activeConn = make(map[connID]*udpConn, 16)
	}

	return nil
}

func (w *udpWorker) Start() error {
	w.activeConn = make(map[connID]*udpConn, 16)
	ctx := context.Background()
	h, err := udp.ListenUDP(ctx, w.address, w.port, w.stream, udp.HubCapacity(256))
	if err != nil {
		return err
	}

	w.checker = &task.Periodic{
		Interval: time.Second * 16,
		Execute:  w.clean,
	}

	w.hub = h
	go w.handlePackets()
	return nil
}

func (w *udpWorker) Close() error {
	w.Lock()
	defer w.Unlock()

	var errors []interface{}

	if w.hub != nil {
		if err := w.hub.Close(); err != nil {
			errors = append(errors, err)
		}
	}

	if w.checker != nil {
		if err := w.checker.Close(); err != nil {
			errors = append(errors, err)
		}
	}

	if err := common.Close(w.proxy); err != nil {
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		return newError("failed to close all resources").Base(newError(serial.Concat(errors...)))
	}
	return nil
}

func (w *udpWorker) Port() net.Port {
	return w.port
}

func (w *udpWorker) Proxy() proxy.Inbound {
	return w.proxy
}
