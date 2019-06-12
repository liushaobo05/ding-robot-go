package gdingrobot

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

const (
	msgTypeText       = "text"
	msgTypeLink       = "link"
	msgTypeMarkdown   = "markdown"
	msgTypeActionCard = "actionCard"
)

// 参考：https://open-doc.dingtalk.com/microapp/serverapi2/qf2nxq
type Roboter interface {
	SendText(content string, atMobiles []string, isAtAll bool) error
	SendLink(title, text, messageURL, picURL string) error
	SendMarkdown(title, text string, atMobiles []string, isAtAll bool) error
	SendActionCard(title, text, singleTitle, singleURL, btnOrientation, hideAvatar string) error
}

type DingRobot struct {
	webhook   string
	client    *http.Client
	count     uint
	startTime time.Time
}

type errResponse struct {
	Errcode int
	Errmsg  string
}

func NewDingRobot(webhook string) DingRobot {
	return DingRobot{
		webhook:   webhook,
		count:     0,
		client:    &http.Client{},
		startTime: time.Now(),
	}
}

// 发送执行
func (d *DingRobot) do(msg interface{}) error {
	// 限流处理
	d.count += 1
	if (d.count - 20) >= 0 {
		subs := time.Now().Sub(d.startTime)
		if subs.Seconds() < 60 {
			time.Sleep(60)
		}
		d.startTime = time.Now()
		d.count = 0
	}

	m, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", d.webhook, bytes.NewReader(m))
	req.Header.Set("Content-Type", "application/json")

	if err != nil {
		return err
	}
	resp, err := d.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var dr errResponse
	err = json.Unmarshal(data, &dr)
	if err != nil {
		return err
	}
	if dr.Errcode != 0 {
		return fmt.Errorf("dingrobot send failed: %v", dr.Errmsg)
	}

	return nil
}

// 文本
func (d *DingRobot) SendText(content string, atMobiles []string, isAtAll bool) error {
	msg := &textMessage{
		MsgType: msgTypeText,
		Text: textParams{
			Content: content,
		},
		At: atParams{
			AtMobiles: atMobiles,
			IsAtAll:   isAtAll,
		},
	}
	return d.do(msg)
}

// link
func (d *DingRobot) SendLink(title, text, messageURL, picURL string) error {
	msg := &linkMessage{
		MsgType: msgTypeLink,
		Link: linkParams{
			Title:      title,
			Text:       text,
			MessageURL: messageURL,
			PicURL:     picURL,
		},
	}
	return d.do(msg)
}

// markdown
func (d *DingRobot) SendMarkdown(title, text string, atMobiles []string, isAtAll bool) error {
	msg := &markdownMessage{
		MsgType: msgTypeMarkdown,
		Markdown: markdownParams{
			Title: title,
			Text:  text,
		},
		At: atParams{
			AtMobiles: atMobiles,
			IsAtAll:   isAtAll,
		},
	}
	return d.do(msg)
}

// actionCard
func (d *DingRobot) SendActionCard(title, text, singleTitle, singleURL, btnOrientation, hideAvatar string) error {
	msg := &actionCardMessage{
		MsgType: msgTypeActionCard,
		ActionCard: actionCardParams{
			Title:          title,
			Text:           text,
			SingleTitle:    singleTitle,
			SingleURL:      singleURL,
			BtnOrientation: btnOrientation,
			HideAvatar:     hideAvatar,
		},
	}
	return d.do(msg)
}
