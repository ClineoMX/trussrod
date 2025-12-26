package notifications

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

type Twilio struct {
	client     *http.Client
	accountSid string
	authToken  string
	fromNumber string
}

func NewTwilioClient(accountSid, authToken, fromNumber string) *Twilio {
	return &Twilio{
		client:     &http.Client{},
		accountSid: accountSid,
		authToken:  authToken,
		fromNumber: fromNumber,
	}
}

func (t *Twilio) SendWhatsAppMessage(ctx context.Context, notification *WhatsAppMessage) error {
	addr := fmt.Sprintf("https://api.twilio.com/2010-04-01/Accounts/%s/Messages.json", t.accountSid)

	variables, err := json.Marshal(notification.Variables)
	if err != nil {
		return err
	}

	to := fmt.Sprintf("whatsapp:%s", notification.To)
	body := url.Values{
		"From":             {t.fromNumber},
		"To":               {to},
		"ContentSid":       {notification.TemplateID},
		"ContentVariables": {string(variables)},
	}

	req, err := http.NewRequestWithContext(ctx, "POST", addr, strings.NewReader(body.Encode()))
	if err != nil {
		return err
	}

	req.SetBasicAuth(t.accountSid, t.authToken)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := t.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("failed to send WhatsApp message: %s", resp.Status)
	}

	return nil
}
