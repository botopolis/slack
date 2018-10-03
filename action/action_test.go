package action

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/botopolis/bot/mock"
	"github.com/nlopes/slack"
	"github.com/stretchr/testify/assert"
)

const (
	signingSecret = "e6b19c573432dcc6b075501d51b51bb8"

	fooBody      = `payload=%7B%22callback_id%22%3A%22foo%22%7D`
	fooSignature = "v0=d27668944a2857e8495256fc93c7aed9f1119617ec08902b56edf69862b16855"
	barBody      = `payload=%7B%22callback_id%22%3A%22bar%22%7D`
	barSignature = "v0=50179568ccb23da3e0dd88c0ac6da9e336ae9ed2895008d7d736807978e4e8bf"
)

func newHeader(signature string) http.Header {
	h := http.Header{}
	h.Set("Content-Type", "application/x-www-form-urlencoded")
	h.Set("X-Slack-Signature", signature)
	h.Set("X-Slack-Request-Timestamp", "1531431954")
	return h
}

func readCloser(b []byte) io.ReadCloser {
	return ioutil.NopCloser(bytes.NewReader(b))
}

var logger = mock.NewLogger()

func init() {
	logger.WriteFunc = func(l mock.Level, v ...interface{}) {
		fmt.Println(v...)
	}
	logger.WritefFunc = func(l mock.Level, msg string, v ...interface{}) {
		fmt.Printf(msg, v...)
	}
}

func TestWebhook_response(t *testing.T) {
	cases := []struct {
		Name   string
		Body   []byte
		Header http.Header
		Out    int
	}{
		{
			Name:   "With an empty header",
			Body:   []byte(`payload=%7B%22token%22%3A%22foo%22%7D`),
			Header: newHeader(""),
			Out:    http.StatusBadRequest,
		},
		{
			Name:   "With a non-JSON body",
			Body:   []byte(`<xml></xml>`),
			Header: newHeader("v0=242c2f40d58a5dbe4ae2d73ff61e07cf632a1d43a8e52a714e04e3fc4889cb7f"),
			Out:    http.StatusBadRequest,
		},
		{
			Name:   "With a bad header",
			Body:   []byte(`payload=%7B%22token%22%3A%22foo%22%7D`),
			Header: newHeader("bad header"),
			Out:    http.StatusBadRequest,
		},
		{
			Name:   "When everything's right",
			Body:   []byte(`payload=%7B%22token%22%3A%22foo%22%7D`),
			Header: newHeader("v0=9bff0c9ad4804e518f3bc03b7a8b3d3e360b78b368e601294574cc10878e7c13"),
			Out:    http.StatusOK,
		},
	}

	p := Plugin{SigningSecret: signingSecret, registry: &registry{}, logger: logger}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			p.webhook(recorder, &http.Request{
				Body:   readCloser(c.Body),
				Header: c.Header,
				Method: "POST",
			})
			assert.Equal(t, c.Out, recorder.Code)
		})
	}
}

func TestWebhook_callback(t *testing.T) {
	done := make(chan string, 3)
	fooReq := http.Request{
		Header: newHeader(fooSignature),
		Body:   readCloser([]byte(fooBody)),
		Method: "POST",
	}
	barReq := http.Request{
		Header: newHeader(barSignature),
		Body:   readCloser([]byte(barBody)),
		Method: "POST",
	}

	p := Plugin{SigningSecret: signingSecret, registry: &registry{}, logger: logger}
	p.Add("bar", func(slack.AttachmentActionCallback) { done <- "bar" })
	p.Add("foo", func(slack.AttachmentActionCallback) { done <- "foo" })

	p.webhook(httptest.NewRecorder(), &fooReq)
	assert.Equal(t, "foo", <-done)

	p.webhook(httptest.NewRecorder(), &barReq)
	assert.Equal(t, "bar", <-done)
}
