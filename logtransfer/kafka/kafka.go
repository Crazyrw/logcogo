package kafka

import (
	"encoding/json"
	"fmt"
	"github.com/Shopify/sarama"
	"github.com/sirupsen/logrus"
	"logtransfer/es"
)

func Init(address []string, topic string) (err error) {
	consumer, err := sarama.NewConsumer(address, nil)
	if err != nil {
		logrus.Errorf("create a consumer failed, %v\n", err)
		return
	}

	//获取topic下的所有分区
	partitions, err := consumer.Partitions(topic)
	if err != nil {
		logrus.Errorf("get all partitions failed, %v\n", err)
		return
	}

	for partition := range partitions {
		// 针对每个分区 创建一个消费者
		var pc sarama.PartitionConsumer
		pc, err = consumer.ConsumePartition(topic, int32(partition), sarama.OffsetNewest)
		if err != nil {
			logrus.Errorf("create a consumer for a partition failed, %v\n", err)
			return
		}

		logrus.Infof("start to consume....\n")

		go func(partitionConsumer sarama.PartitionConsumer) {
			for msg := range pc.Messages() { //没有数据会阻塞在这
				var mm map[string]interface{}
				fmt.Println(msg.Topic, string(msg.Value))
				err := json.Unmarshal(msg.Value, &mm)
				if err != nil {
					logrus.Errorf("unmarshal msg failed, %v\n", err)
					continue
				}
				es.SendToESChan(mm)
			}
		}(pc)
	}
	return
}
