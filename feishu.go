package _robot

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"github.com/koyeo/_http"
	"time"
)

type FeiShuMessage struct {
	MsgType   string      `json:"msg_type"` // required
	Content   interface{} `json:"content"`
	Timestamp string      `json:"timestamp"`
	Sign      string      `json:"sign"`
}

type FeiShuText struct {
	Text string `json:"text"`
}

type FeiShuRobot struct {
	webhook string
	secret  string
}

func NewFeiShuRobot(webhook string, secret string) *FeiShuRobot {
	return &FeiShuRobot{webhook: webhook, secret: secret}
}

func (p *FeiShuRobot) SendText(text string) (err error) {
	ts := time.Now().Unix()
	sign, err := p.Sign(p.secret, ts)
	if err != nil {
		return
	}
	msg := FeiShuMessage{
		MsgType: "text",
		Content: FeiShuText{
			Text: text,
		},
		Timestamp: fmt.Sprintf("%d", ts),
		Sign:      sign,
	}
	resp := _http.Post(p.webhook, _http.Payload(msg))
	if resp.Error() != nil {
		err = resp.Error()
		return
	}
	fmt.Println(resp.String())
	return
}

func (p *FeiShuRobot) Sign(secret string, timestamp int64) (string, error) {
	stringToSign := fmt.Sprintf("%v", timestamp) + "\n" + secret
	var data []byte
	h := hmac.New(sha256.New, []byte(stringToSign))
	_, err := h.Write(data)
	if err != nil {
		return "", err
	}
	signature := base64.StdEncoding.EncodeToString(h.Sum(nil))
	return signature, nil
}
