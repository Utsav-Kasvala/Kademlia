package main

import "fmt"

func main() {
	// Create a small network
	n1 := NewNode(RandomID())
	n2 := NewNode(RandomID())
	n3 := NewNode(RandomID())

	// Bootstrap nodes
	n2.Bootstrap(n1)
	n3.Bootstrap(n1)

	// Store a key/value
	key := "lop"
	value := "world"
	closest := n1.IterativeFindNode(IDFromKey(key))
	for _, c := range closest {
		c.Node.StoreRPC(Contact{ID: n1.ID, Node: n1}, key, value)
	}

	// Iterative value lookup from another node
	val, ok := n3.IterativeFindValue(key)
	fmt.Println("Lookup from n3:", val, ok)
}
