package queue

type Queue interface {
	Topic(topic string) (Topic, error)
}

type ConsumerFunc func(message interface{}) error

type Topic interface {
	Publish(message interface{}) error
	Consume(target interface{}, consumer ConsumerFunc) error
}
