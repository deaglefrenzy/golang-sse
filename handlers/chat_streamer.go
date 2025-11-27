package handlers

import "sync"

type Streamer struct {
	clients []chan []byte // uses []byte to send raw message
	mu      sync.Mutex    // to keep only 1 goroutine can access the array of channels
}

func NewStreamer() *Streamer {
	return &Streamer{clients: make([]chan []byte, 0)}
}

func (b *Streamer) AddClient(ch chan []byte) { // add client to list
	b.mu.Lock()
	b.clients = append(b.clients, ch)
	b.mu.Unlock()
}

func (b *Streamer) RemoveClient(ch chan []byte) { // remove client from list
	b.mu.Lock()
	defer b.mu.Unlock()

	for i, c := range b.clients {
		if c == ch {
			b.clients = append(b.clients[:i], b.clients[i+1:]...) // :i = all items before i, i+1: = all items after i+1
			return
		}
	}
}

func (b *Streamer) Stream(msg []byte) {
	b.mu.Lock()
	defer b.mu.Unlock()

	for _, ch := range b.clients {
		select { // made into select to make sure the channel still have free space and avoid blocking
		case ch <- msg: // sent successfully
		default:
			// channel is full. skip this client
		}
	}
}
