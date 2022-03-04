package slack

import (
	"github.com/botopolis/bot"
	"github.com/slack-go/slack"
)

type testProxy struct {
	C            chan bot.Message
	SendFunc     func(bot.Message) error
	ReactFunc    func(bot.Message) error
	SetTopicFunc func(room, topic string) error
}

func newTestProxy() *testProxy {
	return &testProxy{
		SendFunc:     func(bot.Message) error { return nil },
		ReactFunc:    func(bot.Message) error { return nil },
		SetTopicFunc: func(string, string) error { return nil },
	}
}

func (p *testProxy) Connect() chan bot.Message         { return p.C }
func (p *testProxy) Disconnect()                       {}
func (p *testProxy) Send(m bot.Message) error          { return p.SendFunc(m) }
func (p *testProxy) React(m bot.Message) error         { return p.ReactFunc(m) }
func (p *testProxy) SetTopic(room, topic string) error { return p.SetTopicFunc(room, topic) }

type testStore struct {
	LoadFunc   func(*Info)
	UpdateFunc func() error
	User       slack.User
	Channel    slack.Channel
	IM         slack.IM
}

func newTestStore() *testStore {
	return &testStore{
		LoadFunc:   func(*Info) {},
		UpdateFunc: func() error { return nil },
		User:       slack.User{},
		Channel:    slack.Channel{},
		IM:         slack.IM{},
	}
}

func (s *testStore) Load(i *Info)  { s.LoadFunc(i) }
func (s *testStore) Update() error { return s.UpdateFunc() }
func (s *testStore) UserByID(id string) (slack.User, bool) {
	if s.User.ID == id {
		return s.User, true
	}
	return s.User, false
}
func (s *testStore) Users() []slack.User { return nil }
func (s *testStore) UserByName(name string) (slack.User, bool) {
	if s.User.Name == name {
		return s.User, true
	}
	return s.User, false
}
func (s *testStore) UserByEmail(email string) (slack.User, bool) {
	if s.User.Profile.Email == email {
		return s.User, true
	}
	return s.User, false
}
func (s *testStore) Channels() []slack.Channel { return nil }
func (s *testStore) ChannelByID(id string) (slack.Channel, bool) {
	if s.Channel.ID == id {
		return s.Channel, true
	}
	return s.Channel, false
}
func (s *testStore) ChannelByName(name string) (slack.Channel, bool) {
	if s.Channel.Name == name {
		return s.Channel, true
	}
	return s.Channel, false
}
func (s *testStore) IMs() []slack.IM { return nil }
func (s *testStore) IMByID(id string) (slack.IM, bool) {
	if s.IM.ID == id {
		return s.IM, true
	}
	return s.IM, false
}
func (s *testStore) IMByUserID(id string) (slack.IM, bool) {
	if s.IM.User == id {
		return s.IM, true
	}
	return s.IM, false
}
