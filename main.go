package main

import (
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

type Message struct {
	index int
	deps  []int
	value string
}

type Node struct {
	sendSeq   int
	delivered []int
	buffer    []Message
	id        int
	mu        sync.Mutex
	messages  chan Message
}

type Network struct {
	nodes []chan Message
}

func cmp_less(slice1, slice2 []int) bool {
	for i := range slice1 {
		if slice1[i] > slice2[i] {
			return false
		}
	}

	return true
}

func (n *Network) broadcast(msg Message) {
	for _, node := range n.nodes {
		node <- msg
	}

}

func (n *Node) receiveMessage() {
	for msg := range n.messages {
		n.mu.Lock()
		n.buffer = append(n.buffer, msg)

		for k := range n.buffer {
			if cmp_less(msg.deps, n.delivered) {
				log.Printf("[%d] Recebeu mensagem: %v de %d", n.id, msg.value, msg.index)
				n.delivered[msg.index]++
				n.buffer = n.buffer[:k+copy(n.buffer[k:], n.buffer[k+1:])]
			}
		}
		n.mu.Unlock()
	}
}

func (n *Node) sendMessage(msg string, net *Network) {

	n.mu.Lock()

	deps := make([]int, len(n.delivered))
	copy(deps, n.delivered)
	deps[n.id] = n.sendSeq

	message := Message{index: n.id, deps: deps, value: msg}
	go net.broadcast(message)

	n.sendSeq++
	n.mu.Unlock()
}

func (n *Node) init(numNodes int) {
	n.sendSeq = 0
	n.delivered = make([]int, numNodes)
	n.buffer = []Message{}
	n.mu = sync.Mutex{}
	n.messages = make(chan Message)

	go n.receiveMessage()
}

const NUMBER_OF_NODES = 3

func main() {

	n1 := Node{id: 0}
	n2 := Node{id: 1}
	n3 := Node{id: 2}

	n1.init(NUMBER_OF_NODES)
	n2.init(NUMBER_OF_NODES)
	n3.init(NUMBER_OF_NODES)

	network := Network{nodes: []chan Message{n1.messages, n2.messages, n3.messages}}

	n1.sendMessage("Valor1", &network)
	n2.sendMessage("Valor2", &network)
	n1.sendMessage("Valor3", &network)
	n3.sendMessage("Valor4", &network)
	n1.sendMessage("Valor5", &network)

	quitChannel := make(chan os.Signal, 1)
	signal.Notify(quitChannel, syscall.SIGINT, syscall.SIGTERM)
	<-quitChannel

}