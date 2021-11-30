package queue

type Queue interface {
	Topic(topic string) (Topic, error)
}

type Topic interface {
	Publish(message interface{}) error
}
