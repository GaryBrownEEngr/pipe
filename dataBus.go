package pipe

import "sync"

/*
This is similar to the data broker, but we shouldn't need to spin up a separate go routine.
There will only be a single publisher, so that publisher should be able to perform the processing when a publish is desired.

*/

type WireSource[T any] struct {
	on         bool
	bufferSize int
	setupMutex sync.Mutex
	subs       map[chan []T]struct{}
}

// Creates a new DataBus of any given type and start it running.
func NewWire[T any](bufferSize int) *WireSource[T] {
	ret := &WireSource[T]{
		on:         true,
		bufferSize: bufferSize,
		subs:       make(map[chan []T]struct{}),
	}

	return ret
}

func (s *WireSource[T]) GetWire() *Wire[T] {
	ret := &Wire[T]{
		dataBus: s,
	}
	return ret
}

func (s *WireSource[T]) Publish(msg []T) {
	s.setupMutex.Lock()
	defer s.setupMutex.Unlock()

	if !s.on {
		panic("Trying to publish to dataBus after stopping it")
	}

	for msgCh := range s.subs {
		msgCh <- msg
	}
}

// Stop the DataBus. Only call this once. All subscriptions will be closed
func (s *WireSource[T]) Stop() {
	s.setupMutex.Lock()
	defer s.setupMutex.Unlock()

	if !s.on {
		panic("Trying to stop to dataBus after already stopping it")
	}

	s.on = false
	for msgCh := range s.subs {
		close(msgCh)
	}
}

// Subscribe to the DataBus. Messages can be received on the returned channel.
func (s *WireSource[T]) Subscribe() chan []T {
	s.setupMutex.Lock()
	defer s.setupMutex.Unlock()

	if !s.on {
		return nil
	}

	msgCh := make(chan []T, s.bufferSize)
	s.subs[msgCh] = struct{}{}
	return msgCh
}

// Unsubscribe from the DataBus. The reference to the channel can be safely discarded after calling this.
func (s *WireSource[T]) Unsubscribe(msgCh chan []T) {
	s.setupMutex.Lock()
	defer s.setupMutex.Unlock()

	if _, found := s.subs[msgCh]; !found {
		return
	}
	if s.on {
		close(msgCh)
	}
	delete(s.subs, msgCh)

	go drainChannelTillClosed(msgCh)
}

/////////////////////////////
/////////////////////////////
/////////////////////////////

type Wire[T any] struct {
	dataBus *WireSource[T]
}

func (s *Wire[T]) NewWireEnd() *WireEnd[T] {
	return &WireEnd[T]{
		dataBus: s.dataBus,
		msgCh:   s.dataBus.Subscribe(),
	}
}

type WireEnd[T any] struct {
	dataBus *WireSource[T]
	msgCh   chan []T
}

func (s *WireEnd[T]) GetData() ([]T, bool) {
	ret, ok := <-s.msgCh
	return ret, ok
}

func (s *WireEnd[T]) Disconnect() {
	s.dataBus.Unsubscribe(s.msgCh)
}

func drainChannelTillClosed[T any](in chan T) {
	for {
		_, ok := <-in
		if !ok {
			return
		}
	}

}
