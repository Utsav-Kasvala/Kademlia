package main

type Bucket struct {
	contacts []Contact
}

func (b *Bucket) Update(c Contact) {
	// if present , send it to last
	for i, x := range b.contacts {
		if x.ID == c.ID {
			copy(b.contacts[i:], b.contacts[i+1:])
			b.contacts[len(b.contacts)-1] = c
			return
		}
	}
	// if size less than k, add at last
	if len(b.contacts) < K {
		b.contacts = append(b.contacts, c)
		return
	}
	// removes the first element
	b.contacts = append(b.contacts[1:], c)
}

func (b *Bucket) GetContacts() []Contact {
	out := make([]Contact, len(b.contacts))
	copy(out, b.contacts)
	return out
}
