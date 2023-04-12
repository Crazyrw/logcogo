日志采集项目

LogAgent + SysAgent + LogTransfer

LogAgent: 数据采集，发送到kafka

技能点: goroutine、Context、kafka、ES

LogTransfer: 消费kafka中的日志数据到ES

SysAgent: 采集系统性能指标，将数据存到InfluxDB, 并通过Grafana进行可视化