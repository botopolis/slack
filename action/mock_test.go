package action_test

import "github.com/botopolis/bot"

type ExampleChat struct{}

var exampleChan = make(chan bot.Message)

func (ExampleChat) Username() string             { return "" }
func (ExampleChat) Messages() <-chan bot.Message { return exampleChan }
func (ExampleChat) Load(*bot.Robot)              { close(exampleChan) }
func (ExampleChat) Send(bot.Message) error       { return nil }
func (ExampleChat) Reply(bot.Message) error      { return nil }
func (ExampleChat) Direct(bot.Message) error     { return nil }
func (ExampleChat) Topic(bot.Message) error      { return nil }
