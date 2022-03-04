package action_test

import (
	"fmt"

	"github.com/botopolis/bot"
	"github.com/botopolis/slack"
	"github.com/botopolis/slack/action"
	oslack "github.com/slack-go/slack"
)

type ExamplePlugin struct{}

func (p ExamplePlugin) Load(r *bot.Robot) {
	fmt.Println("Loaded")

	var actions action.Plugin
	if ok := r.Plugin(&actions); !ok {
		r.Logger.Error("Example plugin requires slack/action.Plugin")
		return
	}

	r.Hear(bot.Regexp("trigger"), func(r bot.Responder) error {
		return r.Send(bot.Message{
			Params: slack.TextAttachmentParams{
				Attachments: []oslack.Attachment{{
					Text:       "Trigger example",
					CallbackID: "example",
					Actions: []oslack.AttachmentAction{{
						Name:  "check",
						Type:  "button",
						Text:  "Do it",
						Value: "true",
					}, {
						Name:  "check",
						Type:  "button",
						Text:  "Nah",
						Style: "danger",
						Value: "false",
					}},
				}},
			},
		})
	})

	// handle example callback ID with a function
	actions.Add("example", func(a oslack.AttachmentActionCallback) {
		attachmentActions := a.ActionCallback.AttachmentActions
		if len(attachmentActions) < 0 {
			return
		}
		if attachmentActions[0].Value == "true" {
			// do the thing
		}
	})
}

func Example() {
	bot.New(
		ExampleChat{},
		action.New("/interaction", "signing secret!"),
		ExamplePlugin{},
	).Run()
	// Output: Loaded
}
