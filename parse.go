package slack

import (
	"errors"
	"fmt"

	"github.com/botopolis/bot"
	"github.com/slack-go/slack"
)

type parser func(*Adapter, *bot.Message) error

func (a *Adapter) parse(m *bot.Message, fns ...parser) error {
	for _, f := range fns {
		if err := f(a, m); err != nil {
			return err
		}
	}

	return nil
}

func parseRoom(a *Adapter, m *bot.Message) error {
	if len(m.Room) > 0 {
		if m.Room[0] == 'C' || m.Room[0] == 'D' {
			return nil
		}
	}

	if m.Room == "" {
		if msg, ok := m.Envelope.(slack.Message); ok {
			m.Room = msg.Channel
		}
		return nil
	}

	if ch, ok := a.Store.ChannelByName(m.Room); ok {
		m.Room = ch.ID
		return nil
	}

	return errors.New("Room not found")
}

func parseUser(a *Adapter, m *bot.Message) error {
	if len(m.User) > 0 {
		if m.User[0] == 'U' {
			return nil
		}
	}

	if m.User == "" {
		if msg, ok := m.Envelope.(slack.Message); ok {
			m.User = msg.User
		}
		return nil
	}

	if u, ok := a.Store.UserByName(m.User); ok {
		m.User = u.ID
		return nil
	}

	return errors.New("User not found")
}

func parseDM(a *Adapter, m *bot.Message) error {
	if len(m.Room) > 0 {
		if m.Room[0] == 'D' {
			return nil
		}
	}

	if im, ok := a.Store.IMByUserID(m.User); ok {
		m.Room = im.ID
		return nil
	}

	channel, _, _, err := a.Client.OpenConversation(&slack.OpenConversationParameters{
		ChannelID: "",
		ReturnIM:  false,
		Users:     []string{m.User},
	})
	if err != nil {
		return fmt.Errorf("Couldn't open IM to User %s: %e", m.User, err)
	}

	m.Room = channel.ID
	return nil
}

func parseParams(a *Adapter, m *bot.Message) error {
	pm, ok := m.Params.(slack.PostMessageParameters)
	if !ok {
		return nil
	}

	pm.AsUser = true
	if pm.User == "" {
		pm.User = a.BotID
	}
	m.Params = pm

	return nil
}
