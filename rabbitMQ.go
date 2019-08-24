package barrage

import (
	"database/sql"
	"encoding/json"
	_ "github.com/go-sql-driver/mysql"
	"github.com/streadway/amqp"
	"log"
	"os"
)

func SendToMQ(body []byte) (rs bool) {
	rs = false

	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	CheckError(err, "Can't connect to RabbitMQ")
	defer conn.Close()

	channel, err := conn.Channel()
	CheckError(err, "Can't create a channel")
	defer channel.Close()

	queue, err := channel.QueueDeclare(
		"order",
		true,
		false,
		false,
		false,
		nil)
	CheckError(err, "Can't declare a queue")


	err = channel.Publish("",
		queue.Name,
		false,
		false,
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "application/json",
			Body:         body,
		})
	CheckError(err, "Can't publish a message")

	log.Printf("send: %s", string(body))

	rs = true
	return
}

func Consumer() {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	CheckError(err, "Can't connect to RabbitMQ")
	defer conn.Close()

	channel, err := conn.Channel()
	CheckError(err, "Can't create a channel")
	defer channel.Close()

	queue, err := channel.QueueDeclare(
		"order",
		true,
		false,
		false,
		false,
		nil)
	CheckError(err, "Can't declare a queue")

	err = channel.Qos(1, 0, false)
	CheckError(err, "Can't configure the QoS")

	messageChannel, err := channel.Consume(
		queue.Name,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	CheckError(err, "Can't register a consumer")

	stopChan := make(chan bool)
	go func() {
		log.Printf("Consumer ready, PID: %d", os.Getpid())
		for d := range messageChannel {
			log.Printf("Received a message: %s", string(d.Body))
			var barrage Barrage
			err = json.Unmarshal(d.Body, &barrage)
			CheckError(err, "Error encoding JSON")


			barrage.AddBarrage()


			if err := d.Ack(false); err != nil {
				log.Printf("Error acknowledging message : %s", err)
			} else {
				log.Printf("Acknowledged message")
			}
		}
	}()

	<-stopChan
}

func (b Barrage) AddBarrage() {
	db, err := sql.Open("mysql", "root:123@tcp(127.0.0.1:3306)/barrage")
	CheckError(err, "Can't connect to mysql")
	defer db.Close()
	_, err = db.Exec("INSERT INTO barrage(message,color,userId)VALUES (?,?,?)", b.Message,b.Color,b.UserId)
	CheckError(err, "Can't add barrage")
}

func CheckError(err error, msg string) {
	if err != nil {
		log.Println(err)
		defer log.Println(msg)
		panic(err)
	}
}
