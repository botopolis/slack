package action

import (
	"sync"

	"github.com/slack-go/slack"
)

type registry struct {
	once        sync.Once
	mu          sync.Mutex
	callbacks   map[string]func(slack.InteractionCallback)
	subscribers map[string]func(slack.InteractionCallback)
}

func (r *registry) init() {
	r.once.Do(func() {
		r.callbacks = make(map[string]func(slack.InteractionCallback))
		r.subscribers = make(map[string]func(slack.InteractionCallback))
	})
}

// Add registers a callback for the given callbackID
func (r *registry) Add(callbackID string, fn func(slack.InteractionCallback)) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.init()
	r.callbacks[callbackID] = fn
}

// Subscribe registers a callback when an interaction callback happens. Use this when there is
// no callbackID to reference in the payload. Caller should do their own filtering to make sure
// the interaction callback is for them.
func (r *registry) Subscribe(clientName string, fn func(slack.InteractionCallback)) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.init()
	r.subscribers[clientName] = fn
}

// Run runs the callback for the slack action
func (r *registry) Run(cb slack.InteractionCallback) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.init()
	if fn, ok := r.callbacks[cb.CallbackID]; ok {
		fn(cb)
	}

	for _, fn := range r.subscribers {
		fn(cb)
	}
}
