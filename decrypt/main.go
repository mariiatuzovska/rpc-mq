package main

import (
	"encoding/base64"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	aes "github.com/mariiatuzovska/rpc-mq/cipher"
)

var (
	intSize = 64
	// flags
	filePath = flag.String("file_path", "./log.txt", "File path")
	logKey   = flag.Int("log_key", 0, "Key is a number four characters")
)

func main() {

	flag.Parse()

	var fatal = func(err error) {
		if err != nil {
			log.Fatal(err)
		}
	}

	// validate flags
	if *logKey < 0 || *logKey > 9999 {
		log.Fatalf("log_key must be a number more than -1 and less than %d", 10000)
	}

	aesgcm, err := aes.New(*logKey)
	fatal(err)

	// read file
	data, err := ioutil.ReadFile(*filePath)
	fatal(err)

	if len(data) < 1 {
		log.Fatalf("Null %s file", *filePath)
	}

	var key int = 0
	for i := 0; i < intSize/16; i++ {
		key = key << 16
		key |= *logKey
	}

	var lines []string
	i := 0
	for {
		if len(data) == i {
			break
		}
		if data[i] == 10 {
			lines = append(lines, string(data[:i]))
			if len(data) > i+1 {
				data = data[i+1:]
				i = 0
			}
		}
		i++
	}

	file, err := os.Create(*filePath)
	fatal(err)
	defer file.Close()

	for _, line := range lines {
		ct, err := base64.StdEncoding.DecodeString(line)
		fatal(err)
		plainText, err := aesgcm.Decrypt(ct)
		fatal(err)
		_, err = file.WriteString(fmt.Sprintf("%s\n", string(plainText)))
		fatal(err)
	}

}
