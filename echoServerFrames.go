// A simple echo server handler in go websocket that can handle different
// types of frames
//
// This is a bare bones websocket implementation that creates an
// echo websocket server using gobwas and netpoll connections
//
// It is implemented using an infinite for loop and is not meant to be used
// anywhere that is performance critical since having persistent connections is a very bad performance
// pitfalls.
//
// This features a more robust echo server that can respond to different frame types
// including,
//
package main

import (
	"context"
	"github.com/cloudwego/netpoll"
	"github.com/gobwas/ws"
	"log"
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
func echoServerWithDifferentFrameTypes(ctx context.Context, connection netpoll.Connection) error {

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
	// Infinite loop this thing.
	for {
		// Read data anticipating the client sent something
		req, err := ws.ReadFrame(connection)
		// unmask data to get the correct data back
		// do it in place to reduce allocations
		reqUnmasked := ws.UnmaskFrameInPlace(req)
		// Create a new frame
		frame := ws.Frame{}
		// dispatch based on type of frame
		// this uses switch+header opcode
		// constants for this are defined in constants.go
		// which are gotten from the websockets RFC
		// @https://www.rfc-editor.org/rfc/rfc6455#section-5.2
		switch reqUnmasked.Header.OpCode {
		case Ping:
			log.Println("[INFO]: Ping request")
			frame = ws.NewPongFrame([]byte("Pong"))
		case Pong:
			log.Println("[INFO]: Pong request")
			frame = ws.NewPingFrame([]byte("Ping"))
		case TextFrame:
			log.Printf("[INFO]: Text request")
			frame = ws.NewTextFrame(reqUnmasked.Payload)
		case BinaryFrame:
			log.Printf("[INFO]: Binary request")
			frame = ws.NewBinaryFrame(reqUnmasked.Payload)
		default:
			log.Printf("[ERROR]: Unkown/(or unsupported) frames found")
			return err
		}
		err = ws.WriteFrame(connection, frame)

		if err != nil {
			log.Printf("[ERROR]: Could not write to websocket; reason: %s\n", err)
			return err
		}
		// Done
		log.Printf("[INFO]: Successfully sent message to websocket.")
	}
}
