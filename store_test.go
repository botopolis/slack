package slack

import (
	"testing"

	"github.com/slack-go/slack"
	"github.com/stretchr/testify/assert"
)

func TestStore_users(t *testing.T) {
	store := newMemoryStore(&slack.Client{})
	info := slackUserInfo()
	store.Load(info)

	var (
		user slack.User
		ok   bool
	)

	user, ok = store.UserByID("U1234")
	assert.Equal(t, info.Users[0], user)
	assert.True(t, ok)

	_, ok = store.UserByID("U4321")
	assert.False(t, ok)

	user, ok = store.UserByName("Jean")
	assert.Equal(t, info.Users[0], user)
	assert.True(t, ok)

	_, ok = store.UserByName("Joan")
	assert.False(t, ok)
}

func TestStore_channels(t *testing.T) {
	store := newMemoryStore(&slack.Client{})
	info := slackUserInfo()
	store.Load(info)

	var (
		channel slack.Channel
		ok      bool
	)

	channel, ok = store.ChannelByID("C1234")
	assert.Equal(t, info.Channels[0], channel)
	assert.True(t, ok)

	_, ok = store.ChannelByID("C4321")
	assert.False(t, ok)

	channel, ok = store.ChannelByName("general")
	assert.Equal(t, info.Channels[0], channel)
	assert.True(t, ok)

	_, ok = store.ChannelByName("random")
	assert.False(t, ok)
}

func TestStore_ims(t *testing.T) {
	store := newMemoryStore(&slack.Client{})
	info := slackUserInfo()
	store.Load(info)

	var (
		im slack.IM
		ok bool
	)

	im, ok = store.IMByID("D1234")
	assert.Equal(t, info.IMs[0], im)
	assert.True(t, ok)

	_, ok = store.IMByID("D4321")
	assert.False(t, ok)

	im, ok = store.IMByUserID("U1234")
	assert.Equal(t, info.IMs[0], im)
	assert.True(t, ok)

	_, ok = store.IMByUserID("U4321")
	assert.False(t, ok)
}

func slackUserInfo() *Info {
	user := slack.User{ID: "U1234", Name: "Jean"}
	channel := slack.Channel{}
	channel.ID = "C1234"
	channel.Name = "general"
	im := slack.IM{}
	im.ID = "D1234"
	im.User = "U1234"
	return &Info{
		Users:    []slack.User{user},
		Channels: []slack.Channel{channel},
		IMs:      []slack.IM{im},
	}
}
