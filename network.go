package main

import "fmt"

func logNode(n *Node, format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Printf("[%s %s] %s\n", n.Name, n.ID.String(), msg)
}
