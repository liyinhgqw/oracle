# Timestamp Oracle

Timestamp Oracle is a golang implementation of timestamp service. It only support simple TS request function, which returns an auto-increment logic timestamp. 

The TS oracle can handle crash recovery by writing WAL to durable disk.

The TS oracle has high performance by batching TS requests. Each client maintains only one in-flight TS request RPC. TS oracle also allocate TSs in batch, which reduces the WAL io cost.

## Usage
* Install
```
    go get github.com/liyinhgqw/oracle
```

* Start Oracle Server
```
    go run server/server.go -address=":7070"
```

* Client stub

The client stub is thread-safe.

```go
    client, err := oracle.NewClient(":7070")
    if err != nil {
        log.Fatalln(err)
    }
    if ts, err := client.TS(); err != nil {
        log.Println("ts error")
    } else {
        ...
    }

    ...
    client.Close()
```

## Reference
Google percolator [percolator](http://static.googleusercontent.com/media/research.google.com/en/us/pubs/archive/36726.pdf).