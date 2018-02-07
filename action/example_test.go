package action_test

import (
	"fmt"

	"github.com/botopolis/bot"
	"github.com/botopolis/slack/action"
	oslack "github.com/nlopes/slack"
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
			Params: oslack.PostMessageParameters{
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
		if len(a.Actions) < 0 {
			return
		}
		if a.Actions[0].Value == "true" {
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
