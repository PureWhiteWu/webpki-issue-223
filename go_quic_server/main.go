package main

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net"
	"os"
	"sync/atomic"
	"time"

	"github.com/felixge/fgprof"
	"github.com/lucas-clemente/quic-go"

	"net/http"
	_ "net/http/pprof"
)

func generateTLSConfig() (*tls.Config, error) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}
	template := x509.Certificate{SerialNumber: big.NewInt(1)}
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		return nil, err
	}

	tlsCert := tls.Certificate{
		Certificate: [][]byte{certDER},
		PrivateKey:  key,
		Leaf:        nil,
	}

	return &tls.Config{Certificates: []tls.Certificate{tlsCert}, NextProtos: []string{"quic-server"}, InsecureSkipVerify: true}, nil
}

func serverRun() {
	go func() {
		fmt.Println(http.ListenAndServe("0.0.0.0:6060", nil))

	}()

	tlsConfig, err := generateTLSConfig()
	if err != nil {
		fmt.Println("generateTLSConfig", err)
		os.Exit(1)
	}
	udpConn, err := net.ListenUDP("udp4", &net.UDPAddr{Port: 9900})
	// ... error handling
	if err != nil {
		panic(err)
	}
	ln, err := quic.Listen(udpConn, tlsConfig, &quic.Config{})
	if err != nil {
		fmt.Println("ListenWithConfig", err)
		os.Exit(1)
	}
	for {
		conn, err := ln.Accept(context.Background())
		if err != nil {
			fmt.Println(err)
			break
		}
		atomic.AddInt64(&cNum, 1)
		sizeSlice := []int{15100, 300, 300, 300} //  1.6Mb/s  200KB/s   50pkt/s   4KB/pkt

		Sleep := 20
		go func(conn quic.Connection) {
			stream, err := conn.AcceptStream(context.Background())
			if err != nil {
				panic(err)
			}
			defer func() {
				stream.Close()
				atomic.AddInt64(&cNum, -1)
			}()
			var data = make([]byte, 4000)
			_, err = stream.Read(data)
			if err != nil && !errors.Is(err, io.EOF) {
				fmt.Println("conn.Read", err)
				return
			}
			t := time.NewTicker(time.Duration(Sleep) * time.Millisecond)
			index := 0
			for {
				<-t.C
				WriteSize := sizeSlice[index%len(sizeSlice)]
				index++
				_ = stream.SetWriteDeadline(time.Now().Add(5 * time.Second))
				data = make([]byte, WriteSize, WriteSize)
				_, err := stream.Write(data)
				atomic.AddUint64(&bitrate, uint64(WriteSize))
				if err != nil {
					fmt.Println("conn.Write", conn.RemoteAddr().String(), err)
					break
				}
			}
		}(conn)
	}
}

var bitrate uint64
var cNum int64

func main() {

	go func() {
		fmt.Println(http.ListenAndServe("0.0.0.0:6000", fgprof.Handler()))
	}()

	go func() {
		for {
			<-time.After(time.Second)
			fmt.Println(
				"bitrate:", atomic.SwapUint64(&bitrate, 0)*8/1024/1024, "Mb/s",
				"sess:", atomic.LoadInt64(&cNum))
		}
	}()

	go serverRun()

	select {}
}
