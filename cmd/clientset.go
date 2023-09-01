package cmd

import (
	"flag"
	"path/filepath"
	"strings"

	kafka "github.com/segmentio/kafka-go"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

var eventFlag = InitFlags()

func InitFlags() *EventFlags {
	var eventFlags EventFlags

	if home := homedir.HomeDir(); home != "" {
		flag.StringVar(&eventFlags.KubeConfig, "kubernetes_kubeConfig", filepath.Join(home, ".kube", "config"), "(可选) kubeconfig 文件的绝对路径")
	} else {
		flag.StringVar(&eventFlags.KubeConfig, "kubernetes_kubeConfig", eventFlags.KubeConfig, "kubeconfig 文件的绝对路径")
	}

	flag.StringVar(&eventFlags.AddrList, "kafka_address", eventFlags.AddrList, "(可选) kafka address list 地址列表")
	flag.StringVar(&eventFlags.Topic, "kafka_topic", eventFlags.Topic, "(可选) kafka 主题名")

	flag.Parse()
	return &eventFlags
}

func InitClient() (*kubernetes.Clientset, error) {
	var err error
	var config *rest.Config

	if config, err = rest.InClusterConfig(); err != nil {
		if config, err = clientcmd.BuildConfigFromFlags("", eventFlag.KubeConfig); err != nil {
			panic(err.Error())
		}
	}
	kubeclient, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return kubeclient, nil
}

func InitKafka() *kafka.Writer {
	brokerList := strings.Split(eventFlag.AddrList, ",")
	return &kafka.Writer{
		Addr:         kafka.TCP(brokerList...),
		Topic:        eventFlag.Topic,
		Balancer:     &kafka.LeastBytes{},
		RequiredAcks: kafka.RequireAll,
		Async:        true,
	}
}
