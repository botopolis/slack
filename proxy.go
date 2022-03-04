package slack

import (
	"fmt"

	"github.com/botopolis/bot"
	"github.com/slack-go/slack"
)

type proxy struct {
	*Adapter
	RTM       *slack.RTM
	formatter formatter
}

func newProxy(a *Adapter) *proxy {
	return &proxy{
		Adapter:   a,
		RTM:       a.Client.NewRTM(),
		formatter: formatter{a.Store},
	}
}

func (p *proxy) onConnect(ev *slack.ConnectedEvent) {
	userInfo := *ev.Info.User

	p.Store.Load(&Info{
		Users: []slack.User{{Name: userInfo.Name, ID: userInfo.ID}},
	})
	if err := p.Store.Update(); err != nil {
		p.Adapter.Robot.Logger.Error("slack:", err)
	}

	p.BotID = userInfo.ID
	p.Name = userInfo.Name

	user, err := p.Client.GetUserInfo(p.BotID)
	if err != nil {
		p.Adapter.Robot.Logger.Error("slack: Unable to fetch bot user information.", err)
		return
	}

	if user.Profile.RealName != "" {
		p.Name = user.Profile.RealName
	}
}

// TextAttachmentParams contains the information needed to post a message with attachments
type TextAttachmentParams struct {
	Text        string
	Attachments []slack.Attachment
	Blocks      []slack.Block
	Username    string

	// When handling interaction requests should be sent here instead.
	ResponseURL string
	// ReplaceOriginal will update the original message if set to true.
	ReplaceOriginal bool
}

func (p *proxy) Send(m bot.Message) error {
	if m.Params == nil {
		p.RTM.SendMessage(p.RTM.NewOutgoingMessage(m.Text, m.Room))
		return nil
	}

	if pm, ok := m.Params.(TextAttachmentParams); ok {
		opts := []slack.MsgOption{
			slack.MsgOptionText(m.Text, false),
			slack.MsgOptionAttachments(pm.Attachments...),
			slack.MsgOptionBlocks(pm.Blocks...),
		}

		if pm.ResponseURL != "" {
			if pm.ReplaceOriginal {
				opts = append(opts, slack.MsgOptionReplaceOriginal(pm.ResponseURL))
			} else {
				opts = append(opts, slack.MsgOptionResponseURL(pm.ResponseURL, "in_channel"))
			}
		}

		_, _, err := p.Client.PostMessage(m.Room, opts...)
		return err
	}

	return nil
}

func (p *proxy) React(m bot.Message) error {
	msg := m.Envelope.(slack.Message)
	msgRef := slack.NewRefToMessage(msg.Channel, msg.Timestamp)
	return p.RTM.AddReaction(m.Text, msgRef)
}

func (p *proxy) SetTopic(room, topic string) error {
	_, err := p.Client.SetTopicOfConversation(room, topic)
	return err
}

func (p *proxy) Connect() chan bot.Message {
	go p.RTM.ManageConnection()
	ch := make(chan bot.Message, 32)
	go p.Forward(p.RTM.IncomingEvents, ch)
	return ch
}

func (p *proxy) Disconnect() {
	if p.RTM == nil {
		return
	}
	if err := p.RTM.Disconnect(); err != nil {
		p.Adapter.Robot.Logger.Error("slack: Unable to disconnect.", err)
	}
}

func (p *proxy) Forward(in <-chan slack.RTMEvent, out chan<- bot.Message) {
	defer close(out)
	for msg := range in {
		switch ev := msg.Data.(type) {
		case *slack.HelloEvent:
		case *slack.ConnectedEvent:
			p.onConnect(ev)
			p.Robot.Logger.Debugf("slack: Connected as %s: %d", ev.Info.User.ID, ev.ConnectionCount)
		case *slack.MessageEvent:
			out <- p.translate(ev)
		case *slack.RTMError:
			p.Robot.Logger.Errorf("slack: RTM Error: %s", ev.Error())
		case *slack.ConnectionErrorEvent:
			p.Robot.Logger.Error("slack: Connection Error: %s", ev.Error())
		case *slack.InvalidAuthEvent:
			p.Robot.Logger.Error("slack: Invalid Credentials")
			return
		}
	}
}

func (p *proxy) translate(ev *slack.MessageEvent) bot.Message {
	user, _ := p.Store.UserByID(ev.User)
	channel, _ := p.Store.ChannelByID(ev.Channel)

	// Prepend the bots name whenever a direct message is parsed
	if ev.Channel[0] == 'D' {
		ev.Text = fmt.Sprintf("@%s %s", p.Name, ev.Text)
	}

	m := bot.Message{
		User:     user.Name,
		Room:     channel.Name,
		Text:     p.formatter.Format(ev),
		Topic:    ev.Topic,
		Envelope: slack.Message(*ev),
	}

	switch ev.SubType {
	case "channel_join":
		m.Type = bot.Enter
	case "channel_leave":
		m.Type = bot.Leave
	case "channel_topic":
		m.Type = bot.Topic
	default:
		m.Type = bot.DefaultMessage
	}

	return m
}
