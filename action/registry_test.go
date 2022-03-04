package action

import (
	"testing"

	"github.com/slack-go/slack"
	"github.com/stretchr/testify/assert"
)

func TestRegistry(t *testing.T) {
	counter := 0
	subCounter := 0
	callbackID := "foobar"
	example := func(slack.InteractionCallback) { counter++ }

	r := registry{}
	r.Add(callbackID, example)
	r.Subscribe("foo", func(slack.InteractionCallback) { subCounter++ })
	r.Subscribe("bar", func(slack.InteractionCallback) { subCounter++ })
	r.Run(slack.InteractionCallback{CallbackID: "foobar"})

	assert.Equal(t, 1, counter)
	assert.Equal(t, 2, subCounter)
}
