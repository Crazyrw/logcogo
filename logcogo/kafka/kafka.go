package kafka

import (
	"github.com/Shopify/sarama"
	"github.com/sirupsen/logrus"
)

var (
	client  sarama.SyncProducer
	msgChan chan *sarama.ProducerMessage
)

func Init(addressList []string, chanSize int64) (err error) {
	// 生产者配置
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll          // 发送完数据需要leader和follow都确认
	config.Producer.Partitioner = sarama.NewRandomPartitioner // 新选出一个partition
	config.Producer.Return.Successes = true                   // 成功交付的消息将在success channel返回

	// 连接kafka
	client, err = sarama.NewSyncProducer(addressList, config)

	msgChan = make(chan *sarama.ProducerMessage, chanSize)

	//启动一个后台的goroutine 从chan中读取数据 发送到kafka
	go sendMsg()
	return
}

// 从通道中读取消息 发送到kafka
func sendMsg() {
	for {
		select {
		case msg := <-msgChan:
			partition, offset, err := client.SendMessage(msg)
			if err != nil {
				logrus.Warning("send msg failed, err:", err)
				return
			}
			logrus.Infof("send msg to kafka successfully. partition: %v, offset: %v", partition, offset)
		}
	}
}

func ToMsgChan(msg *sarama.ProducerMessage) {
	msgChan <- msg
}
