package ws

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	"gitlab.com/oraxpool/orax-cli/common"

	"github.com/cenkalti/backoff"
	"github.com/gorilla/websocket"
	"github.com/spf13/viper"
)

var (
	log = common.GetLog()
)

const (
	redirectDurationLimit = 5 * time.Minute
	pingInterval          = 30 * time.Second
)

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
	NbSubMiners int
	Endpoint    string

	Send    chan []byte
	Receive chan []byte

	Connected    chan *ConnectionInfo
	Disconnected chan bool
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
	cli.Endpoint = orchestratorURL
	cli.NbSubMiners = nbSubMiners

	cli.Connected = make(chan *ConnectionInfo)
	cli.Disconnected = make(chan bool)
	cli.Receive = make(chan []byte)
	cli.Send = make(chan []byte)

	return cli
}

func (cli *Client) Start(stop <-chan struct{}) <-chan struct{} {
	done := make(chan struct{})

	go func() {
		defer func() {
			close(cli.Receive)
			close(cli.Send)
			close(cli.Connected)
			close(cli.Disconnected)
			close(done)
		}()

		conn := cli.connect(stop)
		doneReading := cli.readPump(conn)
		stopWrite := make(chan struct{})
		cli.writePump(conn, stopWrite)

		for {
			select {
			case err := <-doneReading:
				cli.Disconnected <- true
				close(stopWrite)
				conn.Close()

				if err != nil {
					conn = cli.connect(stop)
					doneReading = cli.readPump(conn)
					stopWrite = make(chan struct{})
					cli.writePump(conn, stopWrite)
				} else {
					// Graceful shutdown initiated by the server
					return
				}
			case <-stop:
				// Initiate graceful shutdown
				err := conn.WriteControl(
					websocket.CloseMessage,
					websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""),
					time.Now().Add(2*time.Second),
				)
				close(stopWrite)
				if err != nil {
					log.WithError(err).Error("Failed to gracefully disconnect")
					return
				}

				// Wait for the closing response from the server
				// to shutdown or timeout
				select {
				case <-doneReading:
				case <-time.After(2 * time.Second):
				}
				return
			}
		}
	}()

	return done
}

func (cli *Client) connect(stop <-chan struct{}) (conn *websocket.Conn) {
	id := viper.GetString("miner_id")
	minerSecret := viper.GetString("miner_secret")
	log.Infof("Connecting to Orax as [%s]...", id)

	header := http.Header{
		"Authorization": []string{id + ":" + minerSecret},
		"Version":       []string{common.Version[1:]},
		"HashRate":      []string{strconv.FormatInt(common.GetIndicativeHashRate(cli.NbSubMiners), 10)},
	}

	var connectionInfo *ConnectionInfo
	ctx, cancel := context.WithCancel(context.Background())
	retryStrategy := exponentialBackOff()
	retryWithContext := backoff.WithContext(retryStrategy, ctx)

	// This goroutine cancels the retries if the stop channel returns anything
	backoffOver := make(chan struct{})
	go func() {
		select {
		case <-stop:
			cancel()
		case <-backoffOver:
		}
	}()

	err := backoff.RetryNotify(func() error {
		// If a redirection didn't allow the client to connect after a certain amount of time
		// reset the endpoint to the default
		// This prevents the client to be stuck for ever because of a faulty redirection
		if cli.Endpoint != orchestratorURL && retryStrategy.GetElapsedTime() > redirectDurationLimit {
			cli.Endpoint = orchestratorURL
			retryWithContext.Reset()
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
					retryWithContext.Reset()
					return fmt.Errorf("Redirecting to %s", cli.Endpoint)
				} else
				// 4xx: Rejected by the server (validation)
				if resp.StatusCode == 400 {
					bytes, err := ioutil.ReadAll(resp.Body)
					if err != nil {
						return backoff.Permanent(fmt.Errorf("Failed to read response body: %s", err))
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
						return fmt.Errorf("Failed to read response body: %s", err)
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

		conn = c

		return nil
	}, retryWithContext, func(err error, duration time.Duration) {
		log.Warnf("Failed to connect. Retrying in %s", duration)
	})
	close(backoffOver)

	if err != nil {
		log.WithError(err).Fatal("Failed to connect")
	}

	cli.Connected <- connectionInfo
	return conn
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

func (cli *Client) readPump(conn *websocket.Conn) (doneReading chan error) {
	doneReading = make(chan error)

	go func() {
		defer close(doneReading)

		for {
			_, message, err := conn.ReadMessage()

			if err != nil {
				// Graceful disconnection
				if e, ok := err.(*websocket.CloseError); ok && e.Code == websocket.CloseNormalClosure {
					if e.Text != "" {
						log.Infof("Disconnection reason: %s", e.Text)
					}
				} else {
					log.WithError(err).Error("Unexpected error reading from server")
					doneReading <- err
				}
				return
			}
			if len(message) > 0 {
				cli.Receive <- message
			}
		}
	}()

	return doneReading
}

func (cli *Client) writePump(conn *websocket.Conn, stopWrite chan struct{}) {
	go func() {
		keepAliveTicker := time.NewTicker(pingInterval)

		defer func() {
			keepAliveTicker.Stop()
		}()

		for {
			select {
			case <-stopWrite:
				return
			case msg, ok := <-cli.Send:
				if !ok {
					return
				}
				err := conn.WriteMessage(websocket.BinaryMessage, msg)
				if err != nil {
					log.WithError(err).Error("Failed to send.")
				}

			case <-keepAliveTicker.C:
				if err := conn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(2*time.Second)); err != nil {
					log.WithError(err).Error("Failed to ping server")
				}
			}
		}
	}()
}
