package slack

import (
	"sync"

	"github.com/slack-go/slack"
)

// Info contains lists of relevant slack information (users, channels, ims)
type Info struct {
	Users    []slack.User
	Channels []slack.Channel
	IMs      []slack.IM
}

// Store is the interface to expect from adapter.Store
type Store interface {
	// Load takes Info and adds new users and channels from it
	Load(*Info)
	// Update queries Slack's web API for users and channels
	Update() error
	// UserByID queries the store for a User by ID
	UserByID(id string) (slack.User, bool)
	// UserByName queries the store for a User by Name
	UserByName(name string) (slack.User, bool)
	// UserByEmail queries the store for a User by Name
	UserByEmail(name string) (slack.User, bool)
	// ChannelByID queries the store for a Channel by ID
	ChannelByID(id string) (slack.Channel, bool)
	// ChannelByName queries the store for a Channel by Name
	ChannelByName(id string) (slack.Channel, bool)
	// IMByID queries the store for a IM by ID
	IMByID(id string) (slack.IM, bool)
	// IMByUserID queries the store for a DM by User ID
	IMByUserID(userID string) (slack.IM, bool)
}

type memoryStore struct {
	mu       sync.RWMutex
	client   *slack.Client
	indices  map[string]string
	users    map[string]slack.User
	channels map[string]slack.Channel
	ims      map[string]slack.IM
}

func newMemoryStore(c *slack.Client) *memoryStore {
	m := &memoryStore{
		client:   c,
		indices:  make(map[string]string),
		users:    make(map[string]slack.User),
		channels: make(map[string]slack.Channel),
		ims:      make(map[string]slack.IM),
	}
	return m
}

func (s *memoryStore) Load(i *Info) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, u := range i.Users {
		s.users[u.ID] = u
		s.indices["user:name:"+u.Name] = u.ID
		s.indices["user:email:"+u.Profile.Email] = u.ID
	}

	for _, ch := range i.Channels {
		s.channels[ch.ID] = ch
		s.indices["channel:name:"+ch.Name] = ch.ID
	}

	for _, im := range i.IMs {
		s.ims[im.ID] = im
		s.indices["im:userID:"+im.User] = im.ID
	}
}

func (s *memoryStore) getAllChannelsForTypes(t ...string) (allChannels []slack.Channel, err error) {
	conversationParams := slack.GetConversationsParameters{ExcludeArchived: true, Types: t}
	for {
		channels, cursor, err := s.client.GetConversations(&conversationParams)
		if err != nil {
			return channels, err
		}

		allChannels = append(allChannels, channels...)
		if cursor == "" {
			// End of pagination
			break
		}
		conversationParams.Cursor = cursor
	}

	return allChannels, nil
}

func (s *memoryStore) Update() (err error) {
	info := Info{}
	if info.Users, err = s.client.GetUsers(); err != nil {
		return err
	}

	if info.Channels, err = s.getAllChannelsForTypes("public_channel", "private_channel"); err != nil {
		return err
	}

	imChannels, err := s.getAllChannelsForTypes("im")
	if err != nil {
		return err
	}
	for _, channel := range imChannels {
		info.IMs = append(info.IMs, slack.IM{Conversation: channel.GroupConversation.Conversation})
	}

	s.Load(&info)
	return err
}

func (s *memoryStore) UserByID(id string) (slack.User, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	u, ok := s.users[id]
	return u, ok
}

func (s *memoryStore) UserByName(name string) (slack.User, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.UserByID(s.indices["user:name:"+name])
}

func (s *memoryStore) UserByEmail(name string) (slack.User, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.UserByID(s.indices["user:email:"+name])
}

func (s *memoryStore) ChannelByID(id string) (slack.Channel, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	ch, ok := s.channels[id]
	return ch, ok
}

func (s *memoryStore) ChannelByName(name string) (slack.Channel, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.ChannelByID(s.indices["channel:name:"+name])
}

func (s *memoryStore) IMByID(id string) (slack.IM, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	dm, ok := s.ims[id]
	return dm, ok
}

func (s *memoryStore) IMByUserID(userID string) (slack.IM, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.IMByID(s.indices["im:userID:"+userID])
}
