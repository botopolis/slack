package slack

import (
	"testing"

	"github.com/botopolis/bot"
	"github.com/slack-go/slack"
	"github.com/stretchr/testify/assert"
)

func TestParseRoom(t *testing.T) {
	msg := slack.Message{Msg: slack.Msg{Channel: "C4321"}}
	cases := []struct {
		In  bot.Message
		Out bot.Message
	}{
		{
			In:  bot.Message{Room: "C1234"},
			Out: bot.Message{Room: "C1234"},
		},
		{
			In:  bot.Message{Room: "D1234"},
			Out: bot.Message{Room: "D1234"},
		},
		{
			In:  bot.Message{Room: "", Envelope: msg},
			Out: bot.Message{Room: "C4321", Envelope: msg},
		},
		{
			In:  bot.Message{Room: "general"},
			Out: bot.Message{Room: "C1234"},
		},
	}

	store := newTestStore()
	store.Channel.ID = "C1234"
	store.Channel.Name = "general"

	for _, c := range cases {
		parseRoom(&Adapter{Store: store}, &c.In)
		assert.Equal(t, c.Out, c.In)
	}
}

func TestParseUser(t *testing.T) {
	msg := slack.Message{Msg: slack.Msg{User: "U4321"}}
	cases := []struct {
		In  bot.Message
		Out bot.Message
	}{
		{
			In:  bot.Message{User: "U1234"},
			Out: bot.Message{User: "U1234"},
		},
		{
			In:  bot.Message{User: "", Envelope: msg},
			Out: bot.Message{User: "U4321", Envelope: msg},
		},
		{
			In:  bot.Message{User: "bob"},
			Out: bot.Message{User: "U1234"},
		},
	}

	store := newTestStore()
	store.User = slack.User{ID: "U1234", Name: "bob"}

	for _, c := range cases {
		parseUser(&Adapter{Store: store}, &c.In)
		assert.Equal(t, c.Out, c.In)
	}
}

func TestParseDM(t *testing.T) {
	cases := []struct {
		In  bot.Message
		Out bot.Message
	}{
		{
			In:  bot.Message{Room: "D1234"},
			Out: bot.Message{Room: "D1234"},
		},
		{
			In:  bot.Message{User: "U4321"},
			Out: bot.Message{User: "U4321", Room: "D1234"},
		},
	}

	store := newTestStore()
	store.IM.ID = "D1234"
	store.IM.User = "U4321"

	for _, c := range cases {
		parseDM(&Adapter{Store: store}, &c.In)
		assert.Equal(t, c.Out, c.In)
	}
}

func TestParseParams(t *testing.T) {
	id := "B1234"
	cases := []struct {
		In  bot.Message
		Out bot.Message
	}{
		{
			In: bot.Message{Params: slack.PostMessageParameters{}},
			Out: bot.Message{Params: slack.PostMessageParameters{
				User:   id,
				AsUser: true,
			}},
		},
		{
			In: bot.Message{Params: slack.PostMessageParameters{
				User: "U5432",
			}},
			Out: bot.Message{Params: slack.PostMessageParameters{
				User:   "U5432",
				AsUser: true,
			}},
		},
		{
			In:  bot.Message{Params: slack.Message{}},
			Out: bot.Message{Params: slack.Message{}},
		},
	}

	for _, c := range cases {
		parseParams(&Adapter{BotID: id}, &c.In)
		assert.Equal(t, c.Out, c.In)
	}
}

func TestParseChain(t *testing.T) {
	in := bot.Message{
		User:   "U4321",
		Params: slack.PostMessageParameters{},
	}
	out := bot.Message{
		User: "U4321",
		Room: "D1234",
		Params: slack.PostMessageParameters{
			AsUser: true,
			User:   "B1234",
		},
	}

	store := newTestStore()
	store.IM.ID = "D1234"
	store.IM.User = "U4321"
	a := Adapter{BotID: "B1234", Store: store}

	a.parse(&in, parseDM, parseParams)
	assert.Equal(t, out, in)
}
