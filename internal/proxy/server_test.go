package proxy

import (
	"context"
	"encoding/binary"
	"errors"
	"io"
	"net"
	"testing"
	"time"
)

func TestHandleConnectDialsTargetThroughOutboundIP(t *testing.T) {
	client, server := net.Pipe()
	defer client.Close()

	dialer := &fakeDialer{}
	srv := NewServer(Config{
		BindHost:       "127.0.0.1",
		BindPort:       1080,
		OutboundIP:     "10.8.0.2",
		MaxConnections: 2,
		Dialer:         dialer,
	})

	done := make(chan error, 1)
	go func() {
		done <- srv.Handle(context.Background(), server)
	}()

	writeNoAuthGreeting(t, client)
	if got := readBytes(t, client, 2); string(got) != "\x05\x00" {
		t.Fatalf("greeting reply = %#v, want no-auth", got)
	}

	writeConnectDomain(t, client, "example.test", 443)
	reply := readBytes(t, client, 10)
	if reply[1] != 0x00 {
		t.Fatalf("connect reply = %#v, want success", reply)
	}

	if dialer.network != "tcp" {
		t.Fatalf("network = %q, want tcp", dialer.network)
	}
	if dialer.address != "example.test:443" {
		t.Fatalf("address = %q, want example.test:443", dialer.address)
	}
	if dialer.outboundIP != "10.8.0.2" {
		t.Fatalf("outboundIP = %q, want 10.8.0.2", dialer.outboundIP)
	}

	client.Close()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("Handle() did not return after client close")
	}
}

func TestHandleRejectsUnsupportedCommand(t *testing.T) {
	client, server := net.Pipe()
	defer client.Close()

	srv := NewServer(Config{OutboundIP: "10.8.0.2", Dialer: &fakeDialer{}})
	done := make(chan error, 1)
	go func() {
		done <- srv.Handle(context.Background(), server)
	}()

	writeNoAuthGreeting(t, client)
	_ = readBytes(t, client, 2)
	if _, err := client.Write([]byte{0x05, 0x02, 0x00, 0x01}); err != nil {
		t.Fatalf("write unsupported command: %v", err)
	}
	reply := readBytes(t, client, 10)
	if reply[1] != 0x07 {
		t.Fatalf("reply code = %#x, want command-not-supported", reply[1])
	}
	<-done
}

func TestHandleFailsClosedWhenOutboundIPIsMissing(t *testing.T) {
	client, server := net.Pipe()
	defer client.Close()

	srv := NewServer(Config{})
	done := make(chan error, 1)
	go func() {
		done <- srv.Handle(context.Background(), server)
	}()

	writeNoAuthGreeting(t, client)
	_ = readBytes(t, client, 2)
	writeConnectDomain(t, client, "example.test", 443)
	reply := readBytes(t, client, 10)
	if reply[1] == 0x00 {
		t.Fatalf("reply code = success, want failure when outbound IP is missing")
	}
	<-done
}

type fakeDialer struct {
	network    string
	address    string
	outboundIP string
	err        error
}

func (d *fakeDialer) Dial(ctx context.Context, network, address, outboundIP string) (net.Conn, error) {
	d.network = network
	d.address = address
	d.outboundIP = outboundIP
	if d.err != nil {
		return nil, d.err
	}
	local, remote := net.Pipe()
	go func() {
		io.Copy(io.Discard, remote)
		remote.Close()
	}()
	return local, nil
}

func TestDefaultDialerRejectsMissingOutboundIP(t *testing.T) {
	_, err := DefaultDialer{}.Dial(context.Background(), "tcp", "example.test:443", "")
	if !errors.Is(err, ErrOutboundIPRequired) {
		t.Fatalf("Dial() error = %v, want ErrOutboundIPRequired", err)
	}
}

func writeNoAuthGreeting(t *testing.T, conn net.Conn) {
	t.Helper()
	if _, err := conn.Write([]byte{0x05, 0x01, 0x00}); err != nil {
		t.Fatalf("write greeting: %v", err)
	}
}

func writeConnectDomain(t *testing.T, conn net.Conn, host string, port uint16) {
	t.Helper()
	req := []byte{0x05, 0x01, 0x00, 0x03, byte(len(host))}
	req = append(req, []byte(host)...)
	portBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(portBytes, port)
	req = append(req, portBytes...)
	if _, err := conn.Write(req); err != nil {
		t.Fatalf("write connect: %v", err)
	}
}

func readBytes(t *testing.T, conn net.Conn, n int) []byte {
	t.Helper()
	buf := make([]byte, n)
	if _, err := io.ReadFull(conn, buf); err != nil {
		t.Fatalf("read %d bytes: %v", n, err)
	}
	return buf
}
