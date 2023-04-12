package main

import (
	"fmt"
	"github.com/go-ini/ini"
	"github.com/sirupsen/logrus"
	"logcogo/common"
	"logcogo/etcd"
	"logcogo/kafka"
	"logcogo/tailf"
)

//日志收集客户端
// 开源项目：filebeat go开发的ES收集 yaml做的配置项
// 我们做的是etcd 热加载

// 收集指定目录下日志文件，发送到kafka中

// 技能包 kafka发数据、tail读日志文件

type Config struct {
	KafkaConfig `ini:"kafka"`
	Etcd        `ini:"etcd"`
}
type KafkaConfig struct {
	Address  string `ini:"address"`
	ChanSize int64  `ini:"chan_size"`
}
type Etcd struct {
	Address    string `ini:"address"`
	CollectKey string `ini:"collect_key"`
}

func main() {
	// 获取本机的IP
	ip, err := common.GetOutboundIP()
	if err != nil {
		logrus.Errorf("get local ip address failed, err: %v\n", err)
		return
	}
	//0. 读配置文件 `go-ini` 高效的go配置文件操作库
	var config = new(Config)
	err = ini.MapTo(config, "./conf/conf.ini")
	if err != nil {
		logrus.Error("ini open failed, err: %v", err)
		return
	}
	logrus.Info("read conf file successfully")
	//1. 初始化工作
	// 连接kafka
	err = kafka.Init([]string{config.KafkaConfig.Address}, config.ChanSize)
	if err != nil {
		logrus.Error("init kafka failed, err:%v", err)
		return
	}

	logrus.Info("init kafka successfully")

	//2. 初始化ETCD

	// 连接etcd
	err = etcd.Init([]string{config.Etcd.Address})
	if err != nil {
		logrus.Errorf("etcd init failed, err: %v\n", err)
		return
	}

	//从etcd中拉取要收集的日志项
	collectKey := fmt.Sprintf(config.Etcd.CollectKey, ip)
	logrus.Infof("colelct key: %s\n", collectKey)
	collectConf, err := etcd.GetConf(collectKey)
	if err != nil {
		logrus.Errorf("get collect config failed, err: %v\n", err)
		return
	}

	fmt.Printf("collectConf: %v\n", collectConf)

	// 监控etcd中key的变化
	go etcd.WatchConf(collectKey)

	logrus.Info("init etcd successfully")

	//3. tailfile启动多个tailTask去采集日志 并发送到kafka
	err = tailf.Init(collectConf)
	if err != nil {
		logrus.Errorf("tail to collect log failed, err: %v", err)
		return
	}
	run()
}

// 等待 让所有的tailTask都在执行
func run() {
	select {}
}
