package signals_test

import (
	"testing"
	//"sync"
	"time"
	"log"
	"github.com/D10221/tinystore/signals"
)


func Test_Signal(t *testing.T) {

	signal:= signals.NewSignal()

	go func() {
		for i := 0; i < 5; i++ {
			time.Sleep(time.Second * 1)
			signal.Publish(i)
		}
		signal.Close()
	}()


	while := func(e signals.Event)bool {
		if value, ok := e.GetPayload().(int); ok && value > 0 && value < 6 {
			return true
		}
		return false
	}
	do:= func(e signals.Event){
		log.Printf("Received : %v", e.GetPayload())
	}

	signal.Subscribe(do, while)

	e:= signal.Close()

	if e!=nil {
		t.Log(e)
	} else {
		t.Error("Should be already closed")
	}

}
