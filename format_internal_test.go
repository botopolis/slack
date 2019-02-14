package slack // don't want to export, do want to test
import (
	"testing"

	"github.com/nlopes/slack"
	"github.com/stretchr/testify/assert"
)

var formatTestCases = []struct {
	In     string
	Out    string
	Should string
}{
	{
		In:     "foo <@U1234> bar",
		Out:    "foo @bob bar",
		Should: "decodes user links",
	},
	{
		In:     "foo <@U1234|label> bar",
		Out:    "foo @label bar",
		Should: "decodes labels in user links",
	},
	{
		In:     "foo <#C1234> bar",
		Out:    "foo #general bar",
		Should: "decodes channel links",
	},
	{
		In:     "foo <#C1234|label> bar",
		Out:    "foo #label bar",
		Should: "decodes labels in channel links",
	},
	{
		In:     "foo <!everyone> bar",
		Out:    "foo @everyone bar",
		Should: "decodes everyone links",
	},
	{
		In:     "foo <!channel> bar",
		Out:    "foo @channel bar",
		Should: "decodes channel links",
	},
	{
		In:     "foo <!here> bar",
		Out:    "foo @here bar",
		Should: "decodes here links",
	},
	{
		In:     "foo <!group> bar",
		Out:    "foo @group bar",
		Should: "decodes group links",
	},
	{
		In:     "foo <!subteam^S123|@subteam> bar",
		Out:    "foo @subteam bar",
		Should: "decodes team links",
	},
	{
		In:     "foo <!foobar|hello> bar",
		Out:    "foo hello bar",
		Should: "decodes links",
	},
	{
		In:     "foo <!foobar> bar",
		Out:    "foo <!foobar> bar",
		Should: "decodes leaves unlabelled links",
	},
	{
		In:     "foo <http://example.com> bar",
		Out:    "foo http://example.com bar",
		Should: "decodes URLs",
	},
	{
		In:     "foo <http://example.com|example.com> bar",
		Out:    "foo example.com bar",
		Should: "decodes URLs with labels",
	},
	{
		In:     "foo <skype:echo123?call> bar",
		Out:    "foo skype:echo123?call bar",
		Should: "decodes skype links",
	},
	{
		In:     "foo <mailto:info@example.net> bar",
		Out:    "foo info@example.net bar",
		Should: "decode emails",
	},
	{
		In:     "foo <mailto:info@example.net|info@example.net> bar",
		Out:    "foo info@example.net bar",
		Should: "decode emails with labels",
	},
	{
		In:     "foo <@U1234|label> bar <#C1234> <!channel> <https://www.example.com|example.com>",
		Out:    "foo @label bar #general @channel example.com",
		Should: "decode multiple links",
	},
}

func TestFormatter(t *testing.T) {
	assert := assert.New(t)

	store := newTestStore()
	store.Channel.ID = "C1234"
	store.Channel.Name = "general"
	store.User = slack.User{ID: "U1234", Name: "bob"}
	store.IM.ID = "D1234"
	store.IM.User = "U1234"
	f := formatter{store}

	for _, c := range formatTestCases {
		assert.Equal(c.Out, f.Format(&slack.MessageEvent{Msg: slack.Msg{Text: c.In}}), c.Should)
	}

	store.User = slack.User{ID: "U1234", RealName: "bobby", Name: "bob"}
	assert.Equal("foo @bobby bar", f.Format(&slack.MessageEvent{Msg: slack.Msg{Text: "foo <@U1234> bar"}}), "use real name")
}

func TestFormatter_fallback(t *testing.T) {
	f := formatter{}
	in := slack.Msg{
		Text: "foo",
		Attachments: []slack.Attachment{
			{Fallback: "bar"},
			{Fallback: "baz"},
		},
	}
	assert.Equal(t,
		f.Format(&slack.MessageEvent{Msg: in}),
		"foo\nbar\nbaz",
		"should flatten attachment fallback text",
	)
}
