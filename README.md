# soc-eventer
`soc-eventer` 旨在对 kubernetes 原生功能的补充和强化
- 提供一个控制器收集 `Event` 信息到 `Kafka` 中实现 `Event` 数据持久化
  - 可通过 `go run main.go -h` 查看参数
```
go run main.go -h                                                              
Usage of /var/folders/h7/dnghff2d3f11mkbsrnqp75880000gn/T/go-build3024363050/b001/exe/main:
  -kafka_address string
        (可选) kafka address list 地址列表
  -kafka_topic string
        (可选) kafka 主题名
  -kubernetes_kubeConfig string
        (可选) kubeconfig 文件的绝对路径 (default "C:\\Users\\xxx\\.kube\\config")
```