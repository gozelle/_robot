package _feishu

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

type FeiShuImage struct {
	ImageKey string `json:"image_key"`
}

type FeiShuPostText struct {
	Text     string `json:"text"`
	UnEscape bool   `json:"un_escape"`
}

type FeiShuPostA struct {
	Text     string `json:"text"`
	UnEscape bool   `json:"un_escape"`
	Href     string `json:"href"`
}

type FeiShuPostAt struct {
	UserId string `json:"user_id"`
}

type FeiShuPostImg struct {
	ImageKey string `json:"image_key"`
	Height   int    `json:"height"`
	Width    int    `json:"width"`
}

func NewFeiShuRobot(webhook string, secret string) *FeiShuRobot {
	return &FeiShuRobot{webhook: webhook, secret: secret}
}

type FeiShuRobot struct {
	webhook string
	secret  string
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
	return p.Send(msg)
}

func (p *FeiShuRobot) SendPost(text string) (err error) {
	ts := time.Now().Unix()
	sign, err := p.Sign(p.secret, ts)
	if err != nil {
		return
	}
	msg := FeiShuMessage{
		MsgType:   "text",
		Timestamp: fmt.Sprintf("%d", ts),
		Sign:      sign,
		Content: FeiShuText{
			Text: text,
		},
	}
	return p.Send(msg)
}

func (p *FeiShuRobot) Send(msg interface{}) (err error) {
	resp := _http.Post(p.webhook, _http.Payload(msg))
	if resp.Error() != nil {
		err = resp.Error()
		return
	}
	res, err := resp.Json()
	if err != nil {
		err = fmt.Errorf("cast to json error: %s", err)
		return
	}
	message, err := res.Get("msg").String()
	if err != nil {
		err = fmt.Errorf("get code error: %s", err)
		return
	}
	if message != "ok" {
		err = fmt.Errorf("send error: %s", message)
	}
	return
}
