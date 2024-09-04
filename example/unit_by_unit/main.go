package main

import (
	"context"
	"github.com/craftdome/go-pipeline"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"time"
)

func main() {
	unit := pipeline.NewUnit[string, string]()
	unit.OnExecute = func(s string) (string, error) {
		// imitating a long processing...
		time.Sleep(time.Duration(rand.Intn(5)) * time.Second)
		// processing an input...
		// returning some result...
		return strings.ToUpper(s), nil
	}

	unit2 := pipeline.NewUnit[string, int]()
	unit2.OnExecute = func(s string) (int, error) {
		// imitating a long processing...
		time.Sleep(time.Duration(rand.Intn(5)) * time.Second)
		// processing an input...
		// returning some result...
		return len(s), nil
	}

	// Connecting unit to unit2
	unit.SetNextUnit(unit2)

	unit.Start()
	unit2.Start()

	// Generating some data
	var stopped bool
	done := make(chan error)
	go func() {
		for i := 0; !stopped; i++ {
			unit.Input() <- strconv.FormatUint(rand.Uint64(), 36)
		}

		// Stopping the units with timeout context
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		// stopping 1st unit...
		if err := unit.Stop(ctx); err != nil {
			done <- err
			return
		}
		// stopping 2nd unit...
		if err := unit2.Stop(ctx); err != nil {
			done <- err
			return
		}
		done <- nil
	}()

	// Getting results
	go func() {
		for output := range unit2.Output() {
			log.Printf("%#v", output)
		}
	}()

	// Graceful shutdown
	interrupt := make(chan os.Signal)
	signal.Notify(interrupt, os.Interrupt, os.Kill)
	<-interrupt
	stopped = true
	if err := <-done; err != nil {
		log.Fatal(err)
	}

	log.Println("Done")
}
