package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"sort"
	"strconv"
	"strings"
)

//GetEdgesFromFile config
func GetEdgesFromFile(fname string, id int) (Edges, int) {
	rawContent, err := ioutil.ReadFile(fname)
	if err != nil {
		log.Fatal(err)
	}

	strs := strings.Split(string(rawContent), "\n")

	wakeup, err := strconv.Atoi(strs[0])
	if err != nil {
		log.Print(err)
	}
	fmt.Printf("wakeup: %v\n", wakeup)

	strs = strs[1:]
	var myedges Edges
	for _, edge := range strs {
		if edge != "" {
			vals := strings.Split(edge, " ")
			s, _ := strconv.Atoi(vals[0])
			d, _ := strconv.Atoi(vals[1])
			w, _ := strconv.Atoi(vals[2])
			if s == id {
				myedges = append(myedges, Edge{SE: "Basic", Weight: w, AdjNodeID: d})
			} else if d == id {
				myedges = append(myedges, Edge{SE: "Basic", Weight: w, AdjNodeID: s})
			}
			fmt.Printf("My edge %v\n", myedges)
		}
	}

	sort.Sort(myedges)

	return myedges, wakeup
}
