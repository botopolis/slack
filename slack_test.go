package slack

import (
	"testing"

	"github.com/botopolis/bot"
	"github.com/slack-go/slack"
	"github.com/stretchr/testify/assert"
)

func TestSend_blank(t *testing.T) {
	proxy := newTestProxy()
	proxy.SendFunc = func(bot.Message) error {
		t.Errorf("Should not call Send")
		return nil
	}
	adapter := Adapter{proxy: proxy}
	assert.Nil(t, adapter.Send(bot.Message{}))
}

func TestSend_parse(t *testing.T) {
	user := "U1234"
	cases := []struct {
		In  bot.Message
		Out bot.Message
		Err bool
	}{
		{
			In:  bot.Message{Room: "general", Text: "foo"},
			Out: bot.Message{Room: "C1234", Text: "foo"},
		},
		{
			In: bot.Message{Room: "general", Params: slack.PostMessageParameters{}},
			Out: bot.Message{Room: "C1234", Params: slack.PostMessageParameters{
				AsUser: true,
				User:   user,
			}},
		},
		{
			In:  bot.Message{Room: "random", Text: "foo"},
			Err: true,
		},
	}
	store := newTestStore()
	store.Channel.ID = "C1234"
	store.Channel.Name = "general"

	for _, c := range cases {
		proxy, run := setUpProxySend(t, c.Out)
		adapter := Adapter{Store: store, proxy: proxy, BotID: user}
		err := adapter.Send(c.In)
		if c.Err {
			assert.NotNil(t, err)
			assert.False(t, *run)
		} else {
			assert.Nil(t, err)
			assert.True(t, *run)
		}
	}
}

func TestDirect_blank(t *testing.T) {
	proxy := newTestProxy()
	proxy.SendFunc = func(bot.Message) error {
		t.Errorf("Should not call Send")
		return nil
	}
	adapter := Adapter{proxy: proxy}
	assert.Nil(t, adapter.Direct(bot.Message{}))
}

func TestReact(t *testing.T) {
	proxy := newTestProxy()
	adapter := Adapter{proxy: proxy}
	cases := []struct {
		Name string
		In   interface{}
		Err  bool
	}{
		{Name: "With no Envelope", In: nil, Err: true},
		{Name: "With a mistyped Envelope", In: bot.Message{}, Err: true},
		{Name: "With an Envelope", In: slack.Message{}, Err: false},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			err := adapter.React(bot.Message{Envelope: c.In})
			if c.Err {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDirect(t *testing.T) {
	user := "U1234"
	cases := []struct {
		In  bot.Message
		Out bot.Message
		Err bool
	}{
		{
			In:  bot.Message{Room: "D4321", Text: "foo"},
			Out: bot.Message{Room: "D4321", Text: "foo"},
		},
		{
			In:  bot.Message{User: user, Text: "foo"},
			Out: bot.Message{User: user, Room: "D1234", Text: "foo"},
		},
		{
			In:  bot.Message{User: "Jean", Room: "general", Text: "foo"},
			Out: bot.Message{User: user, Room: "D1234", Text: "foo"},
		},
		{
			In:  bot.Message{User: "Jane", Room: "general", Text: "foo"},
			Err: true,
		},
		{
			In: bot.Message{User: "Jean", Room: "general", Params: slack.PostMessageParameters{}},
			Out: bot.Message{User: user, Room: "D1234", Params: slack.PostMessageParameters{
				AsUser: true,
				User:   user,
			}},
		},
	}
	store := newTestStore()
	store.Channel.ID = "C1234"
	store.Channel.Name = "general"
	store.User.ID = user
	store.User.Name = "Jean"
	store.IM.ID = "D1234"
	store.IM.User = user

	for _, c := range cases {
		proxy, run := setUpProxySend(t, c.Out)
		adapter := Adapter{Store: store, proxy: proxy, BotID: user}
		err := adapter.Direct(c.In)
		if c.Err {
			assert.NotNil(t, err)
			assert.False(t, *run)
		} else {
			assert.Nil(t, err)
			assert.True(t, *run)
		}
	}
}

func TestReply_blank(t *testing.T) {
	proxy := newTestProxy()
	proxy.SendFunc = func(bot.Message) error {
		t.Errorf("Should not call Send")
		return nil
	}
	adapter := Adapter{proxy: proxy}
	assert.Nil(t, adapter.Reply(bot.Message{}))
}

func TestReply(t *testing.T) {
	user := "U1234"
	envelope := slack.Message{}
	envelope.User = user

	cases := []struct {
		In  bot.Message
		Out bot.Message
		Err bool
	}{
		{
			In:  bot.Message{Room: "D4321", Text: "foo", Envelope: envelope},
			Out: bot.Message{User: user, Room: "D4321", Text: "foo", Envelope: envelope},
		},
		{
			In:  bot.Message{Room: "general", Text: "foo", Envelope: envelope},
			Out: bot.Message{User: user, Room: "C1234", Text: "<@U1234> foo", Envelope: envelope},
		},
		{
			In:  bot.Message{User: "Jane", Room: "general", Text: "foo"},
			Err: true,
		},
		{
			In:  bot.Message{User: user, Text: "foo"},
			Err: true,
		},
	}
	store := newTestStore()
	store.Channel.ID = "C1234"
	store.Channel.Name = "general"
	store.User.ID = user
	store.User.Name = "Jean"

	for _, c := range cases {
		proxy, run := setUpProxySend(t, c.Out)
		adapter := Adapter{Store: store, proxy: proxy, BotID: user}
		err := adapter.Reply(c.In)
		if c.Err {
			assert.NotNil(t, err)
			assert.False(t, *run)
		} else {
			assert.Nil(t, err)
			assert.True(t, *run)
		}
	}
}

func TestTopic(t *testing.T) {
	cases := []struct {
		In  bot.Message
		Out bot.Message
		Err bool
	}{
		{
			In:  bot.Message{Room: "general", Topic: "foo"},
			Out: bot.Message{Room: "C1234", Topic: "foo"},
		},
		{
			In:  bot.Message{Room: "general", Topic: ""},
			Out: bot.Message{Room: "C1234", Topic: ""},
		},
		{
			In:  bot.Message{Room: "random", Topic: ""},
			Err: true,
		},
		{
			In:  bot.Message{Topic: ""},
			Err: true,
		},
	}
	store := newTestStore()
	store.Channel.ID = "C1234"
	store.Channel.Name = "general"

	for _, c := range cases {
		var run bool
		proxy := newTestProxy()
		proxy.SetTopicFunc = func(room, topic string) error {
			assert.Equal(t, c.Out.Room, room)
			assert.Equal(t, c.Out.Topic, topic)
			run = true
			return nil
		}
		adapter := Adapter{Store: store, proxy: proxy}
		err := adapter.Topic(c.In)
		if c.Err {
			assert.NotNil(t, err)
			assert.False(t, run)
		} else {
			assert.Nil(t, err)
			assert.True(t, run)
		}
	}
}

func setUpProxySend(t *testing.T, out bot.Message) (*testProxy, *bool) {
	var run bool
	proxy := newTestProxy()
	proxy.SendFunc = func(m bot.Message) error {
		assert.Equal(t, out, m)
		run = true
		return nil
	}

	return proxy, &run
}
