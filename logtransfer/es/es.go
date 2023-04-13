package es

import (
	"context"
	"github.com/olivere/elastic/v7"
	"github.com/sirupsen/logrus"
)

type ESClient struct {
	client   *elastic.Client
	index    string // 数据库名称
	dataChan chan interface{}
}

var (
	esClient *ESClient
)

func Init(address, index string, chanSize int64) (err error) {
	client, err := elastic.NewClient(elastic.SetURL("http://" + address))
	if err != nil {
		logrus.Errorf("connect elastic failed, %v\n", err)
		return
	}

	esClient = &ESClient{
		client:   client,
		index:    index,
		dataChan: make(chan interface{}, chanSize),
	}

	//启动一个后台协程从通道中读取数据，然后将数据写入es
	go sendToES()
	return
}
func sendToES() {

	for {
		select {
		case msg := <-esClient.dataChan:
			do, err := esClient.client.Index().Index(esClient.index).BodyJson(msg).Do(context.Background())
			if err != nil {
				logrus.Errorf("write data to es failed, %v\n", err)
				continue
			}
			logrus.Infof("do.Id: %s, do.Type: %s, do.Index: %s\n", do.Id, do.Type, do.Index)
		}
	}

}

func SendToESChan(msg interface{}) {
	esClient.dataChan <- msg
}