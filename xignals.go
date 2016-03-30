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
	Subsciptions []*Subscription
}

// NewSignal() Signaler
func NewSignal() *Xignal {
	signal:= &Xignal{
		channel : make(chan Event),
		Subsciptions: make([]*Subscription,0),
	}
	return signal
}

// Publish
func (s *Xignal) Publish(payload interface{}) error {

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
	return nil
}

func (s *Xignal) Complete() error {
	// and Now? , no catch
	var err error
	s.channel <- MakeEvent(nil, true, err)
	return err
}

func (this *Xignal) GO() error {

	var err error

	for event := range (this.channel) {

		// IsFaulted then close channel
		if event.IsFaulted() {
			err = event.GetError()
			log.Printf("Xignal: event(%v) ERROR: %s", event, err.Error())
			// No Need
			break
		}

		// Event stream completed then close channel
		if event.IsCompleted() {
			log.Printf("Xignal: COMPLETED: %v", event)
			break
		}

		// Loop subscriptions
		for i, s:= range this.Subsciptions[:] {

			if s.completed(event) {
				log.Printf("Xignal: INFO: Subscription Completed when : %v", event)
				this.Subsciptions[i] = nil
				continue
			}

			if s.when(event) {
				log.Printf("Xignal: INFO: subscription action(event: %v)", event)

				err = s.action(event)

				// onError remove
				// TODO: s.OnError()
				if err!=nil {
					this.Subsciptions[i] = nil
				}
				continue
			}
			log.Printf("Xignal: Subscription Condition Not met , Skipped")
		}
		// Remove nils
		notNil:= make([]*Subscription,0)
		for _,s := range this.Subsciptions[:]{
			if s != nil {
				notNil = append(notNil, s)
			}
		}
		this.Subsciptions = notNil

		log.Printf("Xignal: NEXT: event %v" , event)
		// Next
	}

	close(this.channel)
	return err
}

func (x *Xignal) Work(worker func(x *Xignal)){

	go func() {
		worker(x)
	}()

}

// Short for func(e Event) bool
type Filter func(e Event) bool;

// Short for func(e Event) bool { return true }
var Always Filter = func(Event) bool { return true }

// Short for func(e Event) bool { return false }
var Never Filter = func(Event) bool { return false }

// When this Filter returns true, subscription execs action
func (this *Xignal) When(f Filter) *Subscription {
	s:= NewSubscription()
	s.xignal = this
	s.when = f
	return s
}

// While this func return false, subscription runs  , if true , subscription is completed , and will be removed
// completes subscription cycle NOT event Cycle
// publisher publish event regardless of subscribers
func (s *Subscription) CompleteWhen(f Filter) *Subscription {
	s.completed = f
	return s
}


// Subscribe to event
// TODO: add subscription
func (this *Subscription) Subscribe(action func(e Event) error ) *Subscription {
	this.action = action
	this.xignal.Subsciptions = append(this.xignal.Subsciptions, this )
	return this
}

type Subscription struct {
	xignal    *Xignal
	when      Filter
	completed Filter
	action    func(e Event) error
}

var Nothing = func(e Event) error {
	return nil
}

func NewSubscription() *Subscription{
	s:= &Subscription{
		completed: Never,
		when: Always,
		action: Nothing,
	}
	return s
}