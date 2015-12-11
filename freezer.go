package crawl

import "sync"

// Freezer - Makes things freeze.
type Freezer struct {
	mutex  *sync.RWMutex
	freeze bool
	notify chan chan bool
}

// NewFreezer - Creates a new freezer.
func NewFreezer(capacity int) *Freezer {
	return &Freezer{
		mutex:  new(sync.RWMutex),
		notify: make(chan chan bool, capacity),
	}
}

// IsFreezed - Returns either the freezer is freezed.
func (freezer *Freezer) IsFreezed() bool {
	freezer.mutex.RLock()
	defer freezer.mutex.RUnlock()
	return freezer.freeze
}

// Freeze - Freezes the freezer.
func (freezer *Freezer) Freeze() {
	freezer.mutex.Lock()
	defer freezer.mutex.Unlock()
	freezer.freeze = true
}

// Unfreeze - Unfreezes the freezer.
func (freezer *Freezer) Unfreeze() {
	freezer.mutex.Lock()
	defer freezer.mutex.Unlock()
	if !freezer.freeze {
		return
	}
	freezer.freeze = false
	for {
		select {
		case ch := <-freezer.notify:
			ch <- true
		default:
			return
		}
	}
}

// Wait - If freezed waits until unfreezed.
func (freezer *Freezer) Wait() {
	if !freezer.freeze {
		return
	}
	// Add freeze notifier
	freezer.mutex.Lock()
	ch := make(chan bool, 1)
	freezer.notify <- ch
	freezer.mutex.Unlock()
	<-ch
	return
}
