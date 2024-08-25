package service

const (
	LogQueueName = "queue:log"
)

// func NewRabbitMQConn() (*amqp.Connection, error) {
// 	conn, err := amqp.Dial(config.AMQP_URL)
// 	if err != nil {
// 		return nil, err
// 	}
// 	util.RegisterCleanUp("rabbitmq", func() error {
// 		return conn.Close()
// 	})
// 	return conn, nil
// }

// func NewRabbitMQChannel(conn *amqp.Connection, queue string) (*amqp.Channel, error) {
// 	ch, err := conn.Channel()
// 	if err != nil {
// 		return nil, err
// 	}
// 	connectChannelToQueue(ch, queue)
// 	return ch, nil
// }

// func connectChannelToQueue(ch *amqp.Channel, queueName string) amqp.Queue {
// 	q, err := ch.QueueDeclare(
// 		queueName, // name
// 		true,      // durable
// 		false,     // delete when unused
// 		false,     // exclusive
// 		false,     // no-wait
// 		nil,       // arguments
// 	)
// 	if err != nil {
// 		panic(err)
// 	}
// 	return q
// }
