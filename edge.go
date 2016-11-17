package main

import (
	"encoding/gob"
	"log"
	"net"
	"strconv"
)

//Edge States
const (
	Basic string = "Basic" //not yet decided whether the edge is part
	//of the MST or not
	Branch   string = "Branch"   //The edge is part of the MST
	Rejected string = "Rejected" //The edge is NOT part of the MST

)

//Edge is a weighted link between Nodes
type Edge struct {
	AdjNodeID int
	Weight    int    //Edge weight
	SE        string //Edge state

}

//Edges is a sortable edgelist
type Edges []Edge

func (e Edges) Len() int           { return len(e) }
func (e Edges) Swap(i, j int)      { e[i], e[j] = e[j], e[i] }
func (e Edges) Less(i, j int) bool { return e[i].Weight < e[j].Weight }

//Message
type Message struct {
	Type     string
	L        int
	F        int
	S        string
	W        int
	SourceID int
}

//SUBNET of docker network
const SUBNET string = "172.17.0."

//PORT
const PORT string = ":7575"

//Connect type message created and sent
func (e *Edge) Connect(l int) {
	m := Message{Type: "Connect", L: l, SourceID: ThisNode.ID}
	e.send(m)
	Logger.Printf("SENT-->: %v(L: %v) to %v\n", m.Type, m.L, e.AdjNodeID)

}

//Initiate type message created and sent
func (e *Edge) Initiate(l, f int, s string) {
	m := Message{Type: "Initiate", L: l, F: f, S: s, SourceID: ThisNode.ID}
	e.send(m)
	Logger.Printf("SENT-->: %v(L: %v, F: %v, S: %v) to %v\n", m.Type, m.L, m.F, m.S, e.AdjNodeID)

}

//Test type message created and sent
func (e *Edge) Test(l, f int) {
	m := Message{Type: "Test", L: l, F: f, SourceID: ThisNode.ID}
	e.send(m)
	Logger.Printf("SENT-->: %v(L: %v, F: %v) to %v\n", m.Type, m.L, m.F, e.AdjNodeID)

}

//Reject type message created and sent
func (e *Edge) Reject() {
	m := Message{Type: "Reject", SourceID: ThisNode.ID}
	e.send(m)
	Logger.Printf("SENT-->: %v() to %v\n", m.Type, e.AdjNodeID)

}

//Accept type message created and sent
func (e *Edge) Accept() {
	m := Message{Type: "Accept", SourceID: ThisNode.ID}
	e.send(m)
	Logger.Printf("SENT-->: %v() to %v\n", m.Type, e.AdjNodeID)

}

//Report type message created and sent
func (e *Edge) Report(we int) {
	m := Message{Type: "Report", W: we, SourceID: ThisNode.ID}
	e.send(m)
	Logger.Printf("SENT-->: %v(W: %v) to %v\n", m.Type, m.W, e.AdjNodeID)

}

//ChangeCore type message created and sent
func (e *Edge) ChangeCore() {
	m := Message{Type: "ChangeCore", SourceID: ThisNode.ID}
	e.send(m)
	Logger.Printf("SENT-->: %v() to %v\n", m.Type, e.AdjNodeID)

}

var enc *gob.Encoder

func (e *Edge) send(m Message) {
	conn, err := net.Dial("tcp", SUBNET+strconv.Itoa(e.AdjNodeID)+PORT)
	if err != nil {
		log.Println(err)
		log.Printf("conn null? %v\n", conn == nil)
	} else {
		enc = gob.NewEncoder(conn)
		err = enc.Encode(m)
		if err != nil {
			log.Fatal(err)
		}
	}
}
