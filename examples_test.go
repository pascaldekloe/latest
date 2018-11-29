package latest_test

import (
	"fmt"

	"github.com/pascaldekloe/latest"
)

func ExampleNewFeed() {
	// instantiate feed
	notify := make(chan interface{})
	update := latest.NewFeed(notify)
	// cleanup Go routine
	defer close(update)

	// send 3 updates
	update <- 1
	update <- 2
	update <- 3

	fmt.Println("got", <-notify)

	// send 2 more updates
	update <- 4
	update <- 5

	fmt.Println("got", <-notify)

	// Output:
	// got 3
	// got 5
}

func ExampleBroadcast() {
	// instantiate broadcast
	var b latest.Broadcast
	// cleanup Go routines
	defer b.UnsubscribeAll()

	// register 2 subscribers
	notify1 := make(chan interface{})
	notify2 := make(chan interface{})
	b.Subscribe(notify1)
	b.Subscribe(notify2)

	// demo update + notification sequence
	b.Update("1st update")
	b.Update("2nd update")
	b.Update("3rd update")
	fmt.Println("subscription 1 got", <-notify1)
	b.Update("4th update")
	fmt.Println("subscription 2 got", <-notify2)
	b.Update("5th update")
	fmt.Println("subscription 1 got", <-notify1)
	fmt.Println("subscription 2 got", <-notify2)

	// Output:
	// subscription 1 got 3rd update
	// subscription 2 got 4th update
	// subscription 1 got 5th update
	// subscription 2 got 5th update
}
