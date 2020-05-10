// Package latest provides safe ways to keep track of a current version.
package latest

import "sync"

// NewFeed starts a notification routine and returns an update input channel.
// A close on the input channel terminates the processing; notify stays open.
// Slow acceptance on notify does not block input. Instead, the notification
// continues with the latest value, discarding all pending [unused] updates.
// Be careful with buffered channels as they interfear with data freshness.
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

// Broadcast offers a publish–subscribe for update notification. Each subscriber
// has it's own isolated update process. Slow receivals do not block operation.
// Instead, the notification continues with the latest value, discarding all
// pending [unused] updates. All methods may be called concurrently.
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

// SubscriptionCount returns the number of broadcast channels.
func (b *Broadcast) SubscriptionCount() int {
	b.RLock()
	defer b.RUnlock()

	return len(b.feeds)
}
