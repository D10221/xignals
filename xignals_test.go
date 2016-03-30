package xignals_test

import (
	"testing"
	"time"
	"log"
	"github.com/D10221/xignals"
	"errors"
)


func Test_Signal(t *testing.T) {

	x := xignals.NewSignal()

	condition := func(e xignals.Event) bool {
		canRun:= false
		if value, ok := e.GetPayload().(int); ok && value > 0 && value < 4 {
			canRun = true
		}
		// accepts 1,2,3
		return canRun
	}

	workerActionCount := 0
	action := func(e xignals.Event) error {
		// Validate payload
		if _, ok := e.GetPayload().(int); !ok {
			return errors.New("Invalid Payload")
		}
		// work
		workerActionCount++
		log.Printf("Worker action count: %v", e.GetPayload())
		return nil
	}


	isCompleted := func(e xignals.Event) bool {
		value , ok := e.GetPayload().(int);
		// Some arbitrary test , Exit Early
		completed := ok && value == 4
		return completed
	}
	event_count := 0
	// ...
	worker := func( x *xignals.Xignal ) {
		for i := 0; i < 5; i++ {
			// Heavy Work
			time.Sleep(time.Millisecond * 500)

			// Notify
			e:= x.Publish(i)
			// Stop Pubishing onError
			if e!=nil {
				log.Printf("Error: %v", e)
				break
			}
			// testing purposes
			event_count++
		}
		// Complete event cycle , Not Subscription Cycle
		e:= x.Complete()
		//
		if e!=nil {
			log.Printf("Error: %v", e)
		}
	}

	//TODO:  Need to get subscriptions before worker starts publishing
	// Currently it doesn't work
	//  Start working
	x.Work(worker) // ...

	// Add Subscriptions
	x.When(condition).CompleteWhen(isCompleted).Subscribe(action)
	x.GO()

	//time.Sleep(time.Millisecond * 500 * 6 )

	expected:= 3
	if workerActionCount != expected {
		t.Error("Worker action count : Expected %v, Got: %v", expected, workerActionCount)
	}
	if event_count != 5 {
		t.Error("event_count: Expected %v, Got: %v", 5, workerActionCount)
	}

}
