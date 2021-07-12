package _dingtalk

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/koyeo/_bucket"
	"github.com/koyeo/_http"
	"log"
	"net/url"
	"strconv"
	"time"
)

const (
	mt_text        = "text"
	mt_markdown    = "markdown"
	mt_link        = "link"
	mt_action_card = "actionCard"
	mt_feed_card   = "feedCard"
)

type Config struct {
	Duration   time.Duration `json:"duration"`
	Title      string        `json:"title"`
	Webhook    string        `json:"webhook"`
	SignSecret string        `json:"sign_secret"`
}

type Message struct {
	MsgType    string      `json:"msgtype"` // required
	At         *At         `json:"at,omitempty"`
	Text       *Text       `json:"text,omitempty"`
	Link       *Link       `json:"link,omitempty"`
	Markdown   *Markdown   `json:"markdown,omitempty"`
	ActionCard *ActionCard `json:"actionCard,omitempty"`
}

type At struct {
	AtMobiles []string `json:"atMobiles,omitempty"` // @用户的手机号
	AtUserIds []string `json:"atUserIds,omitempty"` // @人的用户ID
	IsAtAll   bool     `json:"isAtAll,omitempty"`   // 是否 @所有人
}

type Text struct {
	Content string `json:"content"` // required
}

type Link struct {
	Title      string `json:"title"`      // required
	Text       string `json:"text"`       // required
	MessageUrl string `json:"messageUrl"` // required
	PicUrl     string `json:"picUrl,omitempty"`
}

type Markdown struct {
	Title string `json:"title"` // required
	Text  string `json:"text"`  // required
}

type ActionCard struct {
	Title          string `json:"title"`       // required
	Text           string `json:"text"`        // required
	SingleTitle    string `json:"singleTitle"` // required
	SingleURL      string `json:"singleURL"`   // required
	BtnOrientation string `json:"btnOrientation,omitempty"`
}

func NewRobot(config *Config) *Robot {
	if config.Duration == 0 {
		config.Duration = 1 * time.Second
	}
	robot := &Robot{
		config: config,
		bucket: _bucket.NewBucket(config.Duration),
	}
	return robot
}

type Robot struct {
	bucket         *_bucket.Bucket
	config         *Config
	titleFormatter func(messages []interface{}) string
	at             *At
}

func (p *Robot) Bucket() *_bucket.Bucket {
	return p.bucket
}

func (p *Robot) Push(message interface{}) {
	p.bucket.Push(message)
}

func (p *Robot) Listen() {
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

func (p *Robot) SetTitleFormatter(format func(messages []interface{}) string) {
	p.titleFormatter = format
}

func (p *Robot) PrepareMarkdown(messages []interface{}) *Markdown {
	msg := new(Markdown)
	
	for _, v := range messages {
		switch v.(type) {
		case *Markdown:
			item := v.(*Markdown)
			msg.Text += fmt.Sprintf("## %s\n%s\n", item.Title, item.Text)
		case Markdown:
			item := v.(Markdown)
			msg.Text += fmt.Sprintf("## %s\n%s\n", item.Title, item.Text)
		case *Text:
			item := v.(*Text)
			msg.Text += fmt.Sprintf("%s\n", item.Content)
		case Text:
			item := v.(Text)
			msg.Text += fmt.Sprintf("%s\n", item.Content)
		case *ActionCard:
			item := v.(*ActionCard)
			msg.Text += fmt.Sprintf("##%s\n[%s](%s)\n%s\n", item.Title, item.SingleTitle, item.SingleURL, item.Text)
		case ActionCard:
			item := v.(ActionCard)
			msg.Text += fmt.Sprintf("##%s\n[%s](%s)\n%s\n", item.Title, item.SingleTitle, item.SingleURL, item.Text)
		case *Link:
			item := v.(*Link)
			msg.Text += fmt.Sprintf("##%s\n![](%s)[%s](%s)\n%s\n", item.Title, item.PicUrl, item.MessageUrl, item.MessageUrl, item.Text)
		case Link:
			item := v.(Link)
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

func (p *Robot) sign() (timestamp int64, sign string) {
	timestamp = time.Now().UnixNano() / 1e6
	str := fmt.Sprintf("%d\n%s", timestamp, p.config.SignSecret)
	h := hmac.New(sha256.New, []byte(p.config.SignSecret))
	h.Write([]byte(str))
	sign = base64.StdEncoding.EncodeToString(h.Sum(nil))
	return
}

func (p *Robot) Request(title string, msg *Markdown) (err error) {
	timestamp, sign := p.sign()
	
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
	resp := _http.Post(address.String(), _http.Payload(&Message{
		MsgType:  mt_markdown,
		At:       p.at,
		Markdown: msg,
	}))
	if resp.Error() != nil {
		err = resp.Error()
		return
	}
	return
}
