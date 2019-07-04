package ws

import (
	"net/http"
	"net/url"
	"time"

	_log "gitlab.com/pbernier3/orax-cli/log"

	"github.com/cenkalti/backoff"
	"github.com/gorilla/websocket"
)

var (
	log = _log.New("ws")
)

func exponentialBackOff() *backoff.ExponentialBackOff {
	b := &backoff.ExponentialBackOff{
		InitialInterval:     500 * time.Millisecond,
		RandomizationFactor: 0.5,
		Multiplier:          1.5,
		MaxInterval:         10 * time.Second,
		MaxElapsedTime:      10 * time.Minute,
		Clock:               backoff.SystemClock,
	}
	b.Reset()
	return b
}

type Client struct {
	id       string
	stop     chan struct{}
	Received chan []byte
	done     chan struct{}
	conn     *websocket.Conn
}

var u = url.URL{Scheme: "ws", Host: "localhost:8080", Path: "/miner"}

func (cli *Client) connect() {
	log.Infof("Connecting to Orax as [%s]...", cli.id)

	err := backoff.Retry(func() error {
		m := http.Header{"Authorization": []string{cli.id}}
		d := websocket.Dialer{
			Proxy:             http.ProxyFromEnvironment,
			HandshakeTimeout:  45 * time.Second,
			EnableCompression: true}
		c, _, err := d.Dial(u.String(), m)
		cli.conn = c
		return err
	}, exponentialBackOff())

	if err != nil {
		log.Fatal("dial:", err)
	}

	log.Info("Connected to Orax orchestrator")
}

func (cli *Client) Start(id string) {
	cli.id = id
	cli.connect()

	cli.stop = make(chan struct{})
	cli.done = make(chan struct{})
	cli.Received = make(chan []byte, 256)

	go cli.read()

	for {
		select {
		case <-cli.done:
			return
		case <-cli.stop:
			log.Info("Stopping websocket client")

			// Cleanly close the connection by sending a close message and then
			// waiting (with timeout) for the server to close the connection.
			err := cli.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Error("Failed to gracefully disconnect: ", err)
				return
			}
			select {
			case <-cli.done:
			case <-time.After(time.Second):
			}
			return
		}
	}
}

func (cli *Client) Stop() {
	close(cli.stop)
	<-cli.done
	<-cli.Received
}

func (cli *Client) read() {
	defer close(cli.done)
	defer close(cli.Received)

	for {
		_, message, err := cli.conn.ReadMessage()

		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseNormalClosure) {
				log.Error("Unexpected disconnection from server: ", err)
				cli.connect()
			} else {
				// If it was a gracefull closure, exit the loop
				break
			}
		}
		if len(message) > 0 {
			cli.Received <- message
		}
	}
}

func (cli *Client) Send(message []byte) {
	go func() {
		err := cli.conn.WriteMessage(websocket.BinaryMessage, message)
		if err != nil {
			log.Error("write:", err)
		}
	}()
}
