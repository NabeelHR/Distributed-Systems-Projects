// Implementation of a KeyValueServer. Students should write their code in this file.

package pa1

import (
	"DS_PA1/rpcs"
	"bufio"
	"fmt"
	"net"
	"net/http"
	"net/rpc"
	"strconv"
)

type keyValueServer struct {
	// TODO: implement this!
	count   int
	clients []net.Conn
	thred   []chan string
	counter chan int
	query   chan string

	//These channels are just for the RPC Implementation
	alpha    chan *rpcs.PutArgs
	beta     chan *rpcs.GetArgs
	rpcReply chan []byte
}

// New creates and returns (but does not start) a new KeyValueServer.
func New() KeyValueServer {
	// TODO: implement this!
	newServer := new(keyValueServer)
	newServer.count = 0
	newServer.clients = make([]net.Conn, 250)
	newServer.thred = make([]chan string, 250)
	newServer.counter = make(chan int)
	newServer.query = make(chan string)
	return newServer
}

func (kvs *keyValueServer) StartModel1(port int) error {
	// TODO: implement this!
	go kvs.keepCount()
	for i := range kvs.thred {
		kvs.thred[i] = make(chan string, 500)
	}
	initDB()
	tcpAddr, err := net.ResolveTCPAddr("tcp", ":"+strconv.Itoa(port))
	ln, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		return err
	}
	go kvs.startListening(ln)

	return nil
}
func (kvs *keyValueServer) StartModel2(port int) error {
	kvs.alpha = make(chan *rpcs.PutArgs)
	kvs.beta = make(chan *rpcs.GetArgs)
	kvs.rpcReply = make(chan []byte)

	//RPC CODE
	rpcServer := rpc.NewServer()
	rpcServer.Register(rpcs.Wrap(kvs))
	http.DefaultServeMux = http.NewServeMux()
	rpcServer.HandleHTTP(rpc.DefaultRPCPath, rpc.DefaultDebugPath)

	// if err := rpc.Register(rpcs.Wrap(kvs)); err != nil {
	// 	return err
	// }
	// http.DefaultServeMux = http.NewServeMux()
	// rpc.HandleHTTP()
	initDB()
	listener, err := net.Listen("tcp", ":"+strconv.Itoa(port))
	if err != nil {
		return err
	}
	go http.Serve(listener, nil)

	go kvs.rpcWriter()
	return nil
}

func (kvs *keyValueServer) Close() {
	// TODO: implement this!
}

func (kvs *keyValueServer) Count() int {
	// TODO: implement this!
	kvs.counter <- 0
	return <-kvs.counter
}

func (kvs *keyValueServer) startListening(ln *net.TCPListener) {
	go kvs.handleQueries()
	i := 0
	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Printf("Couldn't accept a client connection: %s\n", err)
			continue
		}
		defer conn.Close()
		kvs.clients[kvs.Count()] = conn
		kvs.counter <- 1
		go kvs.connectionListener(conn)
		go kvs.connectionSender(conn, i)
		i += 1
	}
}

func (kvs *keyValueServer) keepCount() {
	for {
		x := <-kvs.counter
		if x == 1 {
			kvs.count += 1
		} else if x == 0 {
			kvs.counter <- kvs.count
		} else {
			kvs.count -= 1
		}
	}
}

func (kvs *keyValueServer) connectionListener(conn net.Conn) {
	reader := bufio.NewReader(conn)
	for {
		msg, err := reader.ReadString('\n')
		if err != nil {
			kvs.counter <- -1
			break
		}
		kvs.query <- msg
	}
}

func (kvs *keyValueServer) connectionSender(conn net.Conn, index int) {
	for {
		respMsg := <-kvs.thred[index]
		conn.Write([]byte(respMsg))
	}
}

func (kvs *keyValueServer) handleQueries() {
	for {
		msg := <-kvs.query
		msg = msg[:len(msg)-1]
		req := splitter(msg, ",")
		if req[0] == "put" {
			put(req[1], []byte(req[2]))
		} else {
			resp := req[1] + "," + string(get(req[1])) + "\n"
			cnt := kvs.Count()
			for i := 0; i < cnt; i++ {
				// broadcasting message to all
				if cap(kvs.thred[i]) > len(kvs.thred[i]) {
					kvs.thred[i] <- resp
				}
			}
		}
	}
}

func splitter(str, delim string) []string {
	res := make([]string, 0)
	a := 0
	for i := 0; i < len(str); i++ {
		if str[i:i+1] == "," {
			res = append(res, str[a:i])
			a = i + 1
		}
	}
	res = append(res, str[a:len(str)])
	return res
}

func (kvs *keyValueServer) RecvGet(args *rpcs.GetArgs, reply *rpcs.GetReply) error {
	// TODO: implement this!
	kvs.beta <- args
	reply.Value = <-kvs.rpcReply
	return nil
}

func (kvs *keyValueServer) RecvPut(args *rpcs.PutArgs, reply *rpcs.PutReply) error {
	// TODO: implement this!
	kvs.alpha <- args
	return nil
}
func (kvs *keyValueServer) rpcWriter() {
	for {
		select {
		case putargs := <-kvs.alpha:
			put(putargs.Key, putargs.Value)
		case getargs := <-kvs.beta:
			kvs.rpcReply <- get(getargs.Key)
		}
	}

}