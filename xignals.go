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


type Xignal struct {
	eventCounter int
	channel chan Event
	mutex sync.Mutex
	closed bool
	filter Filter
	while Filter
	action func(Event) error
}

// NewSignal() Signaler
func NewSignal() *Xignal {

	signal:= &Xignal{
		channel : make(chan Event),
	}
	signal.filter = Always
	signal.while = Always
	return signal
}

// Publish
func (s *Xignal) Publish(payload interface{}) error {
	if s.IsClosed() {
		return ErrChannelClosed
	}

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

	if s.IsClosed() {
		return ErrChannelClosed
	}else {
		s.channel <- MakeEvent(payload, completed, err)
		return nil
	}
}

func (s *Xignal) Complete() error {
	var err error
	if s.IsClosed(){
		err = ErrChannelClosed
		return err
	}
	s.channel <- MakeEvent(nil, true, err)
	return err
}


// Subscribe to event
// TODO: add subscription to actions?, subscriptions ?  and start monitoring ? instead of exec action
func (this *Xignal) Subscribe(action func(e Event) error ) *Xignal {
	this.action = action
	return this
}

func (this *Xignal) GO() error {
	var err error
	for event := range (this.channel) {

		// IsFaulted then close channel
		if event.IsFaulted() {
			this.Close()
			err = event.GetError()
			// No Need
			break
		}

		// Event stream completed then close channel
		if event.IsCompleted() {
			log.Printf("COMPLETED: %v", event)
			this.Close()
			break
		}

		if !this.while(event) {
			log.Printf("INFO: signal while not met: %v", event)
			this.Complete()
			break
		}

		// Wait for arbitrary condition
		if this.filter(event) {
			log.Printf("ACTION: event : %v", event)
			err = this.action(event)
			if err!=nil {
				break
			}
			continue
		}

		log.Printf("SKIP: %v" , event)
		// Next
	}

	this.Close()

	return err
}

// ErrChannelClosed
var ErrChannelClosed = errors.New("Already closed")


// Closed flag
func (s *Xignal) Close() error {
	if s.IsClosed() {
		return ErrChannelClosed
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.closed = true
	return nil
}

func (x *Xignal) IsClosed() bool {
	x.mutex.Lock()
	defer x.mutex.Unlock()
	return x.closed
}

// Short for func(e Event) bool
type Filter func(e Event) bool;

// Short for func(e Event) bool { return true }
var Always Filter = func(Event) bool { return true }

// When this Filter returns true, subscription execs action
func (s *Xignal) When(f Filter) *Xignal {
	s.filter = f
	return s
}

// While this filter is true , subscription execs action , if <while> is NOT met channel is Closed
func (s *Xignal) While(f Filter) *Xignal {
	s.while = f
	return s
}