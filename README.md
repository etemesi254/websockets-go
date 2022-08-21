## A bunch of examples to work with websockets in Go.


This repository features some examples that help 
one get started with websockets in go.

The third party libraries used are
1. [gobwas/ws](https://github.com/gobwas/ws) - A tiny websocket library for go
2. [cloudwego/netpoll](https://github.com/cloudwego/netpoll) - A high performance non-blocking I/O networking framework
3. [go-redis/redis](https://github.com/go-redis/redis)- Type safe redis client for golang

## Building and Running
To build clone this repo and navigate to it's root
and run
```shell
go build
```

To run use 
```shell
./websockets -mode "TYPE"
```
Modes here specify what type of server to spawn, they are
1. "echo"
2. "echoFrames"
3. "redis"

The default value is "echo"

"redis" mode needs a running Redis instance on the normal 
"6379" port, but you can change this inside `main.go` if you have 
different configurations

Note, one can get profiles using the `-debug` flag which
will spawn a pprof server on `40000`(default), 
to change server port, one can use the `-pprof` command line flag

### Examples

Examples include

### 1. echoServer

file: `./echoServer.go`

A simple echo server in websockets, it returns
the text it was sent.

Only works with text frames.


### 2. echoServerFrames

file: `./echoServerFrames.go`

A slightly more complicated websocket server that returns 
the data it was sent but works with binary,ping-pong and text frames.


### 3. redisServer

file: `./redisServer.go`

This creates a redis client and allows one to pass redis
commands via a websocket which can manipulate the redis instance

Supported commands are `GET key` and `SET key value`


### Todo

- [ ] mysql + websockets