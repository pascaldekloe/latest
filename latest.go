// Package latest provides safe ways to keep track of a current version.
package latest

import "sync"

// NewFeed starts a notification routine and returns an update input channel.
// A close on the input channel terminates the processing; notify stays open.
// Slow acceptance on notify does not block input. Instead, the notification
// continues with the latest value, discarding all pending [unused] updates.
// Be careful with buffered channels as they interfear with data freshness.
func NewFeed(notify chan<- interface{}) chan<- interface{} {
	update := make(chan interface{})

	go func() {
		for {
			// await update
			latest, ok := <-update
			if !ok {
				return
			}

		Notify:
			for {
				select {
				case notify <- latest:
					// passed the update
					break Notify

				case v, ok := <-update:
					if !ok {
						return
					}
					// replace the update with a newer one
					latest = v
				}
			}
		}
	}()

	return update
}

// Broadcast offers a publish–subscribe for update notification.
// All methods from Broadcast are thread-safe and non-blocking.
type Broadcast struct {
	sync.RWMutex // subscription lock
	feeds        map[chan<- interface{}]chan<- interface{}
}

// Update sets the current version.
func (b *Broadcast) Update(v interface{}) {
	b.RLock()
	defer b.RUnlock()

	for _, update := range b.feeds {
		update <- v
	}
}

// Subscribe adds an update receiver.
func (b *Broadcast) Subscribe(notify chan<- interface{}) {
	update := NewFeed(notify)

	b.Lock()
	defer b.Unlock()

	if _, ok := b.feeds[notify]; ok {
		// already subscribed
		close(update)
		return
	}

	if b.feeds == nil {
		b.feeds = make(map[chan<- interface{}]chan<- interface{})
	}
	b.feeds[notify] = update
}

// Unsubscribe terminates a subscription.
func (b *Broadcast) Unsubscribe(notify chan<- interface{}) {
	b.Lock()
	defer b.Unlock()

	update, ok := b.feeds[notify]
	if ok {
		delete(b.feeds, notify)
		close(update)
	}
}

// UnsubscribeAll terminates all subscriptions.
func (b *Broadcast) UnsubscribeAll() {
	b.Lock()
	defer b.Unlock()

	for notify, update := range b.feeds {
		delete(b.feeds, notify)
		close(update)
	}
}
