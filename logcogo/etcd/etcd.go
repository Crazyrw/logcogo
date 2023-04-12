package etcd

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	clientv3 "go.etcd.io/etcd/client/v3"
	"logcogo/common"
	"logcogo/tailf"
	"time"
)

var (
	client *clientv3.Client
)

func Init(address []string) (err error) {
	// client 使用grpc-go连接etcd 使用完必须得关闭它，不然可能会内存泄漏
	client, err = clientv3.New(clientv3.Config{
		Endpoints:   address,
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		logrus.Errorf("etcd init failed, err: %v\n", err)
		return
	}
	return
}

func GetConf(key string) (configs []common.CollectConf, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()
	getResp, err := client.Get(ctx, key)
	if err != nil {
		logrus.Errorf("get collect config from etcd failed, err: %v\n", err)
		return
	}
	if len(getResp.Kvs) == 0 {
		logrus.Warningf("collect config len is 0\n")
		return
	}
	res := getResp.Kvs[0]
	err = json.Unmarshal(res.Value, &configs)
	if err != nil {
		logrus.Errorf("json unmarshal failed, err: %v\n", err)
		return
	}
	return
}

// 监控日志收集项配置的变化

func WatchConf(key string) {
	// 创建watcher
	watcher := clientv3.NewWatcher(client)
	watchRespChan := watcher.Watch(context.Background(), key)

	for watchResp := range watchRespChan {
		for _, event := range watchResp.Events {
			var newConf []common.CollectConf
			fmt.Printf("type: %s, key: %s, value: %s\n", event.Type, event.Kv.Key, event.Kv.Value)
			if event.Type == clientv3.EventTypeDelete {
				// 说明删除了整个key 也就是所有的配置项
				logrus.Infof("delete collect conf key!!\n")
				tailf.SendNewConf(newConf) //会阻塞 等待tail取新的配置
			}
			err := json.Unmarshal(event.Kv.Value, &newConf)
			if err != nil {
				logrus.Errorf("json.Unmarshal failed, err: %v\n", err)
				continue
			}
			// 新的配置来了 返回一个只读通道
			tailf.SendNewConf(newConf) //会阻塞 等待tail取新的配置
		}
	}
}
