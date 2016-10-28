package ghsmst

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"log"
	"net"
	"os/exec"
	"strconv"
	"strings"

	"github.com/parpat/ghsmst"
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
	LN int    //

	ID int

	bestEdge      Edge
	bestWt        int
	testEdge      Edge
	inBranch      Edge
	findCount     int
	adjacencyList Edges
}

//Wakeup initializes the node
func (n *Node) Wakeup() {
	n.adjacencyList[0].SE = Branch
	n.LN = 0
	n.SN = Found
	n.findCount = 0

	n.adjacencyList[0].Connect(0)
}

//ConnectResponse responds to the Connect message
func (n *Node) ConnectResponse(L int, j Edge) {
	if n.SN == Sleeping {
		n.Wakeup()
	}

	if L < n.LN {
		j.SE = Branch
		j.Initiate(n.LN, n.FN, n.SN)
		if n.SN == Find {
			n.findCount++
		}
		}else if j.SE == Basic {
			//place received message in the end of Q

		} else {
			j.Initiate(n.LN+1, j.Weight, Find)
		}

}

//InitiateResponse responds to the Initiate message
func (n *Node) InitiateResponse(L, F int, S string, j Edge) {
	n.LN = L
	n.FN = F
	n.SN = S
	n.inBranch = j

	n.bestEdge = nil//Edge{0, 0, ""}
	n.bestWt = Infinity

	for i := range n.adjacencyList{
		if i!= j && i.SE == Branch{
			Edge(i).Initiate(L, F, S)
			if S==Find{
				n.findCount++
			}
		}
	}

	if S==Find{
		n.Test()
	}
}

//Test picks the minimum Basic Edge and send test message
func (n *Node) Test() {
	// if there are adjacent Edges in state Basic{
	// 	n.testEdge = min weight adjacent edge with state Basic
	// 	send Test(n.LN, n.FN) on n.testEdge
	// }else{
	// 	n.testEdge = nil
	// 	Report()
	// }
}


//TestResponse responds to Test message
func (n *Node) TestResponse(l, f int, j Edge){

}


func (n *Node) AcceptResponse(j Edge){

}



func (n *Node) RejectResponse(j Edge){

}

func (n *Node) Report(){

}

func (n *Node) ReportResponse(w int, j Edge){

}

func (n *Node) ChangeCore(){

}

func (n *Node) ChangeCoreResponse(){
	n.ChangeCore()
}

//Find the edge to the adj Node
func (n *Node) findEdge(an int) Edge{
	for _, e := range n.adjacencyList{
		if e.AdjNodeID == an{
			return e
		}
	}
}




var (
	HostName string
	HostIP   string
	requests chan string
	wakeup   int
	ThisNode Node
	requests chan Message
)

func init() {
	HostName, HostIP = GetHostInfo()
	octets := strings.Split(HostIP, ".")
	fmt.Printf("My ID is: %s\n", octets[3])
	nodeID, err := strconv.Atoi(octets[3])
	edges, wakeup = GetEdgesFromFile("ghs.conf", nodeID)
	if err != nil {
		log.Fatal(err)
	}

	ThisNode = Node{
		SN: Sleeping,
		ID: nodeID
		adjacencyList: edges

	}

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
	var resp *Message
	dec = gob.NewDecoder(c)
	err := dec.Decode(resp)
	if err != nil {
		log.Print(err)
	}

	reqs <- resp
}

func processMessage(reqs chan *Message){
	for m := range reqs{
		j := 			ThisNode.findEdge(m.SourceID)
		switch {
		case m.Type = "Connect":
			ThisNode.ConnectResponse(m.L, j)

		case m.Type = "Initiate":
			ThisNode.InitiateResponse(m.L, m.F, m.S, j)

		case m.Type = "Test":
			ThisNode.TestResponse(m.L, m.F, j)

		case m.Type = "Reject":
			ThisNode.RejectResponse(j)

		case m.Type = "Accept":
			ThisNode.AcceptResponse(j)

		case m.Type = "Report":
			ThisNode.ReportResponse(m.W, j)

		case m.Type = "ChangeCore":
			ThisNode.ChangeCoreResponse()


		}
	}

}



func main() {
	requests = make(chan *Message, 10)

	//Initialize Server
	notListening := make(chan bool)

	go func(nl chan bool) {
		defer nl <- true
		l, err := net.Listen("tcp", PORT)
		fmt.Println("Listening")
		if err != nil {
			log.Fatal(err)
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

//Wait until listening routine sends signal
<-notListening
}
