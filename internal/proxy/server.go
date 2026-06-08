package proxy

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"sync"
)

var ErrOutboundIPRequired = errors.New("proxy outbound ip is required")

type Dialer interface {
	Dial(ctx context.Context, network, address, outboundIP string) (net.Conn, error)
}

type DefaultDialer struct{}

func (DefaultDialer) Dial(ctx context.Context, network, address, outboundIP string) (net.Conn, error) {
	outboundIP = strings.TrimSpace(outboundIP)
	if outboundIP == "" {
		return nil, ErrOutboundIPRequired
	}
	ip := net.ParseIP(outboundIP)
	if ip == nil {
		return nil, fmt.Errorf("invalid outbound ip %q", outboundIP)
	}
	dialer := net.Dialer{LocalAddr: &net.TCPAddr{IP: ip}}
	return dialer.DialContext(ctx, network, address)
}

type Config struct {
	BindHost       string
	BindPort       int
	OutboundIP     string
	MaxConnections int
	Dialer         Dialer
}

type Server struct {
	config Config
	dialer Dialer
	sem    chan struct{}
}

func NewServer(config Config) *Server {
	if config.BindHost == "" {
		config.BindHost = "0.0.0.0"
	}
	if config.MaxConnections <= 0 {
		config.MaxConnections = 200
	}
	dialer := config.Dialer
	if dialer == nil {
		dialer = DefaultDialer{}
	}
	return &Server{
		config: config,
		dialer: dialer,
		sem:    make(chan struct{}, config.MaxConnections),
	}
}

func (s *Server) Address() string {
	return net.JoinHostPort(s.config.BindHost, strconv.Itoa(s.config.BindPort))
}

func (s *Server) ListenAndServe(ctx context.Context) error {
	if s.config.BindPort <= 0 || s.config.BindPort > 65535 {
		return fmt.Errorf("proxy bind port must be between 1 and 65535")
	}
	var lc net.ListenConfig
	listener, err := lc.Listen(ctx, "tcp", s.Address())
	if err != nil {
		return err
	}
	defer listener.Close()

	go func() {
		<-ctx.Done()
		_ = listener.Close()
	}()

	for {
		conn, err := listener.Accept()
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			return err
		}
		select {
		case s.sem <- struct{}{}:
			go func() {
				defer func() { <-s.sem }()
				_ = s.Handle(ctx, conn)
			}()
		default:
			_ = conn.Close()
		}
	}
}

func (s *Server) Handle(ctx context.Context, client net.Conn) error {
	defer client.Close()

	header, err := readExact(client, 2)
	if err != nil {
		return err
	}
	if header[0] != 0x05 {
		return fmt.Errorf("unsupported socks version %d", header[0])
	}
	methods, err := readExact(client, int(header[1]))
	if err != nil {
		return err
	}
	if !containsByte(methods, 0x00) {
		_, _ = client.Write([]byte{0x05, 0xff})
		return fmt.Errorf("no supported socks auth method")
	}
	if _, err := client.Write([]byte{0x05, 0x00}); err != nil {
		return err
	}

	request, err := readExact(client, 4)
	if err != nil {
		return err
	}
	if request[0] != 0x05 {
		return fmt.Errorf("unsupported socks request version %d", request[0])
	}
	if request[1] != 0x01 {
		_ = sendReply(client, 0x07, nil)
		return nil
	}

	target, err := s.readTarget(client, request[3])
	if err != nil {
		_ = sendReply(client, 0x08, nil)
		return err
	}

	remote, err := s.dialer.Dial(ctx, "tcp", target, s.config.OutboundIP)
	if err != nil {
		_ = sendReply(client, 0x01, nil)
		return nil
	}
	defer remote.Close()

	if err := sendReply(client, 0x00, remote.LocalAddr()); err != nil {
		return err
	}
	relay(client, remote)
	return nil
}

func (s *Server) readTarget(conn net.Conn, addrType byte) (string, error) {
	var host string
	switch addrType {
	case 0x01:
		ip, err := readExact(conn, 4)
		if err != nil {
			return "", err
		}
		host = net.IP(ip).String()
	case 0x03:
		length, err := readExact(conn, 1)
		if err != nil {
			return "", err
		}
		name, err := readExact(conn, int(length[0]))
		if err != nil {
			return "", err
		}
		host = string(name)
	default:
		return "", fmt.Errorf("unsupported address type %d", addrType)
	}
	portBytes, err := readExact(conn, 2)
	if err != nil {
		return "", err
	}
	port := binary.BigEndian.Uint16(portBytes)
	return net.JoinHostPort(host, strconv.Itoa(int(port))), nil
}

func readExact(conn net.Conn, size int) ([]byte, error) {
	buf := make([]byte, size)
	_, err := io.ReadFull(conn, buf)
	return buf, err
}

func containsByte(values []byte, target byte) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func sendReply(conn net.Conn, code byte, addr net.Addr) error {
	ip := net.IPv4zero
	port := 0
	if tcpAddr, ok := addr.(*net.TCPAddr); ok {
		if v4 := tcpAddr.IP.To4(); v4 != nil {
			ip = v4
		}
		port = tcpAddr.Port
	}
	reply := []byte{0x05, code, 0x00, 0x01}
	reply = append(reply, ip.To4()...)
	portBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(portBytes, uint16(port))
	reply = append(reply, portBytes...)
	_, err := conn.Write(reply)
	return err
}

func relay(a, b net.Conn) {
	var once sync.Once
	closeBoth := func() {
		_ = a.Close()
		_ = b.Close()
	}
	done := make(chan struct{}, 2)
	go func() {
		_, _ = io.Copy(a, b)
		once.Do(closeBoth)
		done <- struct{}{}
	}()
	go func() {
		_, _ = io.Copy(b, a)
		once.Do(closeBoth)
		done <- struct{}{}
	}()
	<-done
}
