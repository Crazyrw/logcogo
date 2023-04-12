package tailf

import (
	"context"
	"fmt"
	"github.com/hpcloud/tail"
	"github.com/sirupsen/logrus"
	"logcogo/common"
	"strings"
)

type tailTaskManager struct {
	tailTaskMap      map[string]*tailTask      //key: path_topic 所有 tailTask任务
	collectEntryList []common.CollectConf      //所有配置项
	confChan         chan []common.CollectConf //等待新配置的通道
}

var (
	ttManager  *tailTaskManager
	tailConfig tail.Config
)

func Init(allConf []common.CollectConf) (err error) {
	tailConfig = tail.Config{
		ReOpen:    true,
		MustExist: false,
		Poll:      true, //轮询
		Follow:    true,
		Location:  &tail.SeekInfo{Offset: 0, Whence: 2}, //从哪个地方开始读数据 2表示从文件的末尾去读
	}
	// 初始化tailTask管理者
	ttManager = &tailTaskManager{
		tailTaskMap:      make(map[string]*tailTask, 20),
		collectEntryList: allConf,
		confChan:         make(chan []common.CollectConf),
	}

	fmt.Println(allConf)
	// 打开文件 开始读取数据
	for _, conf := range allConf {
		ctx, cancel := context.WithCancel(context.Background())
		task := &tailTask{
			path:   conf.Path,
			topic:  conf.Topic,
			ctx:    ctx,
			cancel: cancel,
		}
		err = task.Init(tailConfig)
		if err != nil {
			logrus.Errorf("create tailTask failed, path: %v, err %v\n", task.path, err)
			continue
		}
		logrus.Infof("create a tailTask for path:%v\n successfully", task.path)

		// 登记task
		ttManager.tailTaskMap[task.path+"_"+task.topic] = task
		// 开始收集日志
		go task.run()
	}
	go ttManager.watch() // 后台等待新的配置的到来

	return
}

func (tt *tailTaskManager) watch() {
	// 需要循环等待 不能只等一次
	for {
		// 阻塞 等待新的配置到来
		newConf := <-tt.confChan
		logrus.Infof("get new conf from etcd, conf:%v\n", newConf)

		//管理task
		for _, conf := range newConf {
			if tt.isExist(conf) {
				continue
			}
			// 不存在 创建新的
			ctx, cancel := context.WithCancel(context.Background())
			task := &tailTask{
				path:   conf.Path,
				topic:  conf.Topic,
				ctx:    ctx,
				cancel: cancel,
			}
			err := task.Init(tailConfig)
			if err != nil {
				logrus.Errorf("create tailTask failed, path: %v, err %v\n", task.path, err)
				continue
			}
			logrus.Infof("create a tailTask for path:%v\n successfully", task.path)

			// 登记task
			ttManager.tailTaskMap[task.path+"_"+task.topic] = task
			// 开始收集日志
			go task.run()
		}

		//原来有 现在需要停掉
		for key, task := range ttManager.tailTaskMap {
			find := false
			for _, conf := range newConf {
				if strings.Compare(key, conf.Path+"_"+conf.Topic) == 0 {
					find = true
					break
				}
			}
			if !find {
				// 停掉tailTask
				logrus.Infof("tailTask Path: %v is need to stop.\n", task.path)
				delete(ttManager.tailTaskMap, key)
				task.cancel()
			}
		}

	}

}
func (tt *tailTaskManager) isExist(conf common.CollectConf) (ok bool) {
	_, ok = tt.tailTaskMap[conf.Path+"_"+conf.Topic]
	return
}
func SendNewConf(newConf []common.CollectConf) {
	ttManager.confChan <- newConf
}

//func (tt *tailTaskManager) createTailTask
