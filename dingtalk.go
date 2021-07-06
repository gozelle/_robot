package _robot

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/koyeo/goutils/_http"
	"log"
	"net/url"
	"strconv"
	"time"
)

const (
	Text       = "text"
	Markdown   = "markdown"
	Link       = "link"
	ActionCard = "actionCard"
	FeedCard   = "feedCard"
)

type DingTalkConfig struct {
	Duration   time.Duration `json:"duration"`
	Title      string        `json:"title"`
	Webhook    string        `json:"webhook"`
	SignSecret string        `json:"sign_secret"`
}

type DingTalkMessage struct {
	MsgType    string                   `json:"msgtype"` // required
	At         *DingTalkAt              `json:"at,omitempty"`
	Text       *DingTalkMessageText     `json:"text,omitempty"`
	Link       *DingTalkMessageLink     `json:"link,omitempty"`
	Markdown   *DingTalkMessageMarkdown `json:"markdown,omitempty"`
	ActionCard *DingTalkActionCard      `json:"actionCard,omitempty"`
}

type DingTalkAt struct {
	AtMobiles []string `json:"atMobiles,omitempty"` // @用户的手机号
	AtUserIds []string `json:"atUserIds,omitempty"` // @人的用户ID
	IsAtAll   bool     `json:"isAtAll,omitempty"`   // 是否 @所有人
}

type DingTalkMessageText struct {
	Content string `json:"content"` // required
}

type DingTalkMessageLink struct {
	Title      string `json:"title"`      // required
	Text       string `json:"text"`       // required
	MessageUrl string `json:"messageUrl"` // required
	PicUrl     string `json:"picUrl,omitempty"`
}

type DingTalkMessageMarkdown struct {
	Title string `json:"title"` // required
	Text  string `json:"text"`  // required
}

type DingTalkActionCard struct {
	Title          string `json:"title"`       // required
	Text           string `json:"text"`        // required
	SingleTitle    string `json:"singleTitle"` // required
	SingleURL      string `json:"singleURL"`   // required
	BtnOrientation string `json:"btnOrientation,omitempty"`
}

func NewDingTalkRobot(config *DingTalkConfig) *DingTalkRobot {
	if config.Duration == 0 {
		config.Duration = 1 * time.Second
	}
	robot := &DingTalkRobot{
		config: config,
		bucket: NewBucket(config.Duration),
	}
	return robot
}

type DingTalkRobot struct {
	bucket         *Bucket
	config         *DingTalkConfig
	titleFormatter func(messages []interface{}) string
	at             *DingTalkAt
}

func (p *DingTalkRobot) Bucket() *Bucket {
	return p.bucket
}

func (p *DingTalkRobot) Push(message interface{}) {
	p.bucket.Push(message)
}

func (p *DingTalkRobot) Listen() {
	p.Bucket().PopTimely(func(messages []interface{}) {
		l := len(messages)
		if l == 0 {
			return
		}
		var title string
		if p.titleFormatter != nil {
			title = p.titleFormatter(messages)
		} else if p.config.Title != "" {
			title = p.config.Title
		} else {
			if l > 1 {
				title = fmt.Sprintf("%d messages", l)
			} else {
				title = fmt.Sprintf("%d messages", l)
			}
		}
		err := p.Request(title, p.PrepareMarkdown(messages))
		if err != nil {
			log.Println(err)
			return
		}
	})
}

func (p *DingTalkRobot) SetTitleFormatter(format func(messages []interface{}) string) {
	p.titleFormatter = format
}

func (p *DingTalkRobot) PrepareMarkdown(messages []interface{}) *DingTalkMessageMarkdown {
	msg := new(DingTalkMessageMarkdown)
	
	for _, v := range messages {
		switch v.(type) {
		case *DingTalkMessageMarkdown:
			item := v.(*DingTalkMessageMarkdown)
			msg.Text += fmt.Sprintf("## %s\n%s\n", item.Title, item.Text)
		case DingTalkMessageMarkdown:
			item := v.(DingTalkMessageMarkdown)
			msg.Text += fmt.Sprintf("## %s\n%s\n", item.Title, item.Text)
		case *DingTalkMessageText:
			item := v.(*DingTalkMessageText)
			msg.Text += fmt.Sprintf("%s\n", item.Content)
		case DingTalkMessageText:
			item := v.(DingTalkMessageText)
			msg.Text += fmt.Sprintf("%s\n", item.Content)
		case *DingTalkActionCard:
			item := v.(*DingTalkActionCard)
			msg.Text += fmt.Sprintf("##%s\n[%s](%s)\n%s\n", item.Title, item.SingleTitle, item.SingleURL, item.Text)
		case DingTalkActionCard:
			item := v.(DingTalkActionCard)
			msg.Text += fmt.Sprintf("##%s\n[%s](%s)\n%s\n", item.Title, item.SingleTitle, item.SingleURL, item.Text)
		case *DingTalkMessageLink:
			item := v.(*DingTalkMessageLink)
			msg.Text += fmt.Sprintf("##%s\n![](%s)[%s](%s)\n%s\n", item.Title, item.PicUrl, item.MessageUrl, item.MessageUrl, item.Text)
		case DingTalkMessageLink:
			item := v.(DingTalkMessageLink)
			msg.Text += fmt.Sprintf("##%s\n![](%s)[%s](%s)\n%s\n", item.Title, item.PicUrl, item.MessageUrl, item.MessageUrl, item.Text)
		default:
			d, err := json.Marshal(v)
			if err != nil {
				msg.Text += fmt.Sprintf("%+v\n", v)
			} else {
				msg.Text += fmt.Sprintf("%s\n", string(d))
			}
		}
	}
	
	return msg
}

func (p *DingTalkRobot) sign() (timestamp int64, sign string) {
	timestamp = time.Now().UnixNano() / 1e6
	str := fmt.Sprintf("%d\n%s", timestamp, p.config.SignSecret)
	h := hmac.New(sha256.New, []byte(p.config.SignSecret))
	h.Write([]byte(str))
	sign = base64.StdEncoding.EncodeToString(h.Sum(nil))
	return
}

func (p *DingTalkRobot) Request(title string, msg *DingTalkMessageMarkdown) (err error) {
	timestamp, sign := p.sign()
	
	req := _http.NewRequest()
	req.SetContentType(_http.APPLICATION_JSON)
	
	address, err := url.Parse(p.config.Webhook)
	if err != nil {
		log.Println(err)
		return
	}
	query := address.Query()
	query.Add("timestamp", strconv.FormatInt(timestamp, 10))
	query.Add("sign", sign)
	address.RawQuery = query.Encode()
	msg.Title = title
	_, err = req.Post(address.String(), &DingTalkMessage{
		MsgType:  Markdown,
		At:       p.at,
		Markdown: msg,
	})
	if err != nil {
		return
	}
	return
}
