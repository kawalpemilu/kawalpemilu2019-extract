package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
)

const (
	ApiRoot  = "https://kawal-c1.appspot.com/api/c/"
	TpsDepth = 4
	DataDir  = "c"
)

func main() {
	q := make([]int, 0)
	q = append(q, 0)

	for len(q) > 0 {
		id := q[0]
		q = q[1:]

		node, err := readOrGetNode(id)
		if err != nil {
			log.Fatalf("Could not get the node %d: %v", id, err)
		}

		log.Printf("Node %d Depth %d", node.Id, node.Depth)
		if node.Depth >= TpsDepth {
			continue
		}

		var wg sync.WaitGroup

		for _, ch := range node.Children {
			cid := ch.Id()
			q = append(q, cid)

			wg.Add(1)
			go func(id int) {
				readOrGetNode(id)
				wg.Done()
			}(cid)
		}

		wg.Wait()
	}
}

func readOrGet(id int) (io.ReadCloser, error) {
	path := fmt.Sprintf("%s/%d.json", DataDir, id)

	if _, err := os.Stat(path); os.IsNotExist(err) {
		url := fmt.Sprintf("%s%d", ApiRoot, id)
		resp, err := http.Get(url)
		if err != nil {
			return nil, err
		}

		defer resp.Body.Close()

		tmp := fmt.Sprintf("%s.tmp", path)
		f, err := os.Create(tmp)
		if err != nil {
			return nil, err
		}

		_, err = io.Copy(f, resp.Body)
		if err != nil {
			return nil, err
		}
		f.Close()

		err = os.Rename(tmp, path)
		if err != nil {
			return nil, err
		}
	}

	return os.Open(path)
}

func readOrGetNode(id int) (Node, error) {
	var node Node

	b, err := readOrGet(id)
	if err != nil {
		return node, err
	}

	defer b.Close()

	dec := json.NewDecoder(b)
	err = dec.Decode(&node)
	return node, err
}

type Node struct {
	Id       int     `json:"id"`
	Depth    int     `json:"depth"`
	Children []Child `json:"children"`
}

type Child []interface{}

func (c Child) Id() int {
	f := c[0].(float64)
	return int(f)
}
