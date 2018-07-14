package allocator

import (
	"fmt"
	"io"
	"net"
	"time"

	"go.uber.org/zap"

	"github.com/gortc/turn"
)

// Addr is ip:port.
type Addr struct {
	IP   net.IP
	Port int
}

// Equal returns true if b == a.
func (a Addr) Equal(b Addr) bool {
	if a.Port != b.Port {
		return false
	}
	return a.IP.Equal(b.IP)
}

func (a Addr) String() string {
	return fmt.Sprintf("%s:%d", a.IP, a.Port)
}

// FiveTuple represents 5-TUPLE value.
type FiveTuple struct {
	Client Addr
	Server Addr
	Proto  turn.Protocol
}

func (t FiveTuple) String() string {
	return fmt.Sprintf("%s->%s (%s)",
		t.Client, t.Server, t.Proto,
	)
}

type PeerHandler interface {
	HandlePeerData(d []byte, t FiveTuple, a Addr)
}

// Permission as described in "Permissions" section.
//
// See RFC 5766 Section 2.3
type Permission struct {
	Addr    Addr
	Timeout time.Time
}

func (p Permission) String() string {
	return fmt.Sprintf("%s [%s]", p.Addr, p.Timeout.Format(time.RFC3339))
}

// Allocation as described in "Allocations" section.
//
// See RFC 5766 Section 2.2
type Allocation struct {
	Tuple       FiveTuple
	Permissions []Permission
	Channels    []Binding
	Callback    PeerHandler
	Log         *zap.Logger
	Conn        net.PacketConn
}

func (a *Allocation) ReadUntilClosed() {
	a.Log.Debug("ReadUntilClosed")
	buf := make([]byte, 1024)
	for {
		a.Conn.SetReadDeadline(time.Now().Add(time.Minute))
		n, addr, err := a.Conn.ReadFrom(buf)
		if err != nil && err != io.EOF {
			a.Log.Error("read", zap.Error(err))
			break
		}
		udpAddr := addr.(*net.UDPAddr)
		a.Callback.HandlePeerData(buf[:n], a.Tuple, Addr{
			IP:   udpAddr.IP,
			Port: udpAddr.Port,
		})
	}
}

// Binding is binding channel.
type Binding struct {
	Number int
	Addr   Addr
}