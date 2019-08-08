package ws

import (
	"encoding/hex"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"sync"
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
	id          string
	Received    chan []byte
	done        chan struct{}
	conn        *websocket.Conn
	NoncePrefix []byte // Not the best place to store but simple and convenient
	sendMux     sync.Mutex
}

var orchestratorURL string

func init() {
	// Override endpoint with env variable
	if os.Getenv("ORAX_ORCHESTRATOR_ENDPOINT") != "" {
		_, err := url.ParseRequestURI(os.Getenv("ORAX_ORCHESTRATOR_ENDPOINT"))
		if err != nil {
			log.Fatalf("Failed to parse ORAX_ORCHESTRATOR_ENDPOINT: %s", err)
		}
		orchestratorURL = os.Getenv("ORAX_ORCHESTRATOR_ENDPOINT")
	} else if orchestratorURL == "" {
		// If not set at build time fallback to local dev endpoint
		orchestratorURL = "ws://localhost:8077/miner"
	}
}

func (cli *Client) connect() {
	id := viper.GetString("miner_id")
	minerSecret := viper.GetString("miner_secret")
	log.Infof("Connecting to Orax as [%s]...", id)

	header := http.Header{
		"Authorization": []string{id + ":" + minerSecret},
		"Version":       []string{common.Version[1:]}}

	var noncePrefix []byte
	err := backoff.RetryNotify(func() error {
		d := websocket.Dialer{
			Proxy:             http.ProxyFromEnvironment,
			HandshakeTimeout:  45 * time.Second,
			EnableCompression: true}
		c, resp, err := d.Dial(orchestratorURL, header)

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

			// Decode assigned nonce prefix
			noncePrefix, err = hex.DecodeString(resp.Header.Get("NoncePrefix"))
			if err != nil {
				return backoff.Permanent(fmt.Errorf("Failed to get nonce prefix: %s", err))
			}
		}

		cli.conn = c
		cli.NoncePrefix = noncePrefix
		return err
	}, exponentialBackOff(), func(err error, duration time.Duration) {
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
		cli.sendMux.Lock()
		err := cli.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		cli.sendMux.Unlock()

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

func (cli *Client) Connected() bool {
	return cli.conn != nil
}

func (cli *Client) Send(message []byte) {
	cli.sendMux.Lock()
	err := cli.conn.WriteMessage(websocket.BinaryMessage, message)
	cli.sendMux.Unlock()

	if err != nil {
		log.Error("write:", err)
	}

}
