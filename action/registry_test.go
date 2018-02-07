package action

import (
	"testing"

	"github.com/nlopes/slack"
	"github.com/stretchr/testify/assert"
)

func TestRegistry(t *testing.T) {
	counter := 0
	callbackID := "foobar"
	example := func(slack.AttachmentActionCallback) { counter++ }

	r := registry{}
	r.Add(callbackID, example)
	r.Run(slack.AttachmentActionCallback{CallbackID: "foobar"})

	assert.Equal(t, 1, counter)
}
