// A simple echo server handler in go websocket
//
// This is a bare bones websocket implementation that creates an
// echo websocket server using gobwas and netpoll connections
//
// It is implemented using an infinite for loop and is not meant to be used
// anywhere that is performance critical since having persistent connections is a very bad performance
// pitfalls.
//
//
// This is primitive and only works with string frames, not binary frames
package main

import (
	"context"
	"github.com/cloudwego/netpoll"
	"github.com/gobwas/ws"
	"log"
	"time"
)

// Create an echo server handler
//
// @params
//
// - ctx : Google's answer to passing request scoped values,
// not required for this demonstration
//
// - connection: The encapsulation of a client request, this request should come via the ws:// protocol
// otherwise it will error out.
func echoServer(ctx context.Context, connection netpoll.Connection) error {

	log.Printf("[INFO]: Received request from %s\n", connection.RemoteAddr().String())

	// Ensure we close this connection after we are done
	defer func(connection netpoll.Connection) {
		log.Printf("[INFO]: Closing connection with %s\n", connection.RemoteAddr().String())
		err := connection.Close()
		if err != nil {
			log.Printf("[ERROR]: Could not close this connection, reason %s\n", err)
		}
	}(connection)

	// Upgrade connection to websocket
	_, err := ws.Upgrade(connection)
	if err != nil {
		log.Printf("[ERROR]: Upgrade error reason:%s", err)
		return err
	}
	log.Printf("[INFO]: Successfully upgraded to websockets")
	// Infinite loop away
	for {
		// Read data anticipating the client sent something
		req, err := ws.ReadFrame(connection)
		if req.Payload == nil || req.Header.Length == 0 {
			// nop no data arrived, so let's sleep a second, and we try again
			time.Sleep(time.Second)
			continue
		}
		// unmask data to get the correct data back
		// do it in place to reduce allocations
		reqUnmasked := ws.UnmaskFrameInPlace(req)
		// now get the actual value, assuming it's a string
		respString := string(reqUnmasked.Payload)
		// check for the close command, if that is the case, close the connection
		if respString == "close" {
			log.Printf("[INFO]: Received close command from %s, quiting", connection.RemoteAddr().String())
			break
		}
		// write back exactly what we received
		resp := ws.NewTextFrame(reqUnmasked.Payload)

		err = ws.WriteFrame(connection, resp)

		if err != nil {
			log.Printf("[ERROR]: Could not write to websocket; reason: %s\n", err)
			return err
		}
		// Done
		log.Printf("[INFO]: Successfully sent message to websocket.")
	}

	err = connection.Close()
	if err != nil {
		log.Printf("[ERROR]:Could not close connection ,reason %s\n", err)
		return err
	}
	return nil
}
