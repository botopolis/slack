package slack

import (
	"errors"

	"github.com/botopolis/bot"
	"github.com/slack-go/slack"
)

type slackLogger struct{ bot.Logger }

func (l slackLogger) Output(i int, s string) error {
	l.Debug("slack:", s)
	return nil
}

// Adapter is the bot slack adapter it implements
// bot.Plugin and bot.Chat interfaces
type Adapter struct {
	proxy interface {
		Connect() chan bot.Message
		Disconnect()
		Send(bot.Message) error
		React(bot.Message) error
		SetTopic(room, topic string) error
	}

	Robot  *bot.Robot
	Client *slack.Client
	Store  Store

	BotID string
	Name  string
}

// New called with one's slack token provides a new adapter
func New(secret string) *Adapter {
	client := slack.New(secret)
	a := &Adapter{
		Client: client,
		Store:  newMemoryStore(client),
	}
	a.proxy = newProxy(a)
	return a
}

// Load provides the slack adapter access to the Robot's logger
func (a *Adapter) Load(r *bot.Robot) {
	slack.OptionLog(&slackLogger{r.Logger})
	slack.OptionDebug(true)(a.Client)
	a.Robot = r
}

// Unload disconnects from slack's RTM socket
func (a *Adapter) Unload(r *bot.Robot) { a.proxy.Disconnect() }

// Username returns the bot's username
func (a *Adapter) Username() string { return a.Name }

// Messages connects to Slack's RTM API and channels messages through
func (a *Adapter) Messages() <-chan bot.Message { return a.proxy.Connect() }

func emptyMessage(m bot.Message) bool {
	return m.Text == "" && m.Params == nil
}

// Send send messages to Slack. If only text is provided, it uses
// the already open RTM connection. If slack.PostMessageParamters
// are provided in the message.Params field, it will send a web
// API request.
func (a *Adapter) Send(m bot.Message) error {
	if emptyMessage(m) {
		return nil
	}

	if err := a.parse(&m, parseRoom, parseParams); err != nil {
		return err
	}

	return a.proxy.Send(m)
}

// Direct does the same thing as send, but also ensures the message
// is sent directly to the user
func (a *Adapter) Direct(m bot.Message) error {
	if emptyMessage(m) {
		return nil
	}

	if err := a.parse(
		&m,
		parseRoom,
		parseUser,
		parseDM,
		parseParams,
	); err != nil {
		return err
	}

	return a.proxy.Send(m)
}

// Reply does the same thing as send, but prefixes the message
// with <@userID>, notifying the user of the message.
func (a *Adapter) Reply(m bot.Message) error {
	if emptyMessage(m) {
		return nil
	}

	if err := a.parse(
		&m,
		parseRoom,
		parseUser,
		parseParams,
	); err != nil {
		return err
	}

	if m.Room == "" {
		return errors.New("No room provided")
	}

	// No need to @ the user if it's a DM
	if m.Room[0] != 'D' {
		m.Text = "<@" + m.User + "> " + m.Text
	}

	return a.proxy.Send(m)
}

// Topic uses the web API to change the topic. It prefers
// the message.Room and falls back to message.Extra.Channel
// to determine what channel's topic should be updated.
func (a *Adapter) Topic(m bot.Message) error {
	if err := parseRoom(a, &m); err != nil {
		return err
	}

	if m.Room == "" {
		return errors.New("No Channel provided")
	}

	return a.proxy.SetTopic(m.Room, m.Topic)
}

// React adds an emote to the last message sent (requires an Envelope to be set).
// It relies on the timestamp and channel for a message to be present
func (a *Adapter) React(m bot.Message) error {
	if _, ok := m.Envelope.(slack.Message); !ok {
		return errors.New("Empty envelope provided")
	}
	return a.proxy.React(m)
}
