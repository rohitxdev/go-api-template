package service

import (
	"fmt"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/diode"
)

/*----------------------------------- Log receiver ----------------------------------- */

// func saveLogsToDb(ctx context.Context, logCollection *mongo.Collection, logs *[]any) {
// 	_, err := logCollection.InsertMany(ctx, *logs, options.InsertMany())
// 	if err != nil {
// 		fmt.Println("could not save logs to database:", err.Error())
// 	}
// }

// func ListenLogQueue(bufferSize uint, flushTime time.Duration) {
// 	logMsgCh, err := ConsumerCh.Consume(LogQueueName, "", false, false, false, false, nil)
// 	if err != nil {
// 		panic(err)
// 	}
// 	docsBuffer := make([]any, 0, bufferSize)
// 	logCollection := repo.MongoDBClient.Database("logs").Collection("logs")
// 	ticker := time.NewTicker(flushTime)
// 	ctx := context.TODO()

// 	for {
// 		select {
// 		case msg := <-logMsgCh:
// 			if uint(len(docsBuffer)) >= bufferSize {
// 				saveLogsToDb(ctx, logCollection, &docsBuffer)
// 				docsBuffer = docsBuffer[:0]
// 			}
// 			var doc any
// 			if err := bson.UnmarshalExtJSON(msg.Body, true, &doc); err != nil {
// 				fmt.Println("coud not unmarshal:", err.Error())
// 			}
// 			docsBuffer = append(docsBuffer, doc)
// 			msg.Ack(false)
// 		case <-ticker.C:
// 			if len(docsBuffer) > 0 {
// 				saveLogsToDb(ctx, logCollection, &docsBuffer)
// 				docsBuffer = docsBuffer[:0]
// 			}
// 		}
// 	}
// }

// /*----------------------------------- Log publisher ----------------------------------- */

// type logger struct {
// 	logChannel   *amqp.Channel
// 	logQueueName string
// }

// func (q logger) Write(data []byte) (int, error) {
// 	err := q.logChannel.PublishWithContext(
// 		context.TODO(),
// 		"",
// 		q.logQueueName,
// 		false,
// 		false,
// 		amqp.Publishing{
// 			Body: data,
// 		},
// 	)
// 	if err != nil {
// 		return 0, err
// 	}
// 	return len(data), nil
// }

// var QueueLogger = func() *logger {
// 	if ConsumerCh != nil {
// 		return &logger{
// 			logChannel:   ConsumerCh,
// 			logQueueName: LogQueueName,
// 		}
// 	}
// 	return nil
// }()

var Logger = func() zerolog.Logger {
	asyncLogWriter := diode.NewWriter(os.Stdout, 10000, time.Millisecond*10, func(missed int) {
		fmt.Printf("Logger dropped %d messages\n", missed)
	})
	return zerolog.New(asyncLogWriter)
}()
