package _robot

import (
	"testing"
	"time"
)

func TestNewDingTalkRobotText(t *testing.T) {
	robot := NewDingTalkRobot(&DingTalkConfig{
		Duration:   6 * time.Second,
		Title:      "系统预警",
		Webhook:    "https://oapi.dingtalk.com/robot/send?access_token=767c18a0c45c610d4b44af399261ee06860b444db4e3de32eea73462a90eb849",
		SignSecret: "SEC93dd4c5817ea6578eaef0b7fcfde7f5802dcf0876444c6fe7d417dae8fb6a58c",
	})
	robot.Push(DingTalkMessageText{
		Content: "#### 杭州天气 @150XXXXXXXX \n > 9度，西北风1级，空气良89，相对温度73%\n > ![screenshot](https://img.alicdn.com/tfs/TB1NwmBEL9TBuNjy1zbXXXpepXa-2400-1218.png)\n > ###### 10点20分发布 [天气](https://www.dingtalk.com) \n",
	})
	robot.Listen()
}

func TestNewDingTalkRobotMarkdown(t *testing.T) {
	robot := NewDingTalkRobot(&DingTalkConfig{
		Webhook:    "https://oapi.dingtalk.com/robot/send?access_token=767c18a0c45c610d4b44af399261ee06860b444db4e3de32eea73462a90eb849",
		SignSecret: "SEC93dd4c5817ea6578eaef0b7fcfde7f5802dcf0876444c6fe7d417dae8fb6a58c",
	})
	
	robot.Push(&DingTalkMessage{
		MsgType: Markdown,
		//At: &DingTalkAt{
		//	AtMobiles: []string{"18817392521"},
		//},
		Markdown: &DingTalkMessageMarkdown{
			Title: "你好明天",
			Text:  "#### 杭州天气 @150XXXXXXXX \n > 9度，西北风1级，空气良89，相对温度73%\n > ![screenshot](https://img.alicdn.com/tfs/TB1NwmBEL9TBuNjy1zbXXXpepXa-2400-1218.png)\n > ###### 10点20分发布 [天气](https://www.dingtalk.com) \n",
		},
	})
	robot.Listen()
}
