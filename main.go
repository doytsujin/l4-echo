package main

import (
	"fmt"
	"log"
	"net"
	"regexp"
	"runtime"
	"strings"

	"golang.org/x/sync/errgroup"
)

var version = "0.0.0-src"
var port = 1500

func main() {
	log.Printf("l4-echo (version %s)\n", version)
	// run TCP and UDP servers concurrently
	eg := errgroup.Group{}
	eg.Go(tcp)
	eg.Go(udp)
	if err := eg.Wait(); err != nil {
		log.Fatal(err)
	}
}

func tcp() error {
	// ask OS to listen on TCP port 1500
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return err
	}
	log.Printf("listening on %d/tcp", port)
	for {
		conn, err := l.Accept()
		if err != nil {
			return err
		}
		go tcpConn(conn)
	}
}

func tcpConn(conn net.Conn) {
	// handle TCP connection
	defer conn.Close()
	ip, port, _ := net.SplitHostPort(conn.RemoteAddr().String())
	// greet client
	log.Printf("tcp: %15s: connected on port %s", ip, port)
	fmt.Fprintf(conn, "hello %s, you are connected to demo.jpillora.com on port %s\n", ip, port)
	// read data stream
	buff := make([]byte, 32*1024)
	p := 1
	for {
		// read ~one packet
		n, err := conn.Read(buff)
		if err != nil {
			return
		}
		r := reply(buff[:n])
		// include packet number
		r = []byte(fmt.Sprintf("#%d: %s", p, r))
		p++
		// write back
		conn.Write(r)
		// log to console
		log.Printf("tcp: %15s: port %5s: %s", ip, port, string(r))
	}
}

func udp() error {
	// ask OS to listen on UDP port 1500
	l, err := net.ListenPacket("udp", fmt.Sprintf("%s:%d", host(), port))
	if err != nil {
		return err
	}
	log.Printf("listening on %d/udp", port)
	defer l.Close()
	// handle individual UDP packets (no UDP "connections")
	buff := make([]byte, 32*1024)
	for {
		// read ~one packet
		n, addr, err := l.ReadFrom(buff)
		if err != nil {
			return err
		}
		// reply back
		r := reply(buff[:n])
		l.WriteTo(r, addr)
		ip, port, _ := net.SplitHostPort(addr.String())
		log.Printf("udp: %15s: port %5s: %s", ip, port, string(r))
	}
}

var special = regexp.MustCompile(`[^\w\s]`)

func reply(b []byte) []byte {
	s := strings.TrimSpace(special.ReplaceAllString(string(b), ""))
	return []byte(fmt.Sprintf("recieved %d bytes '%s'\n", len(b), s))
}

func host() string {
	if runtime.GOOS == "darwin" {
		return ""
	}
	// only needed to run UDP on fly.io
	return "fly-global-services"
}
