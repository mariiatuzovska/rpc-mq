package main

import (
	"crypto/tls"
	"crypto/x509"
	"flag"
	"log"
	"net/rpc"
	"time"

	"github.com/mariiatuzovska/rpc-mq/logger"
)

var (
	address = "localhost:6000"
	// flags
	generationSpeed = flag.Uint("generation_speed", 10, "Generation speed number/second")
)

func main() {

	flag.Parse()

	var fatal = func(err error) {
		if err != nil {
			log.Fatal(err)
		}
	}

	// validate flags and create service
	if *generationSpeed < 1 {
		log.Fatal("generation_speed must be more than 0")
	}

	// prepare secure connection and client
	certs, err := tls.LoadX509KeyPair("./certs/producer-crt.pem", "./certs/producer-key.pem")
	fatal(err)
	if certs.Certificate == nil || len(certs.Certificate) < 2 {
		log.Fatal("Certificate for server is nil or count of certificates is not equal two")
	}
	certPool := x509.NewCertPool()
	caCertX509, err := x509.ParseCertificate(certs.Certificate[1])
	fatal(err)
	certPool.AddCert(caCertX509)
	ca, err := x509.ParseCertificate(certs.Certificate[1])
	fatal(err)
	certPool.AddCert(ca)

	conn, err := tls.Dial("tcp", address, &tls.Config{
		Certificates: []tls.Certificate{certs},
		RootCAs:      certPool,
		ClientCAs:    certPool,
	})
	fatal(err)
	client := rpc.NewClient(conn)

	fib, index := make([]int, 2), 0
	fib[0] = 0
	fib[1] = 1
	for {
		t1 := time.Now()
		var i uint = 0
		for i = 0; i < *generationSpeed; i++ {
			response := &logger.LoggerResponse{Ok: false}
			if err := client.Call("Logger.Write", &logger.LoggerRequest{Number: fib[index]}, &response); err != nil {
				fatal(err)
			}
			if fib[index+1] < fib[index] {
				fatal(client.Close())
				return
			}
			fib = append(fib, fib[index]+fib[index+1])
			index++
		}
		if t2, sec := time.Now().Sub(t1).Nanoseconds(), time.Second.Nanoseconds(); t2 < sec {
			time.Sleep(time.Duration(sec-t2) * time.Nanosecond)
		}
	}
}
