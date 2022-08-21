// Simple utilities used around this project
package main

// Interpretation of payload data
//, see https://www.rfc-editor.org/rfc/rfc6455#section-5.2
const (
	ContinuationFrame = 0x0
	TextFrame         = 0x1
	BinaryFrame       = 0x2
	Ping              = 0x9
	Pong              = 0xA
	Reserved
)
