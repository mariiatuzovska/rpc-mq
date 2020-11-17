## all: Building binaries
all:
	go build -o ./consumer/consumer ./consumer  
	go build -o ./producer/producer ./producer
	go build -o ./decrypt/decrypt ./decrypt

