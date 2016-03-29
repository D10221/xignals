package xignals_test

import (
	"testing"
	//"sync"
	"time"
	"log"
	"github.com/D10221/xignals"
)


func Test_Signal(t *testing.T) {

	signal:= xignals.NewSignal()

	go func() {
		for i := 0; i < 5; i++ {
			time.Sleep(time.Second * 1)
			signal.Publish(i)
		}
		signal.Close()
	}()


	condition := func(e xignals.Event) bool {
		if value, ok := e.GetPayload().(int); ok && value > 0 && value < 6 {
			return true
		}
		return false
	}

	action := func(e xignals.Event){
		log.Printf("Received : %v", e.GetPayload())
	}

	canRun:= func(e xignals.Event) bool {
		value , ok := e.GetPayload().(int);
		// Some arbitrary test
		return 	ok && value < 3
	}

	signal.While(canRun).When(condition).Subscribe(action)

	e:= signal.Close()

	if e!=nil {
		t.Log(e)
	} else {
		t.Error("Should be already closed")
	}

}
