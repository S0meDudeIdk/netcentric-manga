package protocol

import (
	"fmt"
	grpcClient "mangahub/internal/grpc"
)

// ConnectGRPC connects to the gRPC server
func (c *Client) ConnectGRPC() {
	client, err := grpcClient.NewClient(grpcAddr)
	if err != nil {
		fmt.Printf("%s⚠️  Failed to connect to gRPC server: %v%s\n", colorYellow, err, colorReset)
		c.grpcEnabled = false
		return
	}

	c.grpcClient = client
	c.grpcEnabled = true
	fmt.Println(colorGreen + "✅ Connected to gRPC server" + colorReset)
}
