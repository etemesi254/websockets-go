// A websocket that allows one to manipulate redis instances
package main

import (
	"context"
	"fmt"
	"github.com/cloudwego/netpoll"
	"github.com/go-redis/redis/v8"
	"github.com/gobwas/ws"
	"log"
	"regexp"
	"strings"
	"time"
)

var (
	redisCommand, _ = regexp.Compile("^\\S*") // Matches anything before the first word

)

func redisServer(ctx context.Context, rdb *redis.Client, connection netpoll.Connection) error {

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
		// only deal with text commands
		if req.Header.OpCode == TextFrame {
			// unmask data to get the correct data back
			// do it in place to reduce allocations
			reqUnmasked := ws.UnmaskFrameInPlace(req)

			command := string(reqUnmasked.Payload)
			// Get our command to execute,
			// This can be replaced with string.Fields but I'm too lazy
			match := redisCommand.FindString(command)

			words := strings.Fields(command)
			// Dispatch depending on request type
			// we only support get and set
			switch match {
			case "GET":
				log.Println("Got GET command")
				// For GET commands we need a key and command, so it looks like GET key
				// meaning if we are splitting, our key is at words[1]
				if len(words) < 1 {
					resp := ws.NewTextFrame([]byte("[ERROR]: No key passed, syntax is GET key"))
					err = ws.WriteFrame(connection, resp)
					if err != nil {
						log.Printf("[ERROR]: Could not write to websocket; reason: %s\n", err)
						return err
					}
				}
				// fetch  the key
				// I feel like I should not be commenting this much
				result, err := rdb.Get(ctx, words[1]).Result()

				if err != nil {
					resp := ws.NewTextFrame([]byte(fmt.Sprintf("[ERROR]: No value for key %s, try giving it a value", words[1])))
					err = ws.WriteFrame(connection, resp)
					if err != nil {
						log.Printf("[ERROR]: Could not write to websocket; reason: %s\n", err)
						return err
					}
				}
				// Write out data back
				resp := ws.NewTextFrame([]byte(result))

				err = ws.WriteFrame(connection, resp)

				if err != nil {
					log.Printf("[ERROR]: Could not write to websocket; reason: %s\n", err)
					return err
				}
			case "SET":
				log.Println("Got SET command")

				if len(words) < 2 {
					resp := ws.NewTextFrame([]byte("[ERROR]: No key or value, command is SET KEY VALUE"))

					err = ws.WriteFrame(connection, resp)

					if err != nil {
						log.Printf("[ERROR]: Could not write to websocket; reason: %s\n", err)
						return err
					}
				}
				// Set the value
				err := rdb.Set(ctx, words[1], words[2], time.Hour).Err()

				if err != nil {

					resp := ws.NewTextFrame([]byte(fmt.Sprintf("[ERROR]: Could not set value to redis, reason :%d", err)))

					err = ws.WriteFrame(connection, resp)

					if err != nil {
						log.Printf("[ERROR]: Could not write to websocket; reason: %s\n", err)
						return err
					}
				}
				resp := ws.NewTextFrame([]byte(fmt.Sprintf("Successfully set %s to %s, retrieve it using GET %s\n", words[1], words[2], words[1])))

				err = ws.WriteFrame(connection, resp)

				if err != nil {
					log.Printf("[ERROR]: Could not write to websocket; reason: %s\n", err)
					return err
				}
			default:
				log.Printf("[WARNING]: Unkown command  %s \n", command)
				// Not one of the two commands
				resp := ws.NewTextFrame([]byte("Unknown command, supported commands are GET key and SET key value"))

				err = ws.WriteFrame(connection, resp)

				if err != nil {
					log.Printf("[ERROR]: Could not write to websocket; reason: %s\n", err)
					return err
				}
			}
		} else {
			resp := ws.NewTextFrame([]byte("[ERROR]: Redis instance needs a Text Frame"))

			err = ws.WriteFrame(connection, resp)

			if err != nil {
				log.Printf("[ERROR]: Could not write to websocket; reason: %s\n", err)
				return err
			}
		}
	}

}
