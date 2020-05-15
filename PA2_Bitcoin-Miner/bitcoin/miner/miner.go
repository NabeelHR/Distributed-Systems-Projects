package main

import (
	"bitcoin"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"reflect"
)

// Attempt to connect miner as a client to the server.
func joinWithServer(hostport string) (net.Conn, error) {
	c, err := net.Dial("tcp", hostport)
	if err != nil {
		return nil, err
	}
	return c, nil
}
func minHash(msg string, min, max uint64) (uint64, uint64) {
	hash, nonce := bitcoin.Hash(msg, min), min
	for i := min + 1; i <= max; i++ {
		h := bitcoin.Hash(msg, i)
		if h < hash {
			hash, nonce = h, i
		}
	}
	return hash, nonce
}
func main() {
	_ = reflect.TypeOf(5)
	const numArgs = 2
	if len(os.Args) != numArgs {
		fmt.Printf("Usage: ./%s <hostport>", os.Args[0])
		return
	}

	hostport := os.Args[1]
	miner, err := joinWithServer(hostport)
	if err != nil {
		fmt.Println("Failed to join with server:", err)
		return
	}

	defer miner.Close()

	msg, err := json.Marshal(bitcoin.NewJoin())
	_, err = miner.Write(msg)
	if err != nil {
		fmt.Println("Failed to join with server:", err)
		return
	}

	for {
		resp := make([]byte, 1024)
		var n int
		n, err = miner.Read(resp)
		if err != nil {
			fmt.Println("exiting ", err)
			break
		}
		var data bitcoin.Message
		err = json.Unmarshal(resp[:n], &data)
		if err != nil {
			fmt.Println("error: ", err)
			continue
		}
		if data.Type != bitcoin.Request {
			fmt.Println("wtf ERORR, server sending miner ", data.Type)
			continue
		}
		hash, nonce := minHash(data.Data, data.Lower, data.Upper)

		result := bitcoin.NewResult(hash, nonce)
		msg, err = json.Marshal(result)
		_, err = miner.Write(msg)
		if err != nil {
			fmt.Println("error, ", err)
		}
	}
	// TODO: implement this!
}
