package cmd

import (
	// "bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	kafka "github.com/segmentio/kafka-go"

	// "github.com/elastic/go-elasticsearch/v6"
	// "github.com/elastic/go-elasticsearch/v6/esapi"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"
)

// 这段代码是一个使用 Go 编程语言编写的控制器（Controller）结构体，用于处理队列中的任务。它似乎与缓存、工作队列和 Elasticsearch 客户端等内容有关。indexer：一个缓存索引器。queue：一个工作队列，使用速率限制接口进行限制。informer：一个缓存控制器。es：一个 Elasticsearch 客户端。
type Controller struct {
	indexer  cache.Indexer
	queue    workqueue.RateLimitingInterface
	informer cache.Controller
	kafka    *kafka.Writer
	// es       *elasticsearch.Client
}

// 代码中定义了一个用于创建 Controller 实例的函数：// 代码中定义了一个用于创建 Controller 实例的函数：代码中定义了一个用于创建 Controller 实例的函数：
func NewController(indexer cache.Indexer, queue workqueue.RateLimitingInterface, informer cache.Controller, kafka *kafka.Writer) *Controller {
	return &Controller{
		queue:    queue,
		indexer:  indexer,
		informer: informer,
		kafka:    kafka,
		// es:       es,
	}
}

// 代码中定义了一个名为 runWorker 的方法，该方法会在一个无限循环中调用 processNextItem 方法来处理队列中的任务。
func (c *Controller) runWorker() {
	for c.processNextItem() {
	}
}

// 代码中定义了 processNextItem 方法，它的作用是从队列中获取下一个任务的键（key），然后同步数据到标准输出，并处理可能出现的错误。
func (c *Controller) processNextItem() bool {
	key, quit := c.queue.Get()
	if quit {
		return false
	}
	defer c.queue.Done(key)

	err := c.syncToStdout(key.(string))
	c.handleErr(err, key)
	return true
}

// 在这个函数中，c 是一个控制器对象，用于处理事件同步的逻辑。它首先从存储中获取特定键的事件对象。如果事件对象存在，它将构造一个新的结构体 EventResp 并将事件对象的属性映射到这个结构体中。然后，它将结构体序列化为 JSON 数据，并创建一个 Elasticsearch 索引请求，将 JSON 数据存储到名为 "event" 的索引中。最后，函数会打印一条消息，表示数据插入操作已完成。
func (c *Controller) syncToStdout(key string) error {

	// 通过键获取事件对象
	obj, exists, err := c.indexer.GetByKey(key)
	if err != nil {
		klog.Errorf("Fetching object with key %s from store failed with %v", key, err)
		return err
	}

	// 检查事件对象是否存在
	if !exists {
		fmt.Printf("Event %s does not exists anymore\n", key)
	} else {

		// 构造 EventResp 结构体，将事件对象的属性映射到结构体字段
		eventResp := &EventResp{
			NAMESPACE:     obj.(*v1.Event).InvolvedObject.Namespace,
			NAME:          obj.(*v1.Event).InvolvedObject.Name,
			UID:           obj.(*v1.Event).InvolvedObject.UID,
			LastTimestamp: obj.(*v1.Event).LastTimestamp.UTC().Add(8 * time.Hour),
			Type:          obj.(*v1.Event).Type,
			Reason:        obj.(*v1.Event).Reason,
			Object:        fmt.Sprintf("%s/%s", obj.(*v1.Event).InvolvedObject.Kind, obj.(*v1.Event).InvolvedObject.Name),
			Count:         obj.(*v1.Event).Count,
			Message:       obj.(*v1.Event).Message,
		}

		// 将结构体序列化为 JSON 数据
		data, err := json.Marshal(eventResp)
		if err != nil {
			log.Fatalf("Error marshaling document: %s", err)
			return err
		}

		// 创建 Kafka 消息体
		messages := kafka.Message{
			Value: data,
		}

		const retries = 3
		// 重试3次
		for i := 0; i < retries; i++ {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			err = c.kafka.WriteMessages(ctx, messages)
			if errors.Is(err, kafka.LeaderNotAvailable) || errors.Is(err, context.DeadlineExceeded) {
				time.Sleep(time.Millisecond * 250)
				continue
			}

			if err != nil {
				log.Fatalf("kafka write failed, unexpected error %v", err)
			}
			break
		}

		fmt.Printf("插入数据完成  %s\n", obj.(*v1.Event).InvolvedObject.Name)
		// 关闭Writer
		// if err := c.kafka.Close(); err != nil {
		// 	log.Fatal("failed to close writer:", err)
		// }
		// 创建 ES 索引请求
		// req := esapi.IndexRequest{
		// 	Index:      "event",
		// 	DocumentID: string(obj.(*v1.Event).UID),
		// 	Body:       bytes.NewReader(data),
		// 	Refresh:    "true",
		// }

		// 执行索引请求
		// res, err := req.Do(context.Background(), c.es)
		// if err != nil {
		// 	log.Fatalf("Error getting response: %s", err)
		// 	return err
		// }
		// defer res.Body.Close()
		// fmt.Printf("插入数据完成  %s\n", obj.(*v1.Event).InvolvedObject.Name)
	}
	return nil
}

