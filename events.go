package signals

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
}

// NewSignal() Signaler
func NewSignal() *Signal{

	signal:= &Signal {
		channel : make(chan Event),
	}
	return signal
}

// Publish
func (s *Signal) Publish(payload interface{}) {

	s.eventCounter++

	channelLength := len(s.channel)

	completed:= s.eventCounter >= channelLength

	var err error
	if  channelLength !=0 && s.eventCounter > channelLength {
		err = errors.New("Error")
	}

	s.channel <- MakeEvent(payload, completed, err)
}

// Subscribe to event
func (s *Signal)Subscribe(action func(e Event), condition func(e Event) bool){

	for event := range (s.channel) {

		// IsFaulted then close channel
		if event.IsFaulted() {
			log.Printf("Error: %s", event.err.Error())
			// No Need
			s.Close()
		}

		// Event stream completed then close channel
		if event.IsCompleted() {
			log.Printf("INFO: event sequence completed")
			s.Close()
			return
		}

		// Wait for arbitrary condition , value as 3(int) ?, finish early
		if condition(event) {
			action(event)
			return
		}

		log.Printf("Idle: %v" , event)
		// Next
	}
}

// Close inner channel
func (s *Signal) Close() error {
	s.mutex.Lock()
	if s.closed {
		return errors.New("Already closed")
	}
	s.closed = true
	s.mutex.Unlock()
	close(s.channel)
	return nil
}