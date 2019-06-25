package ws

import (
	"net/http"
	"net/url"
	"time"

	_log "gitlab.com/pbernier3/orax-cli/log"

	"github.com/gorilla/websocket"
)

var (
	log = _log.New("ws")
)

type Client struct {
	stop     chan struct{}
	Received chan []byte
	done     chan struct{}
	conn     *websocket.Conn
}

func (cli *Client) Start(id string) {
	u := url.URL{Scheme: "ws", Host: "localhost:8080", Path: "/miner"}
	log.Infof("Connecting to Orxas as [%s]", id)

	m := http.Header{"Authorization": []string{id}}
	d := websocket.Dialer{
		Proxy:             http.ProxyFromEnvironment,
		HandshakeTimeout:  45 * time.Second,
		EnableCompression: true}
	c, _, err := d.Dial(u.String(), m)

	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()
	cli.conn = c

	cli.stop = make(chan struct{})
	cli.done = make(chan struct{})
	cli.Received = make(chan []byte, 256)

	log.Info("Connection to Orax established")
	go cli.read()

	for {
		select {
		case <-cli.done:
			return
		case <-cli.stop:
			log.Info("Stopping websocket client")

			// Cleanly close the connection by sending a close message and then
			// waiting (with timeout) for the server to close the connection.
			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Error("write close:", err)
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
				log.Error("Read:", err)
			}
			break
		}

		cli.Received <- message
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
