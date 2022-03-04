package slack

import (
	"testing"

	"github.com/botopolis/bot"
	"github.com/botopolis/bot/mock"
	"github.com/slack-go/slack"
	"github.com/stretchr/testify/assert"
)

func TestProxy(t *testing.T) {
	assert := assert.New(t)

	room := "C1234"
	user := "U1234"
	text := "foo bar"
	topic := "generally awesome"

	proxyTestCases := []struct {
		In     slack.MessageEvent
		Assert func(bot.Message, bot.Message)
	}{
		{
			In: slack.MessageEvent(slack.Message{Msg: slack.Msg{
				SubType: "",
				User:    user,
				Channel: room,
				Text:    text,
			}}),
			Assert: func(expected bot.Message, result bot.Message) {
				assert.Equal(bot.DefaultMessage, result.Type)
				assert.Equal(expected.Text, result.Text)
			},
		},
		{
			In: slack.MessageEvent(slack.Message{Msg: slack.Msg{
				SubType: "channel_join",
				User:    user,
				Channel: room,
				Text:    text,
			}}),
			Assert: func(expected bot.Message, result bot.Message) {
				assert.Equal(bot.Enter, result.Type)
				assert.Equal(expected.Text, result.Text)
			},
		},
		{
			In: slack.MessageEvent(slack.Message{Msg: slack.Msg{
				SubType: "channel_leave",
				User:    user,
				Channel: room,
				Text:    text,
			}}),
			Assert: func(expected bot.Message, result bot.Message) {
				assert.Equal(bot.Leave, result.Type)
				assert.Equal(expected.Text, result.Text)
			},
		},
		{
			In: slack.MessageEvent(slack.Message{Msg: slack.Msg{
				SubType: "channel_topic",
				User:    user,
				Channel: room,
				Topic:   topic,
				Text:    text,
			}}),
			Assert: func(expected bot.Message, result bot.Message) {
				assert.Equal(bot.Topic, result.Type)
				assert.Equal(expected.Topic, result.Topic)
			},
		},
	}

	store := newTestStore()
	p := proxy{
		Adapter:   New(""),
		formatter: formatter{store},
	}
	for _, c := range proxyTestCases {
		in := make(chan slack.RTMEvent, 2)
		out := make(chan bot.Message, 2)
		go p.Forward(in, out)

		in <- slack.RTMEvent{Type: "message", Data: &c.In}
		c.Assert(bot.Message{
			User:     "bob",
			Room:     "general",
			Text:     text,
			Envelope: &c.In,
			Topic:    topic,
		}, <-out)
		close(in)
	}
}

func TestProxyForward_connect(t *testing.T) {
	var run bool
	user := &slack.UserDetails{ID: "B1234", Name: "beardroid"}
	info := &Info{Users: []slack.User{{Name: user.Name, ID: user.ID}}}

	store := newTestStore()
	store.LoadFunc = func(i *Info) {
		assert.Equal(t, info, i)
		run = true
	}
	p := proxy{Adapter: New("")}
	p.Adapter.Store = store
	p.Load(bot.New(mock.NewChat()))

	ch := make(chan slack.RTMEvent, 2)
	ch <- slack.RTMEvent{Data: &slack.ConnectedEvent{Info: &slack.Info{User: user}}}
	close(ch)
	p.Forward(ch, make(chan bot.Message))

	assert.Equal(t, info.Users[0].ID, p.Adapter.BotID)
	assert.Equal(t, info.Users[0].Name, p.Adapter.Name)
	assert.True(t, run)
}

func TestProxyForward_invalidAuth(t *testing.T) {
	done := make(chan bool)
	p := proxy{Adapter: New("")}
	p.Load(&bot.Robot{Logger: mock.NewLogger()})

	// Send the event, explicitly don't close the channel
	ch := make(chan slack.RTMEvent, 2)
	ch <- slack.RTMEvent{Data: &slack.InvalidAuthEvent{}}

	go func() {
		p.Forward(ch, make(chan bot.Message))
		done <- true
	}()

	assert.True(t, <-done)
}
