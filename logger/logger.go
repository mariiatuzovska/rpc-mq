package logger

import (
	"encoding/base64"
	"fmt"
	"log"

	"os"
	"sync"
	"time"

	aes "github.com/mariiatuzovska/rpc-mq/cipher"
)

var fMux sync.Mutex
var intSize = 64

type Logger struct {
	encrypt              bool
	file                 *os.File
	flowSleepNanoseconds int
	bufferSize           int
	aesgcm               *aes.AESGCM
	inputNumbers         chan int
}

type LoggerRequest struct {
	Number int
}

type LoggerResponse struct {
	Ok bool
}

func New(file *os.File, flowSpeed int, bufferSize int, logKey int) *Logger {
	aesgcm, err := aes.New(logKey)
	if err != nil {
		log.Fatalln(err)
	}
	encrypt := false
	if logKey != 0 {
		encrypt = true
	}
	return &Logger{
		encrypt:              encrypt,
		file:                 file,
		flowSleepNanoseconds: (int(time.Second.Nanoseconds()) / flowSpeed) - 5, // let's take five nanoseconds as the approximate time of writting 1 byte to a file
		bufferSize:           bufferSize,
		aesgcm:               aesgcm,
		inputNumbers:         make(chan int, 0x10000)}
}

func (logger *Logger) Write(request *LoggerRequest, response *LoggerResponse) error {
	logger.inputNumbers <- request.Number
	response.Ok = true
	return nil
}

func Process(logger *Logger) {
	bufferedData, index := make([]byte, logger.bufferSize), 0
	for {
		// get number from query
		number, ok := <-logger.inputNumbers
		// fmt.Printf(" %s | %d has been got\n", time.Now().UTC().String(), number^logger.key)
		if !ok {
			return
		}
		// encrypt and write into buffer
		row := fmt.Sprintf(" %s | %d", time.Now().UTC().String(), number)
		byteArray := []byte(row)
		if logger.encrypt {
			cipherText, err := logger.aesgcm.Encrypt(byteArray)
			if err != nil {
				log.Fatalln(err)
			}
			byteArray = []byte(base64.StdEncoding.EncodeToString(cipherText))
		}
		byteArray = append(byteArray, 10)
		for i := range byteArray {
			if index == logger.bufferSize {
				// fmt.Printf(" %s | buffer if full %d\n", time.Now().UTC().String(), index)
				fMux.Lock()
				logger.write(bufferedData)
				fMux.Unlock()
				bufferedData, index = make([]byte, logger.bufferSize), 0
			}
			// write data into buffer
			bufferedData[index] = byteArray[i]
			index++
		}
	}
}

func (logger *Logger) write(data []byte) {
	for {
		logger.file.Write([]byte{data[0]})
		if len(data) > 1 {
			time.Sleep(time.Nanosecond * time.Duration(logger.flowSleepNanoseconds))
		} else {
			return
		}
		data = data[1:]
	}
}
