package main

import (
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"flag"
	"log"
	"net"
	"net/rpc"
	"os"
	"time"

	"github.com/mariiatuzovska/rpc-mq/logger"
)

var (
	address = "localhost:6000"
	// flags
	filePath   = flag.String("file_path", "./log.txt", "File path")
	flowSpeed  = flag.Int("flow_speed", 10000, "Flow speed of writing into file byte/second")
	bufferSize = flag.Int("buffer_size", 10, "Buffer size of writing into file (byte)")
	logKey     = flag.Int("log_key", 0, "Not null key for file encryption. Key is a number up to 10000.")
)

func main() {

	flag.Parse()

	var fatal = func(err error) {
		if err != nil {
			log.Fatal(err)
		}
	}

	// validate flags
	file, err := os.OpenFile(*filePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, os.ModePerm)
	fatal(err)
	defer file.Close()
	if limFlowSpeed := time.Second.Nanoseconds() / 10; *flowSpeed < 1 || int64(*flowSpeed) > limFlowSpeed {
		log.Fatalln("flow_speed must be a number more than 1 and less than", limFlowSpeed)
	}
	if *bufferSize < 1 {
		log.Fatalln("buffer_size must be a number more than 1")
	}
	if *logKey < 0 || *logKey > 9999 {
		log.Fatalf("log_key must be a number more than -1 and less than %d", 10000)
	}

	info, err := file.Stat()
	fatal(err)
	if info.Size() > 0 {
		file.WriteString("\n")
	}

	// create service
	srv := logger.New(file, *flowSpeed, *bufferSize, *logKey)
	go logger.Process(srv)

	// prepare secure connection
	certs, err := tls.LoadX509KeyPair("./certs/consumer-crt.pem", "./certs/consumer-key.pem")
	fatal(err)
	if certs.Certificate == nil || len(certs.Certificate) < 2 {
		log.Fatal("Certificate for server is nil or count of certificates is not equal two")
	}
	certPool := x509.NewCertPool()
	caCertX509, err := x509.ParseCertificate(certs.Certificate[1])
	fatal(err)
	certPool.AddCert(caCertX509)

	// create rpc server
	handler := rpc.NewServer()
	fatal(handler.Register(srv))

	listener, err := tls.Listen("tcp", address, &tls.Config{
		Certificates: []tls.Certificate{certs},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    certPool,
		RootCAs:      certPool,
		Rand:         rand.Reader,
	})
	fatal(err)

	// listen
	for {
		conn, err := listener.Accept()
		fatal(err)
		// log.Printf("Accepted connection %s", conn.RemoteAddr())
		go func(conn net.Conn, handler *rpc.Server) {
			defer conn.Close()
			handler.ServeConn(conn)
			// log.Printf("%s connection is closed", conn.RemoteAddr())
		}(conn, handler)
	}
}
