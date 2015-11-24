package exchange

type Exchange struct {
	topics map[string]*Topic
}

type Subscription struct {
	inbox   chan<- interface{}
	Channel <-chan interface{}
}

type Topic struct {
	name          string
	subscriptions []Subscription
}

func New() Exchange {
	return Exchange{
		topics: make(map[string]*Topic),
	}
}

func (ex *Exchange) Subscribe(topic string) Subscription {
	t := ex.findOrCreateTopic(topic)
	s := t.subscribe()
	return s
}

func (ex *Exchange) Publish(topic string, message interface{}) {
	t := ex.findOrCreateTopic(topic)
	t.publish(message)
}

func (ex *Exchange) findOrCreateTopic(topic string) *Topic {
	if _, ok := ex.topics[topic]; !ok {
		ex.topics[topic] = &Topic{
			name:          topic,
			subscriptions: make([]Subscription, 0),
		}
	}
	t := ex.topics[topic]
	return t
}

func (s Subscription) Close() {}

func (t *Topic) subscribe() Subscription {
	channel := make(chan interface{})
	subscription := Subscription{
		inbox:   channel,
		Channel: channel,
	}
	t.subscriptions = append(t.subscriptions, subscription)
	return subscription
}

func (t Topic) publish(message interface{}) {
	for _, s := range t.subscriptions {
		var subscription = s
		go func() {
			subscription.inbox <- message
		}()
	}
}
