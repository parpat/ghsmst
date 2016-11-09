package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
	//"github.com/parpat/ghsmst"
)

//Node states
const (
	Sleeping string = "Sleeping"
	Find     string = "Find"
	Found    string = "Found"
)

//Infinity represents
const Infinity int = 9999

//Node is a vertex on the graph
type Node struct {
	SN string //Node state
	FN int    //Fragment ID
	LN int    //Level

	ID int

	bestEdge      *Edge
	bestWt        int
	testEdge      *Edge
	inBranch      *Edge
	findCount     int
	adjacencyList *Edges
}

//Wakeup initializes the node
func (n *Node) Wakeup() {
	(*n.adjacencyList)[0].SE = Branch
	n.LN = 0
	n.SN = Found
	n.findCount = 0

	(*n.adjacencyList)[0].Connect(0)
}

//ConnectResponse responds to the Connect message
func (n *Node) ConnectResponse(L int, j *Edge, reqQ chan *Message, m Message) {
	if n.SN == Sleeping {
		n.Wakeup()
	}

	if L < n.LN {
		j.SE = Branch
		j.Initiate(n.LN, n.FN, n.SN)
		if n.SN == Find {
			n.findCount++
		}
	} else if j.SE == Basic {
		fmt.Printf("LEVEL: %v\n", n.LN)
		time.Sleep(time.Millisecond * 15)
		reqQ <- &m //place received message in the end of Q****

	} else {
		j.Initiate(n.LN+1, j.Weight, Find)
	}

}

//InitiateResponse responds to the Initiate message
func (n *Node) InitiateResponse(L, F int, S string, j *Edge) {
	n.LN = L
	n.FN = F
	n.SN = S
	n.inBranch = j

	n.bestEdge = nil
	n.bestWt = Infinity

	for _, i := range *n.adjacencyList {
		if i != *j && i.SE == Branch {
			i.Initiate(L, F, S)
			if S == Find {
				n.findCount++
			}
		}
	}

	if S == Find {
		n.Test()
	}
}

//Test picks the minimum Basic Edge and send test message
func (n *Node) Test() {
	report := true
	for _, e := range *n.adjacencyList {
		if e.SE == Basic {
			n.testEdge = &e
			n.testEdge.Test(n.LN, n.FN)
			report = false
			break
		}
	}
	if report {
		n.testEdge = nil
		n.Report()
	}

	// if there are adjacent Edges in state Basic{
	// 	n.testEdge = min weight adjacent edge with state Basic
	// 	send Test(n.LN, n.FN) on n.testEdge
	// }else{
	// 	n.testEdge = nil
	// 	Report()
	// }
}

//TestResponse responds to Test message
func (n *Node) TestResponse(l, f int, j *Edge, reqQ chan *Message, m Message) {
	if n.SN == Sleeping {
		n.Wakeup()
	}
	if l > n.LN {
		reqQ <- &m //Put message end of Q ***************
	} else if f != n.FN {
		j.Accept()
	} else {
		if j.SE == Basic {
			j.SE = Rejected
		}
		if *n.testEdge != *j {
			j.Reject()
		} else {
			n.Test()
		}
	}
}

//AcceptResponse is a response to Accept message
func (n *Node) AcceptResponse(j *Edge) {
	n.testEdge = nil
	if j.Weight < n.bestWt {
		n.bestEdge = j
		n.bestWt = j.Weight
	}
	n.Report()
}

func (n *Node) RejectResponse(j *Edge) {
	if j.SE == Basic {
		j.SE = Rejected
	}
	n.Test()
}

func (n *Node) Report() {
	if n.findCount == 0 && n.testEdge == nil {
		n.SN = Found
		n.inBranch.Report(n.bestWt)
	}
}

func (n *Node) ReportResponse(w int, j *Edge, reqQ chan *Message, m Message) {
	if *j != *n.inBranch {
		n.findCount--
		if w < n.bestWt {
			n.bestWt = w
			n.bestEdge = j
		}
		n.Report()
	} else {
		if n.SN == Find {
			time.Sleep(time.Millisecond * 15)
			reqQ <- &m // place message end of Q
		} else if w > n.bestWt {
			n.ChangeCore()
		} else if w == Infinity && n.bestWt == Infinity {
			logger.Println("ALGORITHM HALTED!")
		}

	}
}

//ChangeCore procedure
func (n *Node) ChangeCore() {
	if n.bestEdge.SE == Branch {
		n.bestEdge.ChangeCore()
	} else {
		n.bestEdge.Connect(n.LN)
		n.bestEdge.SE = Branch
	}
}

func (n *Node) ChangeCoreResponse() {
	n.ChangeCore()
}

