package rfour

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/glhrmfrts/tyr/rmq"
	"github.com/glhrmfrts/tyr/raid"
)

type IO struct {
	conn        *rmq.RMQ
	uuid        string
	requests    map[string]chan *rmq.Message
	channelPool sync.Pool
	chPublish   *rmq.Channel
}

const (
	exchangeTopic = "rfour"
	exchangeHeaders = "rfour_headers"
)

func (io *IO) Request(exchange, routingKey string, body []byte, headers map[string]interface{}) chan *rmq.Message {
	etag := raid.NewEtag().String()
	result := io.channelPool.Get().(chan *rmq.Message)

	headers["rfour"] = true
	headers["request-etag"] = etag
	io.chPublish.BasicPublish(
		exchange,
		routingKey,
		body,
		headers,
	)
	io.requests[etag] = result
	log.Printf("send %s at %s", etag, time.Now())
	return result
}

func handleInputMessage(io *IO, msg *rmq.Message) {
	if _, ok := msg.Headers["rfour"]; !ok {
		log.Println("invalid input message")
		return
	}
	if _, ok := msg.Headers["response"]; !ok {
		log.Println("invalid input message")
		return
	}

	responseEtag, ok := msg.Headers["response-etag"].(string)
	if !ok {
		log.Println("invalid input message")
		return
	}
	responseChannel, ok := io.requests[responseEtag]
	if !ok {
		log.Println("invalid input message")
		return
	}
	responseChannel <- msg
	delete(io.requests, responseEtag)
	io.channelPool.Put(responseChannel)
	log.Printf("reply %s at %s", responseEtag, time.Now())
}

func NewIO(conn *rmq.RMQ) *IO {
	io := &IO{
		conn: conn,
		uuid: raid.NewEtag().String(),
		channels: make(map[string]chan *rmq.Message),
		channelPool: sync.Pool{
			New: func() interface{} {
				return make(chan *rmq.Message)
			},
		},
	}
	chPublish, err := conn.Channel("publish", 1)
	if err != nil {
		log.Fatal(err)
	}
	io.chPublish = chPublish

	chAnnounce, err := conn.Channel("announce", 1)
	if err != nil {
		log.Fatal(err)
	}

	queue := fmt.Sprintf("rfour_input_%s", io.uuid)
	routingKey := fmt.Sprintf("rfour.input.%s", io.uuid)
	chAnnounce.ExchangeDeclare(exchangeTopic, "topic", rmq.Durable)
	chAnnounce.ExchangeDeclare(exchangeHeaders, "headers", rmq.Durable)
	chAnnounce.QueueDeclare(queue, rmq.Exclusive)
	chAnnounce.QueueBind(queue, routingKey, exchangeTopic, map[string]interface{}{})

	chConsume, err := conn.Channel("consume", 10)
	if err != nil {
		log.Fatal(err)
	}
	chConsume.BasicConsume(queue, io.uuid, func(msg *rmq.Message) {
		handleInputMessage(io, msg)
	})

	return io
}
