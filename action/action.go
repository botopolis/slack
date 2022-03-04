package action

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"

	"github.com/botopolis/bot"
	"github.com/slack-go/slack"
)

// Plugin conforms to the botopolis/bot.Plugin interface
type Plugin struct {
	*registry
	// Path at which our webhook sits.
	Path string
	// Signing secret to verify message comes from slack.
	SigningSecret string

	logger bot.Logger

	// Skip Verifying Header in tests as verifications are based on time
	skipVerifyHeader bool
}

// New returns a new plugin taking arguments for path and token
func New(path, signingSecret string) *Plugin {
	return &Plugin{
		registry:      &registry{},
		Path:          path,
		SigningSecret: signingSecret,
	}
}

// Load installs the webhook
func (p Plugin) Load(r *bot.Robot) {
	p.logger = r.Logger
	r.Router.HandleFunc(p.Path, p.webhook)
}

func (p Plugin) webhook(w http.ResponseWriter, r *http.Request) {
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	r.Body = ioutil.NopCloser(bytes.NewBuffer(b))
	p.logger.Debugf("slack/action: Received webhook to %s\n", p.Path)

	if !p.skipVerifyHeader {
		if err := p.verify(r.Header, b); err != nil {
			p.logger.Errorf("slack/action: Invalid webhook: %v\n", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}

	jsonBody := []byte(r.FormValue("payload"))
	var cb slack.AttachmentActionCallback
	if err := json.Unmarshal(jsonBody, &cb); err != nil {
		p.logger.Errorf("slack/action: Invalid webhook: %v\n", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	go p.Run(cb)
}

func (p Plugin) verify(h http.Header, body []byte) error {
	if h["X-Slack-Signature"] == nil || h["X-Slack-Request-Timestamp"] == nil {
		return errors.New("Missing signing headers")
	}

	verifier, err := slack.NewSecretsVerifier(h, p.SigningSecret)
	if err != nil {
		return err
	}

	if _, err := verifier.Write(body); err != nil {
		return err
	}

	return verifier.Ensure()
}
