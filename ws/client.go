package ws

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"time"

	"gitlab.com/pbernier3/orax-cli/common"

	"github.com/cenkalti/backoff"
	"github.com/gorilla/websocket"
	"github.com/spf13/viper"
)

var (
	log = common.GetLog()
)

func exponentialBackOff() *backoff.ExponentialBackOff {
	b := &backoff.ExponentialBackOff{
		InitialInterval:     500 * time.Millisecond,
		RandomizationFactor: 0.5,
		Multiplier:          2,
		MaxInterval:         1 * time.Minute,
		MaxElapsedTime:      72 * time.Hour,
		Clock:               backoff.SystemClock,
	}
	b.Reset()
	return b
}

type Client struct {
	id       string
	Received chan []byte
	done     chan struct{}
	conn     *websocket.Conn
}

var u, _ = url.Parse("ws://localhost:8080/miner")

func init() {
	if os.Getenv("ORAX_ORCHESTRATOR_ENDPOINT") != "" {
		url, err := url.ParseRequestURI(os.Getenv("ORAX_ORCHESTRATOR_ENDPOINT"))
		if err != nil {
			log.Fatalf("Failed to parse ORAX_ORCHESTRATOR_ENDPOINT: %s", err)
		}
		u = url
	}
}

func (cli *Client) connect() {
	id := viper.GetString("miner_id")
	minerSecret := viper.GetString("miner_secret")
	log.Infof("Connecting to Orax as [%s]...", id)

	expBackOff := exponentialBackOff()
	header := http.Header{
		"Authorization": []string{id + ":" + minerSecret},
		"Version":       []string{common.Version[1:]}}

	err := backoff.RetryNotify(func() error {
		d := websocket.Dialer{
			Proxy:             http.ProxyFromEnvironment,
			HandshakeTimeout:  45 * time.Second,
			EnableCompression: true}
		c, resp, err := d.Dial(u.String(), header)

		if resp != nil {
			if resp.StatusCode == 400 {
				bytes, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					return backoff.Permanent(fmt.Errorf("Unexpected error: %s", err))
				}
				msg := string(bytes)
				return backoff.Permanent(errors.New(msg))
			} else if resp.StatusCode == 401 {
				return backoff.Permanent(errors.New("Failed to authenticate with orax orchestrator"))
			} else if resp.StatusCode == 409 {
				return backoff.Permanent(errors.New("Already connected with the same miner id"))
			}
		}

		cli.conn = c
		return err
	}, expBackOff, func(err error, duration time.Duration) {
		log.Warnf("Failed to connected. Retrying in %s", duration)
	})

	if err != nil {
		log.Fatalf("Failed to connect: %s", err)
	}

	log.Info("Connected to Orax orchestrator")
}

func NewWebSocketClient() (cli *Client) {
	cli = new(Client)
	cli.done = make(chan struct{})
	cli.Received = make(chan []byte)

	return cli
}

func (cli *Client) Start() {
	cli.connect()
	go cli.read()
}

func (cli *Client) Stop() {
	log.Info("Stopping websocket client...")

	// connection can be nil if we are trying to re-connect to the server
	if cli.conn != nil {

		// Cleanly close the connection by sending a close message and then
		// waiting (with timeout) for the server to close the connection.
		err := cli.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		if err != nil {
			log.Error("Failed to gracefully disconnect: ", err)
			return
		}

		// the `done` channel is being closed by the read function
		select {
		case <-cli.done:
		case <-time.After(2 * time.Second):
		}
	}
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
