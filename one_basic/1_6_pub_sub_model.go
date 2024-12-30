package one_basic

import (
	"strings"
	"sync"
	"time"
)

// use golang concurrent model to build a publish/subscriber model

type Subscriber struct {
	name string
	flow chan any
}

type TopicFunc func(any) bool

type Publisher struct {
	buffer      int
	subscribers map[*Subscriber]TopicFunc
	mu          sync.RWMutex // 不需要显式初始化
	timeout     time.Duration
}

func NewPublisher(buffer int, timeout time.Duration) *Publisher {
	return &Publisher{
		buffer:      buffer,
		subscribers: make(map[*Subscriber]TopicFunc),
		timeout:     timeout,
	}
}

// AddSubscriber 添加不限定主题的订阅者
func (p *Publisher) AddSubscriber(topicName string) (*Subscriber, error) {
	return p.AddSubscriberWithTopic(topicName, nil)
}

// AddSubscriberWithTopic 添加限定主题的订阅者
func (p *Publisher) AddSubscriberWithTopic(topicName string, topic TopicFunc) (*Subscriber, error) {
	ch := make(chan any, p.buffer)
	s := &Subscriber{
		name: topicName,
		flow: ch,
	}
	// map竞态处理
	// 避免不可预知的结果
	// delete和send同时进行
	p.mu.Lock()
	p.subscribers[s] = topic
	p.mu.Unlock()
	return s, nil
}

// Evict 关闭发布
func (p *Publisher) Evict(subscriber *Subscriber) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	delete(p.subscribers, subscriber)
	close(subscriber.flow)
	return nil
}

// Close 关闭全部发布
func (p *Publisher) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()
	for subscriber := range p.subscribers {
		delete(p.subscribers, subscriber)
		close(subscriber.flow)
	}
}

// Publish 发布消息
// 无差别地向订阅者发布消息
func (p *Publisher) Publish(v any) error {
	// 使用读锁 防止修改即可 可以并发读
	p.mu.RLock()
	defer p.mu.RUnlock()

	var wg sync.WaitGroup
	for subscriber, topicFunc := range p.subscribers {
		wg.Add(1)
		go p.Send(subscriber, topicFunc, v, &wg)
	}

	wg.Wait()
	return nil
}

// Send 发送消息的具体实现
func (p *Publisher) Send(subscriber *Subscriber, topicFunc TopicFunc, v any, wg *sync.WaitGroup) {
	defer wg.Done()
	if topicFunc != nil && !topicFunc(v) {
		return
	}

	select {
	case subscriber.flow <- v:
	case <-time.After(p.timeout):
	}
}

func main() {
	pub := NewPublisher(1000, 1000*time.Millisecond)
	defer pub.Close()

	all, _ := pub.AddSubscriber("All")
	golang, _ := pub.AddSubscriberWithTopic("Golang", func(v any) bool {
		if s, ok := v.(string); ok {
			return strings.Contains(s, "Golang")
		}
		return false
	})

	pub.Publish("Hello World")
	pub.Publish("Hello Golang")

	go func() {
		for v := range all.flow {
			println("all 's msgs:", v.(string))
		}
	}()

	go func() {
		for v := range golang.flow {
			println("golang's msgss:", v.(string))
		}
	}()

	time.Sleep(3 * time.Second)
}
