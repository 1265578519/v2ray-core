package http_test

import (
	"context"
	"crypto/rand"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"v2ray.com/core/common"
	"v2ray.com/core/common/buf"
	"v2ray.com/core/common/net"
	"v2ray.com/core/common/protocol/tls/cert"
	"v2ray.com/core/testing/servers/tcp"
	"v2ray.com/core/transport/internet"
	. "v2ray.com/core/transport/internet/http"
	"v2ray.com/core/transport/internet/tls"
)

func TestHTTPConnection(t *testing.T) {
	port := tcp.PickPort()

	listener, err := Listen(context.Background(), net.LocalHostIP, port, &internet.MemoryStreamConfig{
		ProtocolName:     "http",
		ProtocolSettings: &Config{},
		SecurityType:     "tls",
		SecuritySettings: &tls.Config{
			Certificate: []*tls.Certificate{tls.ParseCertificate(cert.MustGenerate(nil, cert.CommonName("www.v2ray.com")))},
		},
	}, func(conn internet.Connection) {
		go func() {
			defer conn.Close()

			b := buf.New()
			defer b.Release()

			for {
				if _, err := b.ReadFrom(conn); err != nil {
					return
				}
				_, err := conn.Write(b.Bytes())
				common.Must(err)
			}
		}()
	})
	common.Must(err)

	defer listener.Close()

	time.Sleep(time.Second)

	dctx := context.Background()
	conn, err := Dial(dctx, net.TCPDestination(net.LocalHostIP, port), &internet.MemoryStreamConfig{
		ProtocolName:     "http",
		ProtocolSettings: &Config{},
		SecurityType:     "tls",
		SecuritySettings: &tls.Config{
			ServerName:    "www.v2ray.com",
			AllowInsecure: true,
		},
	})
	common.Must(err)
	defer conn.Close()

	const N = 1024
	b1 := make([]byte, N)
	common.Must2(rand.Read(b1))
	b2 := buf.New()

	nBytes, err := conn.Write(b1)
	common.Must(err)
	if nBytes != N {
		t.Error("write: ", nBytes)
	}

	b2.Clear()
	common.Must2(b2.ReadFullFrom(conn, N))
	if r := cmp.Diff(b2.Bytes(), b1); r != "" {
		t.Error(r)
	}

	nBytes, err = conn.Write(b1)
	common.Must(err)
	if nBytes != N {
		t.Error("write: ", nBytes)
	}

	b2.Clear()
	common.Must2(b2.ReadFullFrom(conn, N))
	if r := cmp.Diff(b2.Bytes(), b1); r != "" {
		t.Error(r)
	}
}
