package main

import (
	"math"
	"sort"
	"sync"
	"time"
)

const RpcDelay = 10 * time.Millisecond

type Node struct {
	ID   ID
	Name string

	rt   *RoutingTable
	data map[string]string
}

// creating a new node
func NewNode(id ID) *Node {
	return &Node{
		ID:   id,
		Name: "N" + id.String(),
		rt:   NewRoutingTable(id),
		data: make(map[string]string),
	}
}

// ---------- RPCs ----------
func (n *Node) Ping(from Contact) bool {
	time.Sleep(RpcDelay)
	n.rt.Update(from)
	logNode(n, "Received Ping from %s", from.Node.Name)
	return true
}

func (n *Node) FindNodeRPC(from Contact, target ID) []Contact {
	time.Sleep(RpcDelay)
	n.rt.Update(from)
	found := n.rt.FindClosest(target, K)
	logNode(n, "FindNode(from=%s, target=%s) -> %d contacts", from.Node.Name, target, len(found))
	return found
}

func (n *Node) StoreRPC(from Contact, key string, value string) {
	time.Sleep(RpcDelay)
	n.rt.Update(from)
	n.data[key] = value
	logNode(n, "Stored (key=%s, val=%s) from %s", key, value, from.Node.Name)
}

func (n *Node) FindValueRPC(from Contact, key string) (string, bool, []Contact) {
	time.Sleep(RpcDelay)
	n.rt.Update(from)
	if val, ok := n.data[key]; ok {
		logNode(n, "FindValue: FOUND key=%s", key)
		return val, true, nil
	}
	closest := n.rt.FindClosest(IDFromKey(key), K)
	logNode(n, "FindValue: NOT found key=%s -> returning %d contacts", key, len(closest))
	return "", false, closest
}

// ---------- Bootstrap ----------
func (n *Node) Bootstrap(other *Node) {
	other.Ping(Contact{ID: n.ID, Node: n})
	n.rt.Update(Contact{ID: other.ID, Node: other})
	logNode(n, "Bootstrapped with %s", other.Name)
	n.IterativeFindNode(n.ID) // populate table
}

// ---------- Iterative Store ----------
func (n *Node) IterativeStore(key, value string) {
	target := IDFromKey(key)
	closest := n.IterativeFindNode(target)

	for _, c := range closest {
		c.Node.StoreRPC(Contact{ID: n.ID, Node: n}, key, value)
	}
	logNode(n, "IterativeStore DONE key=%s -> stored on %d nodes", key, len(closest))
}

// ---------- Iterative Find ----------
func (n *Node) IterativeFindNode(target ID) []Contact {
	logNode(n, "IterativeFindNode START target=%s", target)

	shortlist := n.rt.FindClosest(target, K)
	if len(shortlist) == 0 {
		logNode(n, "No contacts known, abort")
		return nil
	}

	type entry struct {
		c       Contact
		queried bool
	}
	entries := map[ID]*entry{}
	for _, c := range shortlist {
		entries[c.ID] = &entry{c: c}
	}

	closestBefore := ID(math.MaxUint16)
	for {
		cands := []Contact{}
		for _, e := range entries {
			if !e.queried {
				cands = append(cands, e.c)
			}
		}
		if len(cands) == 0 {
			break
		}
		sort.Slice(cands, func(i, j int) bool {
			return target.Distance(cands[i].ID) < target.Distance(cands[j].ID)
		})
		toQuery := cands
		if len(toQuery) > Alpha {
			toQuery = toQuery[:Alpha]
		}

		var wg sync.WaitGroup
		resCh := make(chan []Contact, len(toQuery))

		for _, c := range toQuery {
			entries[c.ID].queried = true
			wg.Add(1)
			go func(ct Contact) {
				defer wg.Done()
				nodes := ct.Node.FindNodeRPC(Contact{ID: n.ID, Node: n}, target)
				resCh <- nodes
			}(c)
		}

		wg.Wait()
		close(resCh)

		progress := false
		for nodes := range resCh {
			for _, nc := range nodes {
				n.rt.Update(nc)
				if _, ok := entries[nc.ID]; !ok {
					entries[nc.ID] = &entry{c: nc}
					progress = true
				}
			}
		}

		best := ID(math.MaxUint16)
		for id := range entries {
			d := target.Distance(id)
			if d < best {
				best = d
			}
		}
		if best >= closestBefore && !progress {
			break
		}
		closestBefore = best
	}

	all := []Contact{}
	for _, e := range entries {
		all = append(all, e.c)
	}
	sort.Slice(all, func(i, j int) bool {
		return target.Distance(all[i].ID) < target.Distance(all[j].ID)
	})
	if len(all) > K {
		all = all[:K]
	}
	logNode(n, "IterativeFindNode DONE -> %d contacts", len(all))
	return all
}

// ---------- Iterative FindValue ----------
func (n *Node) IterativeFindValue(key string) (string, bool) {
	target := IDFromKey(key)

	shortlist := n.rt.FindClosest(target, K)
	if len(shortlist) == 0 {
		return "", false
	}

	type entry struct {
		c       Contact
		queried bool
	}
	entries := map[ID]*entry{}
	for _, c := range shortlist {
		entries[c.ID] = &entry{c: c}
	}

	for {
		cands := []Contact{}
		for _, e := range entries {
			if !e.queried {
				cands = append(cands, e.c)
			}
		}
		if len(cands) == 0 {
			break
		}
		sort.Slice(cands, func(i, j int) bool {
			return target.Distance(cands[i].ID) < target.Distance(cands[j].ID)
		})
		toQuery := cands
		if len(toQuery) > Alpha {
			toQuery = toQuery[:Alpha]
		}

		progress := false
		for _, c := range toQuery {
			entries[c.ID].queried = true
			val, ok, contacts := c.Node.FindValueRPC(Contact{ID: n.ID, Node: n}, key)
			if ok {
				return val, true
			}
			for _, nc := range contacts {
				if _, exists := entries[nc.ID]; !exists {
					entries[nc.ID] = &entry{c: nc}
					progress = true
				}
			}
		}
		if !progress {
			break
		}
	}
	return "", false
}
