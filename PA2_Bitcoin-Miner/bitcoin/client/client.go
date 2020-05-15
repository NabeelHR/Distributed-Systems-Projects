package main

import (
	"bitcoin"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strconv"
)
func main() {
	const numArgs = 4
	if len(os.Args) != numArgs {
		fmt.Printf("Usage: ./%s <hostport> <message> <maxNonce>", os.Args[0])
		return
	}
	hostport := os.Args[1]
	message := os.Args[2]
	maxNonce, err := strconv.ParseUint(os.Args[3], 10, 64)
	if err != nil {
		fmt.Printf("%s is not a number.\n", os.Args[3])
		return
	}

	client, err := net.Dial("tcp", hostport)
	if err != nil {
		fmt.Println("Failed to connect to server:", err)
		return
	}

	defer client.Close()

	msg, err := json.Marshal(bitcoin.NewRequest(message, 0, maxNonce))
	_, err = client.Write(msg)
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	result := make([]byte, 1024)
	var n int
	n, err = client.Read(result)
	if err != nil {
		printDisconnected()
		return
	}
	var data bitcoin.Message
	err = json.Unmarshal(result[:n], &data)

	if data.Type != bitcoin.Result {
		fmt.Println("error, ", data.Type)
		return
	}
	printResult(data.Hash, data.Nonce)

}

// printResult prints the final result to stdout.
func printResult(hash, nonce uint64) {
	fmt.Println("Result", hash, nonce)
}

// printDisconnected prints a disconnected message to stdout.
func printDisconnected() {
	fmt.Println("Disconnected")
}
