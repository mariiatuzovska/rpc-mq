# rpc-secure message queuing

1. `$make all`

2. `$./consumer/consumer` - listens producer's requests

```
Usage of ./consumer/consumer:
  -buffer_size int
        Buffer size of writing into file (byte) (default 10)
  -file_path string
        File path (default "./log.txt")
  -flow_speed int
        Flow speed of writing into file byte/second (default 10000)
  -log_key int
        Not null key for file encryption. Key is a number up to 10000.
```

3. `$./producer/producer` produces fibonacci's sequence 

```
Usage of ./producer/producer:
  -generation_speed uint
        Generation speed number/second (default 10)
```

4. `./decrypt/decrypt` decrypts full lines in file

```
Usage of ./decrypt/decrypt:
  -file_path string
        File path (default "./log.txt")
  -log_key int
        Key is a number four characters
```