//Find the edge to the adj Node
func (n *Node) findEdge(an int) *Edge {

	for i, e := range *n.adjacencyList {
		if e.AdjNodeID == an {
			return &(*n.adjacencyList)[i]
		}
	}
	return nil
}

var (
	HostName string
	HostIP   string
	wakeup   bool
	ThisNode Node
	requests chan *Message
	logger   *log.Logger
)

func init() {
	HostName, HostIP = GetHostInfo()
	octets := strings.Split(HostIP, ".")
	fmt.Printf("My ID is: %s\n", octets[3])
	nodeID, err := strconv.Atoi(octets[3])
	edges, wd := GetEdgesFromFile("ghs.conf", nodeID)
	if err != nil {
		log.Fatal(err)
	}

	if wd == nodeID {
		wakeup = true
	}

	ThisNode = Node{
		SN:            Sleeping,
		ID:            nodeID,
		adjacencyList: &edges}

	logfile, err := os.Create("/logs/log" + strconv.Itoa(nodeID))
	if err != nil {
		log.Fatal(err)
	}
	logger = log.New(logfile, "logger: ", log.Lshortfile|log.Lmicroseconds)
	_ = logger

}

func GetHostInfo() (string, string) {
	HostIP, err := exec.Command("hostname", "-i").Output()
	if err != nil {
		log.Fatal(err)
	}
	HostIP = bytes.TrimSuffix(HostIP, []byte("\n"))

	HostName, err := exec.Command("hostname").Output()
	if err != nil {
		log.Fatal(err)
	}
	HostName = bytes.TrimSuffix(HostName, []byte("\n"))

	return string(HostName), string(HostIP)
}

func serveConn(c net.Conn, reqs chan *Message) {
	defer c.Close()
	var resp Message
	dec := gob.NewDecoder(c)
	err := dec.Decode(&resp)
	if err != nil {
		logger.Print(err)
	}

	reqs <- &resp
}

func processMessage(reqs chan *Message) {
	for m := range reqs {
		fmt.Printf("RECEIVED: %v from %v\n", m.Type, m.SourceID)
		j := ThisNode.findEdge(m.SourceID)
		switch {
		case m.Type == "Connect":
			logger.Println("ConnectResponse")
			ThisNode.ConnectResponse(m.L, j, reqs, *m)

		case m.Type == "Initiate":
			logger.Println("InitiateResponse")
			ThisNode.InitiateResponse(m.L, m.F, m.S, j)

		case m.Type == "Test":
			logger.Println("TestResponse")
			ThisNode.TestResponse(m.L, m.F, j, reqs, *m)

		case m.Type == "Reject":
			logger.Println("RejectResponse")
			ThisNode.RejectResponse(j)

		case m.Type == "Accept":
			logger.Println("AcceptResponse")
			ThisNode.AcceptResponse(j)

		case m.Type == "Report":
			logger.Println("ReportResponse")
			ThisNode.ReportResponse(m.W, j, reqs, *m)

		case m.Type == "ChangeCore":
			logger.Println("ChangeCoreResponse")
			ThisNode.ChangeCoreResponse()

		}
		if ThisNode.inBranch != nil {
			logger.Printf("STATUS: %v  INBRANCH: %v BESTWT: %v, LVL: %v\n", ThisNode.SN, ThisNode.inBranch.Weight, ThisNode.bestWt, ThisNode.LN)
		} else {
			logger.Printf("STATUS: %v  BESTWT: %v\n", ThisNode.SN, ThisNode.bestWt)
		}
		time.Sleep(time.Millisecond * 200)

	}
}

//The node will begin execution by creating a channel
//where requests will be queued for processing.
//Channels are thread-safe so multiple go routines can access
func main() {
	requests = make(chan *Message, 40)

	//Initialize Server
	notListening := make(chan bool)
	//log.Printf("STATUS: %v  INBRANCH: %v FCOUNT: %v", ThisNode.SN, (*ThisNode.inBranch).Weight, ThisNode.findCount)
	go func(nl chan bool) {
		defer func() {
			nl <- true
		}()
		l, err := net.Listen("tcp", PORT)
		fmt.Println("Listening")
		logger.Println("Listening")
		if err != nil {
			logger.Fatal(err)
		}

		for {
			conn, err := l.Accept()
			if err != nil {
				log.Fatal(err)
			}

			// Handle the connection in a new goroutine.
			go serveConn(conn, requests)
		}
	}(notListening)

	//Process incomming messages
	go processMessage(requests)

	if wakeup {
		time.Sleep(time.Second * 9)
		ThisNode.Wakeup()
	}

	//Wait until listening routine sends signal
	<-notListening
}
