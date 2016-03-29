package xignals

import (
	"log"
	"errors"
	"sync"
)


// Event or kind of ...
// is read only
type Event struct {
	payload   interface{}
	err       error
	completed bool
}

// IsFaulted event has Error
func (event Event) IsFaulted() bool {
	return event.err != nil
}

// IsCompleted
func (event Event) IsCompleted() bool {
	// or event.status == status.Completed ?
	return event.completed
}

// GetPayload
func (event Event) GetPayload() interface{} {
	return event.payload
}

// GetError
func (event Event) GetError() error {
	return event.err
}

func MakeEvent(payload interface{}, completed bool, e error) Event {
	return Event{payload: payload, completed: completed, err: e}
}


type Signal struct {
	eventCounter int
	channel chan Event
	mutex sync.Mutex
	closed bool
	filter Filter
	while Filter
}

// NewSignal() Signaler
func NewSignal() *Signal{

	signal:= &Signal {
		channel : make(chan Event),
	}
	signal.filter = Always
	signal.while = Always
	return signal
}

// Publish
func (s *Signal) Publish(payload interface{}) {

	s.eventCounter++

	channelLength := len(s.channel)

	completed:= false
	if channelLength != 0 {
		completed = s.eventCounter >= channelLength
	}

	var err error
	if  channelLength !=0 && s.eventCounter > channelLength {
		err = errors.New("Error")
	}

	s.channel <- MakeEvent(payload, completed, err)
}

func (s *Signal) Complete(err error) {
	s.channel <- MakeEvent(nil, true, err)
}



// Subscribe to event
// TODO: add subscription to actions?, subscriptions ?  and start monitoring ? instead of exec action
func (s *Signal)Subscribe(action func(e Event)){

	for event := range (s.channel) {

		// IsFaulted then close channel
		if event.IsFaulted() {
			log.Printf("Error: %s", event.err.Error())
			// No Need
			s.Close()
		}

		// Event stream completed then close channel
		if event.IsCompleted() {
			log.Printf("INFO: event completed: %v", event)
			s.Close()
			return
		}

		if !s.while(event) {
			log.Printf("INFO: signal while not met: %v", event)
			s.Close()
			return
		}

		// Wait for arbitrary condition
		if s.filter(event) {
			log.Printf("ACTION: event : %v", event)
			action(event)
			return
		}

		log.Printf("Idle: %v" , event)
		// Next
	}
}

// ErrChannelClosed
var ErrChannelClosed = errors.New("Already closed")


// Close inner channel
func (s *Signal) Close() error {
	s.mutex.Lock()
	if s.closed {
		return ErrChannelClosed
	}
	s.closed = true
	s.mutex.Unlock()
	close(s.channel)
	return nil
}

// Short for func(e Event) bool
type Filter func(e Event) bool;

// Short for func(e Event) bool { return true }
var Always Filter = func(Event) bool { return true }

// When this Filter returns true, subscription execs action
func (s *Signal) When(f Filter) *Signal{
	s.filter = f
	return s
}

// While this filter is true , subscription execs action , if <while> is NOT met channel is Closed
func (s *Signal) While(f Filter) *Signal{
	s.filter = f
	return s
}