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
	queue       string
	routingKey  string
	requests    map[string]chan *rmq.Message
	channelPool sync.Pool
	chPublish   *rmq.Channel
}

const (
	exchangeTopic = "rfour"
	exchangeHeaders = "rfour_headers"
)

func datetime() string {
	return time.Now().UTC().Format(time.RFC3339)
}

func (io *IO) Request(exchange, routingKey string, body []byte, headers map[string]interface{}) (chan *rmq.Message, error) {
	etag := raid.NewEtag().String()
	result := io.channelPool.Get().(chan *rmq.Message)

	headers["rfour"] = true
	headers["request"] = true
	headers["request-etag"] = etag
	headers["request-datetime"] = datetime()
	headers["request-reply-exchange"] = exchangeTopic
	headers["request-reply-rk"] = io.routingKey
	err := rmq.BasicPublish(
		io.chPublish,
		exchange,
		routingKey,
		body,
		headers,
		rmq.Mandatory,
	)
	if err != nil {
		return nil, err
	}

	io.requests[etag] = result
	log.Printf("request %s", etag)
	return result, nil
}

func (io *IO) finishRequest(etag string, msg *rmq.Message) bool {
	responseChannel, ok := io.requests[etag]
	if !ok {
		return false
	}

	responseChannel <- msg
	delete(io.requests, etag)
	io.channelPool.Put(responseChannel)
	log.Printf("reply %s", etag)
	return true
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

	if !io.finishRequest(responseEtag, msg) {
		log.Println("invalid input message")
		return
	}
	msg.Ack(true)
}

func NewIO(conn *rmq.RMQ) *IO {
	uuid := raid.NewEtag().String()
	io := &IO{
		conn:        conn,
		uuid:        uuid,
		queue:       fmt.Sprintf("rfour_input_%s", uuid),
		routingKey:  fmt.Sprintf("rfour.input.%s", uuid),
		requests:    make(map[string]chan *rmq.Message),
		channelPool: sync.Pool{
			New: func() interface{} {
				log.Println("Creating chan")
				return make(chan *rmq.Message)
			},
		},
	}
	chPublish, err := conn.Channel("publish", 1)
	if err != nil {
		log.Fatal(err)
	}
	io.chPublish = chPublish

	returnChan := make(chan rmq.Return)
	chPublish.NotifyReturn(returnChan)
	go func() {
		for ret := range returnChan {
			log.Println("Not routed")
			etag := ret.Headers["request-etag"].(string)
			msg := &rmq.Message{
				Headers: map[string]interface{}{
					"__err__": "NOT_ROUTED",
				},
			}
			io.finishRequest(etag, msg)
		}
	}()

	chAnnounce, err := conn.Channel("announce", 1)
	if err != nil {
		log.Fatal(err)
	}

	rmq.ExchangeDeclare(chAnnounce, exchangeTopic, "topic", rmq.Durable)
	rmq.ExchangeDeclare(chAnnounce, exchangeHeaders, "headers", rmq.Durable)
	rmq.QueueDeclare(chAnnounce, io.queue, rmq.Exclusive)
	rmq.QueueBind(chAnnounce, io.queue, io.routingKey, exchangeTopic, map[string]interface{}{})

	chConsume, err := conn.Channel("consume", 1)
	if err != nil {
		log.Fatal(err)
	}
	rmq.BasicConsume(chConsume, io.queue, io.uuid, func(msg *rmq.Message) {
		handleInputMessage(io, msg)
	})

	return io
}

func mapContains(m map[string]interface{}, key string) bool {
	_, ok := m[key]
	return ok
}

func ValidHeaders(headers map[string]interface{}) bool {
	return (
		mapContains(headers, "rfour") &&
		mapContains(headers, "request") &&
		mapContains(headers, "request-etag") &&
		mapContains(headers, "request-reply-exchange") &&
		mapContains(headers, "request-reply-rk"))
}

func Reply(body []byte, headers map[string]interface{}, ch *rmq.Channel) error {
	if ValidHeaders(headers) {
		headers["response"] = true
		headers["response-datetime"] = datetime()
		headers["response-etag"] = headers["request-etag"]

		return rmq.BasicPublish(
			ch,
			headers["request-reply-exchange"].(string),
			headers["request-reply-rk"].(string),
			body,
			headers,
			0,
		)
	}
	return fmt.Errorf("invalid headers")
}
