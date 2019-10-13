package ws

import (
	"encoding/hex"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"sync"
	"time"

	"gitlab.com/oraxpool/orax-cli/common"

	"github.com/cenkalti/backoff"
	"github.com/gorilla/websocket"
	"github.com/spf13/viper"
)

var (
	log = common.GetLog()
)

const redirectDurationLimit = 5 * time.Minute

func exponentialBackOff() *backoff.ExponentialBackOff {
	b := &backoff.ExponentialBackOff{
		InitialInterval:     500 * time.Millisecond,
		RandomizationFactor: 0.5,
		Multiplier:          2,
		MaxInterval:         1 * time.Minute,
		MaxElapsedTime:      time.Duration(1<<63 - 1), // For ever
		Clock:               backoff.SystemClock,
	}
	b.Reset()
	return b
}

type Client struct {
	id          string
	Received    chan []byte
	DoneReading chan struct{}
	Connected   chan *ConnectionInfo
	conn        *websocket.Conn
	sendMux     sync.Mutex
	NbSubMiners int
	Endpoint    string
}

type ConnectionInfo struct {
	NoncePrefix       []byte
	Target            uint64
	BatchingDuration  time.Duration
	InitialBatchDelay time.Duration
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

func NewWebSocketClient(nbSubMiners int) (cli *Client) {
	cli = new(Client)
	cli.NbSubMiners = nbSubMiners
	cli.DoneReading = make(chan struct{})
	cli.Received = make(chan []byte)
	cli.Connected = make(chan *ConnectionInfo)
	cli.Endpoint = orchestratorURL

	return cli
}

func (cli *Client) connect() {
	id := viper.GetString("miner_id")
	minerSecret := viper.GetString("miner_secret")
	log.Infof("Connecting to Orax as [%s]...", id)

	header := http.Header{
		"Authorization": []string{id + ":" + minerSecret},
		"Version":       []string{common.Version[1:]},
		"HashRate":      []string{strconv.FormatInt(common.GetIndicativeHashRate(cli.NbSubMiners), 10)},
	}

	var connectionInfo *ConnectionInfo
	retryStrategy := exponentialBackOff()
	err := backoff.RetryNotify(func() error {
		// If a redirection didn't allow the client to connect after a certain amount of time
		// reset the endpoint to the default
		// This prevents the client to be stuck for ever because of a faulty redirection
		if cli.Endpoint != orchestratorURL && retryStrategy.GetElapsedTime() > redirectDurationLimit {
			cli.Endpoint = orchestratorURL
			retryStrategy.Reset()
			log.Warnf("Resetting endpoint to the default [%s]", cli.Endpoint)
		}

		d := websocket.Dialer{
			Proxy:            http.ProxyFromEnvironment,
			HandshakeTimeout: 45 * time.Second}
		c, resp, err := d.Dial(cli.Endpoint, header)

		if err != nil {
			// Server responded but handshake failed
			if err == websocket.ErrBadHandshake {

				if resp == nil {
					return errors.New("Empty response")
				}

				// 3xx: Redirection
				if 300 <= resp.StatusCode && resp.StatusCode < 400 {
					if resp.Header.Get("Location") == "" {
						return errors.New("Location missing for redirection")
					}

					cli.Endpoint = resp.Header.Get("Location")
					retryStrategy.Reset()
					return fmt.Errorf("Redirecting to %s", cli.Endpoint)
				} else
				// 4xx: Rejected by the server (validation)
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
				} else
				// Unhandled errors
				if resp.StatusCode >= 400 {
					bytes, err := ioutil.ReadAll(resp.Body)
					if err != nil {
						return fmt.Errorf("Unexpected error: %s", err)
					}
					msg := string(bytes)
					return errors.New(msg)
				}
			} else {
				// Failed to connect
				return err
			}
		}

		if resp == nil {
			return errors.New("Empty response")
		}

		// Decode headers
		connectionInfo, err = parseConnectionHeaders(resp.Header)
		if err != nil {
			return err
		}

		cli.conn = c

		return nil
	}, retryStrategy, func(err error, duration time.Duration) {
		log.WithError(err).Warnf("Failed to connect. Retrying in %s", duration)
	})

	if err != nil {
		log.WithError(err).Fatal("Failed to connect")
	}

	cli.Connected <- connectionInfo
}

func parseConnectionHeaders(header http.Header) (*ConnectionInfo, error) {
	connectionInfo := new(ConnectionInfo)

	// NoncePrefix
	noncePrefix, err := hex.DecodeString(header.Get("NoncePrefix"))
	if err != nil {
		return nil, fmt.Errorf("Failed to get nonce prefix: %s", err)
	}
	connectionInfo.NoncePrefix = noncePrefix

	// Target
	targetStr := header.Get("Target")
	if targetStr != "" {
		target, err := strconv.ParseUint(targetStr, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("Failed to get target: %s", err)
		}
		connectionInfo.Target = target
	}

	// BatchingDuration
	bacthingDurationStr := header.Get("BatchingDuration")
	if bacthingDurationStr != "" {
		bacthingDuration, err := strconv.Atoi(bacthingDurationStr)
		if err != nil {
			log.WithField("value", bacthingDurationStr).Warn("Failed to parse BatchingDuration value from the server")
			connectionInfo.BatchingDuration = 60
		} else {
			connectionInfo.BatchingDuration = time.Duration(bacthingDuration) * time.Second
		}
	}

	// InitialBatchDelay
	initialBatchDelayStr := header.Get("InitialBatchDelay")
	if initialBatchDelayStr != "" {
		initialBatchDelay, err := strconv.Atoi(initialBatchDelayStr)
		if err != nil {
			log.WithField("value", initialBatchDelayStr).Warn("Failed to parse InitialBatchDelay value from the server")
			connectionInfo.InitialBatchDelay = 60
		} else {
			connectionInfo.InitialBatchDelay = time.Duration(initialBatchDelay) * time.Second
		}
	}

	return connectionInfo, nil
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
			log.WithError(err).Error("Failed to gracefully disconnect")
			return
		}

		// the `doneReading` channel is being closed by the read function
		select {
		case <-cli.DoneReading:
		case <-time.After(3 * time.Second):
			cli.clean()
		}
	}
}

func (cli *Client) clean() {
	close(cli.DoneReading)
	close(cli.Received)
	close(cli.Connected)
}

func (cli *Client) read() {
	defer cli.clean()

	for {
		_, message, err := cli.conn.ReadMessage()
		// TODO: this deadline should be readjusted after V1
		cli.conn.SetReadDeadline(time.Now().Add(18 * time.Minute))

		if err != nil {
			if e, ok := err.(*websocket.CloseError); ok && e.Code == websocket.CloseNormalClosure {
				if e.Text != "" {
					log.Warnf("Disconnection reason: %s", e.Text)
				}
				// If it was a gracefull closure, exit the loop
				break
			}
			log.WithError(err).Error("Unexpected failure to read from server")
			cli.connect()
		}
		if len(message) > 0 {
			cli.Received <- message
		}
	}
}

func (cli *Client) IsConnected() bool {
	return cli.conn != nil
}

func (cli *Client) Send(message []byte) {
	cli.sendMux.Lock()
	err := cli.conn.WriteMessage(websocket.BinaryMessage, message)
	cli.sendMux.Unlock()

	if err != nil {
		log.WithError(err).Error("Failure to send.")
	}

}
