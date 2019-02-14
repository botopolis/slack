package slack

import (
	"regexp"
	"strings"

	"github.com/nlopes/slack"
)

const linkRegexp = "<" +
	// link type
	"([@#!])?" +
	// link
	"([^>|]+)" +
	// start of |label (optional)
	"(?:\\|" +
	// label
	"([^>]+)" +
	// end of label
	")?>"

var keywords = map[string]interface{}{
	"channel":  nil,
	"group":    nil,
	"everyone": nil,
	"here":     nil,
}

type formatter struct{ store Store }

func (f formatter) Format(msg *slack.MessageEvent) string {
	return f.formatLinks(f.flatten(msg))
}

func (f formatter) formatLinks(in string) string {
	r := regexp.MustCompile(linkRegexp)
	text := replaceAllStringSubmatchFunc(r, in, func(match []string) string {
		t, link, label := match[1], match[2], match[3]
		switch t {
		case "@":
			if label != "" {
				return "@" + label
			}
			if user, ok := f.store.UserByID(link); ok {
				if user.RealName != "" {
					return "@" + user.RealName
				}
				return "@" + user.Name
			}
		case "#":
			if label != "" {
				return "#" + label
			}
			if channel, ok := f.store.ChannelByID(link); ok {
				return "#" + channel.Name
			}
		case "!":
			if _, ok := keywords[link]; ok {
				return "@" + link
			}
			if label != "" {
				return label
			}
			return match[0]
		default:
			link = strings.Replace(link, "mailto:", "", 1)
			if label != "" && strings.Contains(link, label) {
				return label
			}
			return link
		}

		return ""
	})
	text = strings.Replace(text, "&lt;", "<", -1)
	text = strings.Replace(text, "&gt;", ">", -1)
	return strings.Replace(text, "&amp;", "&", -1)
}

func (f formatter) flatten(m *slack.MessageEvent) string {
	text := m.Text
	for _, a := range m.Attachments {
		text = text + "\n" + a.Fallback
	}
	return text
}