// handleErr 方法：这个方法用于处理错误。代码中的函数签名为 func (c *Controller) handleErr(err error, key interface{})。它接受一个错误（err）和一个键（key）作为输入参数。
// 首先，它检查如果错误为 nil，则直接将该键从队列中移除，然后返回。否则，如果某个键的重新排队次数小于 5 次，它会使用日志记录错误信息，并将该键添加到带有速率限制的队列中。之后，该方法返回。如果某个键的重新排队次数达到 5 次或以上，它会将该键从队列中移除，然后调用 runtime.HandleError 处理该错误，并在日志中记录丢弃该键的信息。
func (c *Controller) handleErr(err error, key interface{}) {
	if err == nil {
		c.queue.Forget(key)
		return
	}

	if c.queue.NumRequeues(key) < 5 {
		klog.Infof("Error syncing pod %v: %v", key, err)

		c.queue.AddRateLimited(key)
		return
	}

	c.queue.Forget(key)
	runtime.HandleError(err)
	klog.Infof("Dropping pod %q out of the queue: %v", key, err)
}

// Run 方法：这个方法用于运行控制器的主循环。代码中的函数签名为 func (c *Controller) Run(threadiness int, stopCh chan struct{})。它接受线程数（threadiness）和停止通道（stopCh）作为输入参数。
// 在函数开始部分，使用 defer 关键字注册了两个延迟函数调用，分别是 runtime.HandleCrash() 和 c.queue.ShutDown()。然后，使用日志记录控制器开始运行的信息。接着，通过启动一个 goroutine 来运行事件通知器（informer）。使用 cache.WaitForCacheSync 函数等待缓存同步完成，如果同步超时，则记录错误并返回。接下来，通过一个循环，在多个 goroutine 中运行 runWorker 方法，每个 goroutine 之间的间隔为一秒。最后，通过阻塞从停止通道（stopCh）接收信号来等待控制器停止，一旦接收到信号，记录停止信息并结束主循环。
func (c *Controller) Run(threadiness int, stopCh chan struct{}) {
	defer runtime.HandleCrash()

	defer c.queue.ShutDown()
	klog.Infof("Starting Event Controller")

	go c.informer.Run(stopCh)

	if !cache.WaitForCacheSync(stopCh, c.informer.HasSynced) {
		runtime.HandleError(fmt.Errorf("Timed out waiting for caches to sync"))
		return
	}

	for i := 0; i < threadiness; i++ {
		go wait.Until(c.runWorker, time.Second, stopCh)
	}
	<-stopCh

	klog.Info("Stopping Pod controller")
}
