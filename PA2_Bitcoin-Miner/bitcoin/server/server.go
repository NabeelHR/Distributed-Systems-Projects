package main

import (
	"bitcoin"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"reflect"
	"strconv"
)

/*
I maintain a list of Miners(struct) in my server. All incoming miners are appended there
A separate go routine 'clientThread'is run for each client, which communicates with my scheduler  its
requests(sent via channel). The scheduler goes over all client connections in a loop
repeatedly to check for any requests and handles one chunk at a time by forwarding that
partitioned job to a miner and returning the result to that specific client's go routine.
The client thread keeps on sending chunks of 400 nonces to the scheduler until it has completed
its queries. The thread then returns the result Message to the client and terminates.
*/
func getMsg(conn net.Conn) bitcoin.Message {
	resp := make([]byte, 1024)
	var n int
	n, err := conn.Read(resp)
	if err != nil {
		fmt.Println(err)
	}
	var data bitcoin.Message
	err = json.Unmarshal(resp[:n], &data)
	return data
}

type miner struct {
	conn net.Conn
	busy bool
}
type server struct {
	listener      net.Listener
	miners        []miner
	clientReqs    []chan *bitcoin.Message
	clientReplies []chan *bitcoin.Message
}

func startServer(port int) (*server, error) {
	reflect.TypeOf(3)
	srv := new(server)
	srv.miners = make([]miner, 0)
	srv.clientReqs = make([]chan *bitcoin.Message, 200)
	srv.clientReplies = make([]chan *bitcoin.Message, 200)
	for i := range srv.clientReqs {
		srv.clientReqs[i] = make(chan *bitcoin.Message, 200)
		srv.clientReplies[i] = make(chan *bitcoin.Message, 200)
	}
	ln, err := net.Listen("tcp", ":"+strconv.Itoa(port))
	if err != nil {
		return nil, err
	}
	srv.listener = ln
	return srv, nil
}

func (srv *server) connectionListener() {
	i, cli := 0, 0
	for {
		fmt.Println("ln54")
		conn, err := srv.listener.Accept()
		if err != nil {
			return
		}
		LOGF.Println("Connection made")
		message := getMsg(conn)

		if message.Type == bitcoin.Join {
			var mnr miner
			mnr.conn, mnr.busy = conn, false
			srv.miners = append(srv.miners, mnr)
			i += 1
			LOGF.Println("Minner appended")
		} else if message.Type == bitcoin.Request {
			go srv.clientHandler(conn, message, cli)
			cli += 1
			LOGF.Println("Client boi appended")
			// srv.client = conn
			// srv.serve(data)
		}
		//		srv.addClient <- conn // Sending the newly-accepted client connection to clientMapHandler for storage
	}
}

func (srv *server) clientHandler(conn net.Conn, msg bitcoin.Message, cli int) {
	var result *bitcoin.Message
	var i uint64
	subJobs := 0
	for i = 0; i <= msg.Upper; i += 400 {
		subJobs += 1
		max := i + 400
		if max > msg.Upper {
			max = msg.Upper
		}
		woot := bitcoin.NewRequest(msg.Data, i, max)
		srv.clientReqs[cli] <- woot
		LOGF.Println("89")
		data := <-srv.clientReplies[cli]
		LOGF.Println("92-->", data)

		if i == 0 {
			result = bitcoin.NewResult(data.Hash, data.Nonce)
		}
		if data.Hash < result.Hash {
			result.Hash = data.Hash
			result.Nonce = data.Nonce
		}
		LOGF.Println(result)
	}
	// LOGF.Println("iterations->", subJobs)
	// for j:=0; j<=subJobs; j+=1{
	// 	data := <- srv.clientReplies[cli]
	// 	LOGF.Println("92")

	// 	if j==0 {
	// 		result = bitcoin.NewResult(data.Hash, data.Nonce)
	// 	}
	// 	if data.Hash < result.Hash {
	// 		result.Hash = data.Hash
	// 		result.Nonce = data.Nonce
	// 	}
	// }
	LOGF.Println("results are=> ", result)
	reply, _ := json.Marshal(result)
	_, _ = conn.Write(reply)
}

func (srv *server) scheduler() {
	for {
		for i := 0; i < 50; i += 1 {
			if len(srv.clientReqs[i]) > 0 {
				LOGF.Println("132")
				msg := <-srv.clientReqs[i]
				woot := srv.outsourceTask(msg)
				srv.clientReplies[i] <- woot
				fmt.Println("135", i)
			}
			LOGF.Println("results are=>mmmmmm@134")

		}

	}
}
func (srv *server) outsourceTask(msg *bitcoin.Message) *bitcoin.Message {
	newMsg, err := json.Marshal(msg)
	for len(srv.miners) == 0 {
	}
	_, err = srv.miners[0].conn.Write(newMsg)
	if err != nil {
		fmt.Println("error:", err)
	}
	reply := getMsg(srv.miners[0].conn)
	return bitcoin.NewResult(reply.Hash, reply.Nonce)
}

var LOGF *log.Logger

func main() {

	// You may need a logger for debug purpose
	const (
		name = "log.txt"
		flag = os.O_RDWR | os.O_CREATE
		perm = os.FileMode(0666)
	)

	file, err := os.OpenFile(name, flag, perm)
	if err != nil {
		return
	}
	defer file.Close()

	LOGF = log.New(file, "", log.Lshortfile|log.Lmicroseconds)
	// Usage: LOGF.Println() or LOGF.Printf()
	//LOGF.Println("Yallah")

	const numArgs = 2
	if len(os.Args) != numArgs {
		fmt.Printf("Usage: ./%s <port>", os.Args[0])
		return
	}

	port, err := strconv.Atoi(os.Args[1])
	if err != nil {
		fmt.Println("Port must be a number:", err)
		return
	}

	srv, err := startServer(port)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Println("Server listening on port", port)

	defer srv.listener.Close()

	LOGF.Println("Yallah")
	// TODO: implement this!
	go srv.scheduler()
	srv.connectionListener()
	LOGF.Println("Yallah")
}
