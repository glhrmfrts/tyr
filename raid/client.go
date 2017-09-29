package raid

import (
	"fmt"
	"log"
	"net"
	//"bufio"
	"time"

	d "github.com/tj/go-debug"
	utcode "github.com/glhrmfrts/go-utcode"
)

const (
	ReadBufferSize = 4 * 1024
)

var (
	debug = d.Debug("raid")
)

type Client struct {
	conn       net.Conn
	requests   map[Etag]chan *Message
	notifyRead chan bool
}

func (c *Client) Conn() net.Conn {
	return c.conn
}

func (c *Client) Request(msg *Message) (chan *Message, error) {
	err := c.Send(msg)
	if err != nil {
		return nil, err
	}
	c.notifyRead <- true

	ch := make(chan *Message, 1)
	c.requests[msg.Etag()] = ch

	return ch, nil
}

func (c *Client) Send(msg *Message) error {
	encode, err := utcode.Encode(msg)
	if err != nil {
		return err
	}

	encodestr := string(encode)
	debug("send %s", encodestr)

	fmt.Fprintf(c.conn, encodestr)
	return nil
}

func (c *Client) listen() error {
	readBuf := make([]byte, ReadBufferSize)

	go func() {
		for notification := range c.notifyRead {
			log.Println("WILL READ", notification)

			size := ReadBufferSize
			content := make([]byte, 0, size)
			c.conn.SetReadDeadline(time.Now().Add(5 * time.Second))

			for size == ReadBufferSize {
				var err error
				size, err = c.conn.Read(readBuf)

				if err != nil {
					log.Println(err)
				}

				content = append(content, readBuf[0:size]...)
			}

			c.onMessage(content)
		}
	}()

	return nil
}

func (c *Client) onMessage(content []byte) {
	debug("receive %s", string(content))

	var msg Message
	err := utcode.Decode(content, &msg)
	if err != nil {
		log.Println(err)
		return
	}

	if ch, ok := c.requests[msg.Etag()]; ok {
		ch <- &msg
	} else {
		log.Println("unknown etag ", msg.Etag())
	}
}

func NewClient(mode, addr string) (*Client, error) {
	conn, err := net.Dial(mode, addr)
	if err != nil {
		return nil, err
	}

	c := &Client{
		requests:   make(map[Etag]chan *Message),
		conn:       conn,
		notifyRead: make(chan bool),
	}

	err = c.listen()
	if err != nil {
		return nil, err
	}

	return c, nil
}
