package bully

import "fmt"

// msgType represents the type of messages that can be passed between nodes
type msgType int

// customized messages to be used for communication between nodes
const (
	ELECTION msgType = iota
	OK
	PING
	COORDINATOR
	// TODO: add / change message types as needed
)

// Message is used to communicate with the other nodes
// DO NOT MODIFY THIS STRUCT
type Message struct {
	Pid   int
	Round int
	Type  msgType
}

// Bully runs the bully algorithm for leader election on a node
func Bully(pid int, leader int, checkLeader chan bool, comm map[int]chan Message, startRound <-chan int, quit <-chan bool, electionResult chan<- int) {
	// TODO: initialization code
	leaderPinged, electionInvoked, isLeader := false, false, false
	var pingID, electionID int

	if leader == pid { //checking for this condition to save state in case a leader dies
		isLeader = true
	}
	i := 1
	for {
		// quit / crash the program
		if <-quit {
			return
		}

		// start the round
		roundNum := <-startRound
		// TODO: bully algorithm code

		//going to check if a Leader is being restarted from a failure or not
		if i != roundNum {
			i = roundNum
			if isLeader {
				for i := 0; i < len(comm); i++ {
					comm[i] <- Message{pid, roundNum, COORDINATOR}
					isLeader = true
				}
			}
		}
		msgs := getMessages(comm[pid], roundNum-1)
		//fmt.Println(pid, msgs)

		//Checking what messages were sent in the previous record
		for _, msg := range msgs {
			if msg.Type == COORDINATOR {
				electionResult <- msg.Pid
				isLeader = false
			} else if msg.Type == PING {
				comm[msg.Pid] <- Message{pid, roundNum, OK}
			} else if msg.Type == ELECTION {
				if electionInvoked == false {
					electionInvoked = true
					callElection(comm, roundNum, pid)
					electionID = roundNum
				}
				comm[msg.Pid] <- Message{pid, roundNum, OK}
			} else if msg.Type == OK {
				electionInvoked = false
			} else {
				fmt.Println("wtaf bro", msg.Type)
			}
		}

		if leaderPinged == true && roundNum > (pingID+1) { //condition is invoked if the current leader is dead
			if pid > (len(comm) - 3) {
				for i := 0; i < len(comm); i++ {
					comm[i] <- Message{pid, roundNum, COORDINATOR}
					isLeader = true
				}
			} else {
				electionInvoked = true
				electionID = roundNum
				callElection(comm, roundNum, pid)
			}
			leaderPinged = false
		} else if len(checkLeader) > 0 { //awaits prompt to check up on leader
			<-checkLeader
			comm[leader] <- Message{pid, roundNum, PING}
			leaderPinged = true //boyses aajayooo
			pingID = roundNum
		} else if electionInvoked == true && roundNum > (electionID+1) { //if it wins an election without getting bullied by anyone else
			for i := 0; i < len(comm); i++ {
				comm[i] <- Message{pid, roundNum, COORDINATOR}
			}
			electionInvoked = false
		}
		i += 1
	}
}

// assumes messages from previous rounds have been read already
func getMessages(msgs chan Message, round int) []Message {
	var result []Message

	// check if there are messages to be read
	if len(msgs) > 0 {
		var m Message

		// read all messages belonging to the corresponding round
		for m = <-msgs; m.Round == round; m = <-msgs {
			result = append(result, m)
			if len(msgs) == 0 {
				break
			}
		}

		// insert back message from any other round
		if m.Round != round {
			msgs <- m
		}
	}
	return result
}

// TODO: helper functions
func callElection(comm map[int]chan Message, roundNum int, pid int) { // Self Explanatory
	for i := pid + 1; i < len(comm); i++ {
		comm[i] <- Message{pid, roundNum, ELECTION}
	}
}
