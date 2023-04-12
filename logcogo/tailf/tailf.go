package tailf

import (
	"context"
	"github.com/Shopify/sarama"
	"github.com/hpcloud/tail"
	"github.com/sirupsen/logrus"
	"logcogo/kafka"
	"strings"
	"time"
)

// tail相关

type tailTask struct {
	path     string
	topic    string
	instance *tail.Tail
	ctx      context.Context
	cancel   context.CancelFunc
}

func (t *tailTask) Init(tailConfig tail.Config) (err error) {
	t.instance, err = tail.TailFile(t.path, tailConfig)
	return
}
func (t *tailTask) run() {
	logrus.Infof("path:%v is running to collect log...\n", t.path)
	for {
		select {
		case <-t.ctx.Done():
			logrus.Warningf("tailTask Path: %v is stopping...\n", t.path)
			return
		case msg, ok := <-t.instance.Lines:
			if !ok {
				logrus.Error("tail file close reopen, fileName: %v", t.instance.Filename)
				time.Sleep(time.Second) //读取错误 等待1s
				continue
			}
			msg.Text = strings.Trim(msg.Text, "\r")
			logrus.Info("msg: ", msg.Text)
			// 去除换行符
			if len(msg.Text) == 0 {
				continue
			}
			// 利用通道 异步发送数据
			sendMsg := &sarama.ProducerMessage{
				Topic:     t.topic,
				Value:     sarama.StringEncoder(msg.Text),
				Timestamp: time.Now(),
			}
			// 将数据扔进kafka
			kafka.ToMsgChan(sendMsg)
		}

	}
}
