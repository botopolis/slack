package action

import (
	"sync"

	"github.com/nlopes/slack"
)

type registry struct {
	once      sync.Once
	callbacks map[string]func(slack.AttachmentActionCallback)
	mu        sync.Mutex
}

func (r *registry) init() {
	r.once.Do(func() {
		r.callbacks = make(map[string]func(slack.AttachmentActionCallback))
	})
}

// Add registers a callback for the given callbackID
func (r *registry) Add(callbackID string, fn func(slack.AttachmentActionCallback)) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.init()
	r.callbacks[callbackID] = fn
}

// Run runs the callback for the slack action
func (r *registry) Run(cb slack.AttachmentActionCallback) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.init()
	if fn, ok := r.callbacks[cb.CallbackID]; ok {
		fn(cb)
	}
}
