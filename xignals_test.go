package xignals_test

import (
	"testing"
	//"sync"
	"time"
	"log"
	"github.com/D10221/xignals"
)

type Subscrber struct {

}

type Publisher struct {
	Xignal *xignals.Xignal
}

func (p *Publisher) Work(action func(x *xignals.Xignal)){
	go func() {
		action(p.Xignal)
	}()
}

func Test_Signal(t *testing.T) {

	pub := &Publisher{
	 	Xignal: xignals.NewSignal(),
	}

	condition := func(e xignals.Event) bool {
		canRun:= false
		if value, ok := e.GetPayload().(int); ok && value > 0 && value < 4 {
			canRun = true
		}
		// accepts 1,2,3
		return canRun
	}

	actions:= 0
	action := func(e xignals.Event){
		actions++
		log.Printf("action : %v", e.GetPayload())
	}

	runWhile := func(e xignals.Event) bool {
		//value , ok := e.GetPayload().(int);
		//// Some arbitrary test
		//canRun:= ok && value < 3
		//t.Logf("while: Can run %v", canRun)
		//return 	canRun
		return true
	}

	// ...
	worker:=func( x *xignals.Xignal) {
		for i := 0; i < 5; i++ {
			time.Sleep(time.Millisecond * 500)
			e:= x.Publish(i)
			if e!=nil {
				log.Printf("Error: %v", e)
				break
			}
		}
		e:= x.Complete()
		if e!=nil {
			log.Printf("Error: %v", e)
		}
	}

	pub.Work(worker) // ...

	pub.Xignal.While(runWhile).When(condition).Subscribe(action)


	time.Sleep(time.Millisecond * 500 * 6 )

	expected:= 3
	if actions!= expected {
		t.Error("Actions: Expected %v, Got: %v", expected, actions)
	}

}
