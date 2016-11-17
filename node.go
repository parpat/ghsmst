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
		Logger.Printf("PUT BACK!!! L:%v\n", L)
		time.Sleep(time.Millisecond * 3)
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
			Logger.Printf("basic edge:%v\n", e.Weight)
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
		time.Sleep(time.Millisecond * 150)
		//Logger.Printf("PUT BACK!!! L:%v F:%v", l, f)
		reqQ <- &m //Put message end of Q ***************
	} else if f != n.FN {
		// j.SE = Branch
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
	} else if n.SN == Find {
		time.Sleep(time.Millisecond * 1)
		reqQ <- &m // place message end of Q
	} else if w > n.bestWt {
		n.ChangeCore()
	} else if w == Infinity && n.bestWt == Infinity {
		Logger.Println("ALGORITHM HALTED!")
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
	Logger   *log.Logger
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
	Logger = log.New(logfile, "logger: ", log.Lshortfile|log.Lmicroseconds)
	_ = Logger

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
		Logger.Print(err)
	}

	reqs <- &resp
}

func processMessage(reqs chan *Message) {
	for m := range reqs {
		Logger.Printf("RECEIVED<--: %v from %v (L: %v, F: %v, W: %v)\n", m.Type, m.SourceID, m.L, m.F, m.W)
		j := ThisNode.findEdge(m.SourceID)
		switch {
		case m.Type == "Connect":
			//Logger.Println("ConnectResponse")
			ThisNode.ConnectResponse(m.L, j, reqs, *m)

		case m.Type == "Initiate":
			//Logger.Println("InitiateResponse")
			ThisNode.InitiateResponse(m.L, m.F, m.S, j)

		case m.Type == "Test":
			//Logger.Println("TestResponse")
			ThisNode.TestResponse(m.L, m.F, j, reqs, *m)

		case m.Type == "Reject":
			//Logger.Println("RejectResponse")
			ThisNode.RejectResponse(j)

		case m.Type == "Accept":
			//Logger.Println("AcceptResponse")
			ThisNode.AcceptResponse(j)

		case m.Type == "Report":
			//Logger.Println("ReportResponse")
			ThisNode.ReportResponse(m.W, j, reqs, *m)

		case m.Type == "ChangeCore":
			//Logger.Println("ChangeCoreResponse")
			ThisNode.ChangeCoreResponse()

		}
		if ThisNode.inBranch != nil {
			Logger.Printf("SN: %v  INBRANCH: %v BESTWT: %v, LN: %v FN: %v\n", ThisNode.SN, ThisNode.inBranch.Weight, ThisNode.bestWt, ThisNode.LN, ThisNode.FN)
		} else {
			Logger.Printf("SN: %v  BESTWT: %v\n", ThisNode.SN, ThisNode.bestWt)
		}
		time.Sleep(time.Millisecond * 800)

	}
}

//The node will begin execution by creating a channel
//where requests will be queued for processing.
//Channels are thread-safe so multiple go routines can access
func main() {
	requests = make(chan *Message, 50)

	//Initialize Server
	notListening := make(chan bool)
	//log.Printf("STATUS: %v  INBRANCH: %v FCOUNT: %v", ThisNode.SN, (*ThisNode.inBranch).Weight, ThisNode.findCount)
	go func(nl chan bool) {
		defer func() {
			nl <- true
		}()
		l, err := net.Listen("tcp", PORT)
		fmt.Println("Listening")
		Logger.Println("Listening")
		if err != nil {
			Logger.Fatal(err)
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
		time.Sleep(time.Second * 11)
		ThisNode.Wakeup()
	}

	//Wait until listening routine sends signal
	<-notListening
}
