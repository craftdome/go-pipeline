package main

import (
	"context"
	"github.com/craftdome/go-pipeline"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"strconv"
	"time"
)

type User struct {
	Name string
}

type UserInventory struct {
	User  User
	Items []int
}

func main() {
	// Initializing a unit
	unit := go_pipeline.NewUnit[User, UserInventory](
		go_pipeline.WithWorkers[User, UserInventory](32),
	)

	// Set OnExecute action to the unit
	unit.OnExecute = func(u User) (UserInventory, error) {
		// imitating a long processing...
		time.Sleep(time.Duration(rand.Intn(10)) * time.Second)
		// returning some result...
		return UserInventory{
			User:  u,
			Items: []int{rand.Intn(1000), rand.Intn(1000), rand.Intn(1000)},
		}, nil
	}

	// Set OnError action to the unit
	unit.OnError = func(err error) {
		// error handling...
		log.Println(err)
	}

	// Start the unit to reading input chan
	unit.Start()

	// Generating some data
	var stopped bool
	done := make(chan error)
	go func() {
		for i := 0; !stopped; i++ {
			unit.Input() <- User{Name: "User" + strconv.Itoa(i)}
		}

		// Stopping the unit with timeout context
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		done <- unit.Stop(ctx)
	}()

	// Getting results
	go func() {
		for output := range unit.Output() {
			log.Printf("%#v", output)
		}
	}()

	// Graceful shutdown
	interrupt := make(chan os.Signal)
	signal.Notify(interrupt, os.Interrupt)
	<-interrupt
	stopped = true
	if err := <-done; err != nil {
		log.Fatal(err)
	}

	log.Println("Done")
}
