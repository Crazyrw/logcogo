package main

import (
	"github.com/go-ini/ini"
	"github.com/sirupsen/logrus"
	"logtransfer/es"
	"logtransfer/kafka"
	"logtransfer/model"
)

func main() {
	// 1. 加载配置文件
	var config = new(model.Config)
	err := ini.MapTo(config, "./conf/conf.ini")
	if err != nil {
		logrus.Errorf("load config failed, %v\n", err)
		return
	}

	logrus.Infof("config: %#v\n", config)

	// 2. 连接ES
	err = es.Init(config.ESConfig.Address, config.ESConfig.Index, config.ESConfig.ChanSize)
	if err != nil {
		logrus.Errorf("init es failed, %v\n", err)
		return
	}

	//3. 连接kafka
	err = kafka.Init([]string{config.KafkaConfig.Address}, config.KafkaConfig.Topic)
	if err != nil {
		logrus.Errorf("init kafka failed, %v\n", err)
		return
	}

	select {} //阻塞主协程
}
