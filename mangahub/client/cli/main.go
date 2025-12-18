package main

import (
	"mangahub/client/cli/protocol"
)

func main() {
	client := protocol.NewClient()

	client.ShowWelcome()
	client.MainMenu()
}
