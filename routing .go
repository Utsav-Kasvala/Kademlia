package main

import (
	"sort"
	"sync"
)

type RoutingTable struct {
	self    ID
	buckets []*Bucket
	mutex   sync.Mutex
}

func NewRoutingTable(self ID) *RoutingTable {
	//  number of prefix matching possible = IDBits
	rt := &RoutingTable{self: self, buckets: make([]*Bucket, IDBits)}
	for i := 0; i < IDBits; i++ {
		rt.buckets[i] = &Bucket{}
	}
	return rt
}

// Adding an entry in the routing table

func (rt *RoutingTable) Update(c Contact) {
	// do not enter self id in the routing table
	if c.ID == rt.self {
		return
	}
	// finding bucket index
	idx := bucketIndex(rt.self, c.ID)
	if idx < 0 || idx >= IDBits {
		return
	}
	//adding the entry in the routing table
	rt.mutex.Lock()
	defer rt.mutex.Unlock()
	rt.buckets[idx].Update(c)
}

// get all contacts from the routing table
func (rt *RoutingTable) GetAllContacts() []Contact {
	rt.mutex.Lock()
	defer rt.mutex.Unlock()
	var out []Contact
	for _, b := range rt.buckets {
		out = append(out, b.GetContacts()...)
	}
	return out
}

// Find count closest contacts from a particular ID  from the given routing table

func (rt *RoutingTable) FindClosest(target ID, count int) []Contact {
	all := rt.GetAllContacts()
	type pair struct {
		c Contact
		d ID
	}
	pairs := []pair{}
	for _, c := range all {
		pairs = append(pairs, pair{c: c, d: target.Distance(c.ID)})
	}
	sort.Slice(pairs, func(i, j int) bool { return pairs[i].d < pairs[j].d })
	out := []Contact{}
	for i := 0; i < len(pairs) && i < count; i++ {
		out = append(out, pairs[i].c)
	}
	return out
}
