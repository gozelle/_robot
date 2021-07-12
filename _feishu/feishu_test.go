package _feishu

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNewFeiShuRobot(t *testing.T) {
	robot := NewFeiShuRobot("https://open.feishu.cn/open-apis/bot/v2/hook/07012df3-efe7-49b4-a3f4-ea70fc011f71", "mzKfBasPYWi42Wv5mzTkxh")
	err := robot.SendText("Hello world!")
	require.NoError(t, err, err)
}
