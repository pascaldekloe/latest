// Package latest provides ways to keep track of a current version.
package latest

import "sync"

// NewFeed returns a non-blocking input channel for notify.
// The notification process uses the latest input only. Any
// pending [undelivered] submissions are freely discarded.
// Processing terminates when the input channel is closed.
// Be careful with buffered channels as they interfear with
// data freshness.
func NewFeed(notify chan<- interface{}) chan<- interface{} {
	feed := make(chan interface{})

	go func() {
		for {
			// await update
			latest, ok := <-feed
			for {
				if !ok {
					return
				}
				select {
				case latest, ok = <-feed:
					continue // newer update

				case notify <- latest:
					break // update delivered
				}
				break
			}
		}
	}()

	return feed
}

// Broadcast enqueues the latest Update (publication) for each subscriber
// individually. Slow subscribers do not block Update submission. Instead,
// any pending [undelivered] publications are replaced with the latest.
// Multiple goroutines may invoke methods on a Broadcast simultaneously.
type Broadcast struct {
	sync.RWMutex // subscription lock
	feeds        map[chan<- interface{}]chan<- interface{}
}

// Update sets the current version.
func (b *Broadcast) Update(v interface{}) {
	b.RLock()
	defer b.RUnlock()

	for _, feed := range b.feeds {
		feed <- v
	}
}

// Subscribe adds an update receiver.
// Duplicate subscriptions are ignored.
func (b *Broadcast) Subscribe(notify chan<- interface{}) {
	feed := NewFeed(notify)

	b.Lock()
	defer b.Unlock()

	if _, ok := b.feeds[notify]; ok {
		// already subscribed
		close(feed)
		return
	}

	if b.feeds == nil {
		b.feeds = make(map[chan<- interface{}]chan<- interface{})
	}
	b.feeds[notify] = feed
}

// Unsubscribe terminates a subscription.
func (b *Broadcast) Unsubscribe(notify chan<- interface{}) {
	b.Lock()
	defer b.Unlock()

	feed, ok := b.feeds[notify]
	if ok {
		delete(b.feeds, notify)
		close(feed)
	}
}

// UnsubscribeAll terminates all subscriptions.
func (b *Broadcast) UnsubscribeAll() {
	b.Lock()
	defer b.Unlock()

	for notify, feed := range b.feeds {
		delete(b.feeds, notify)
		close(feed)
	}
}

// SubscriptionCount returns the number of active subscriptions.
func (b *Broadcast) SubscriptionCount() int {
	b.RLock()
	defer b.RUnlock()

	return len(b.feeds)
}
