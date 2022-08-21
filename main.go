package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/cloudwego/netpoll"
	"github.com/go-redis/redis/v8"
	"github.com/gobwas/ws"
	"log"
	"net/http"
	_ "net/http/pprof"
	"time"
)

// Arguments for the command line
var (
	host  = flag.String("host", "127.0.0.1", "Host to bind to")
	port  = flag.String("port", "8080", "Port to bind to")
	mode  = flag.String("mode", "echo", "Type of server to spawn, values are (echo, ping_pong)")
	debug = flag.Bool("debug", false, "Add debug info")
	pprof = flag.String("pprof", "40000", "Pprof port to bind to")
)

func main() {
	// parse command line flags
	flag.Parse()
	//
	if *debug && pprof != nil {
		log.Printf("Starting pprof server on %s \n", *pprof)
		go func() {
			url := "127.0.0.1" + ":" + string(*pprof)

			err := http.ListenAndServe(url, nil)
			if err != nil {
				log.Fatalf("Could not open pprof server : reason %s \n", err)
			}
			log.Printf("Pprof server up and running at  %s \n", url)
		}()
	}
	var addr = *host + ":" + *port
	// Create a new poller,
	poller, err := netpoll.CreateListener("tcp", addr)

	if err != nil {
		log.Printf("[ERROR]: Could not create netpoll interface , reason %s\n", err)
		return
	}
	log.Printf("[INFO]: Successfully binded to %s\n", addr)
	defer func(poller netpoll.Listener) {
		err := poller.Close()
		if err != nil {
			log.Printf("[ERROR]: Could not close netpoll listener, reason :%s\n", err)

		}
	}(poller)
	var eventLoop netpoll.EventLoop

	// deal with redis
	var rdb = redis.Client{}
	// We don't want to open a redis server in case
	// whoever is testing doesn't feel like using redis server
	if *mode == "redis" {
		rdb = *redis.NewClient(&redis.Options{
			Addr:     "localhost:6379",
			Password: "",
			DB:       0,
		})
	}
	eventLoop, _ = netpoll.NewEventLoop(
		func(ctx context.Context, connection netpoll.Connection) error {
			// do dynamic dispatch
			switch *mode {

			case "echo":
				// echo server
				return echoServer(ctx, connection)
			case "echoFrames":
				// echo server that can handle different types of frame
				return echoServerWithDifferentFrameTypes(ctx, connection)
			case "redis":
				// redis server
				return redisServer(ctx, &rdb, connection)

			default:
				return fmt.Errorf("Unknown mode %s,\n", *mode)
			}
		},
		netpoll.WithReadTimeout(time.Minute),
	)

	err = eventLoop.Serve(poller)

	if err != nil {
		log.Printf("[ERROR] : Could not create serve event loop, reason: %d \n", err)
		return
	}

}
func handle(ctx context.Context, connection netpoll.Connection) error {

	log.Printf("[INFO]: Received request from %s\n", connection.RemoteAddr().String())

	// Cleanup in case something goes wrong
	defer func(connection netpoll.Connection) {
		log.Printf("[INFO]: Closing connection with %s\n", connection.RemoteAddr().String())
		err := connection.Close()
		if err != nil {
			log.Printf("[ERROR]: Could not close this connection, reason %s\n", err)
		}
	}(connection)
	//
	_, err := ws.Upgrade(connection)
	if err != nil {
		log.Printf("[ERROR]: Upgrade error reason:%s", err)
		return err
	}
	log.Printf("[INFO]: Successfully upgraded to websockets")

	for {
		req, err := ws.ReadFrame(connection)
		if req.Payload == nil || req.Header.Length == 0 {
			time.Sleep(time.Second)
		}
		reqUnmasked := ws.UnmaskFrameInPlace(req)
		respString := string(reqUnmasked.Payload)

		log.Printf("Received : %d payload size\n", req.Header.Length)
		log.Printf("%s\n", respString)
		if respString == "close" {
			break
		}
		resp := ws.NewTextFrame([]byte("Hello world"))
		err = ws.WriteFrame(connection, resp)
		if err != nil {
			log.Printf("[ERROR]: Could not write to websocket; reason: %s\n", err)
			return err
		}
		log.Printf("[INFO]: Successfully sent message to websocket.")
	}

	err = connection.Close()
	if err != nil {
		log.Printf("[ERROR]:Could not close connection ,reason %s\n", err)
		return err
	}
	return nil
}
