package bambu

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

const (
	mqttPort       = 8883
	mqttUsername    = "bblp"
	reconnectDelay = 5 * time.Second
	pushAllPayload = `{"pushing": {"sequence_id": "0", "command": "pushall"}}`
)

// MessageHandler is called when a new print report is received from a printer.
type MessageHandler func(serial string, report *PrintReport)

// ConnectionHandler is called when the connection status changes.
type ConnectionHandler func(serial string, connected bool)

// ClientConfig holds the connection parameters for a single Bambu printer.
type ClientConfig struct {
	Name       string
	IP         string
	Serial     string
	AccessCode string
}

// Client manages the MQTT connection to a single Bambu Lab printer.
type Client struct {
	cfg       ClientConfig
	handler   MessageHandler
	onConnect ConnectionHandler
	mqtt      mqtt.Client
	mu        sync.Mutex
	cancel    context.CancelFunc
}

// NewClient creates a new Bambu MQTT client.
func NewClient(cfg ClientConfig, handler MessageHandler, onConnect ConnectionHandler) *Client {
	return &Client{
		cfg:       cfg,
		handler:   handler,
		onConnect: onConnect,
	}
}

// Connect starts the MQTT connection and blocks until the context is cancelled.
// The connection is non-blocking — if the broker is unreachable, it retries
// in the background. The app stays running regardless.
func (c *Client) Connect(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	c.mu.Lock()
	c.cancel = cancel
	c.mu.Unlock()

	topic := fmt.Sprintf("device/%s/report", c.cfg.Serial)

	opts := mqtt.NewClientOptions().
		AddBroker(fmt.Sprintf("tls://%s:%d", c.cfg.IP, mqttPort)).
		SetClientID(fmt.Sprintf("spoolman-bridge-%s", c.cfg.Serial)).
		SetUsername(mqttUsername).
		SetPassword(c.cfg.AccessCode).
		SetTLSConfig(&tls.Config{InsecureSkipVerify: true}).
		SetAutoReconnect(true).
		SetConnectTimeout(10 * time.Second).
		SetConnectionLostHandler(func(_ mqtt.Client, err error) {
			log.Printf("[%s] Connection lost: %v", c.cfg.Name, err)
			if c.onConnect != nil {
				c.onConnect(c.cfg.Serial, false)
			}
		}).
		SetReconnectingHandler(func(_ mqtt.Client, _ *mqtt.ClientOptions) {
			log.Printf("[%s] Reconnecting to %s...", c.cfg.Name, c.cfg.IP)
		}).
		SetOnConnectHandler(func(client mqtt.Client) {
			log.Printf("[%s] Connected to %s", c.cfg.Name, c.cfg.IP)
			if c.onConnect != nil {
				c.onConnect(c.cfg.Serial, true)
			}
			// Subscribe on every (re)connect
			token := client.Subscribe(topic, 0, c.onMessage)
			token.Wait()
			if token.Error() != nil {
				log.Printf("[%s] Subscribe error: %v", c.cfg.Name, token.Error())
				return
			}
			log.Printf("[%s] Subscribed to %s", c.cfg.Name, topic)
			// Request full status push
			c.requestPushAll(client)
		})

	c.mqtt = mqtt.NewClient(opts)

	// Manual connect loop — paho's ConnectRetry is silent on errors,
	// so we retry ourselves and log the actual failure reason.
	log.Printf("[%s] Connecting to %s:%d...", c.cfg.Name, c.cfg.IP, mqttPort)
	for {
		token := c.mqtt.Connect()
		token.Wait()
		if token.Error() == nil {
			break
		}
		log.Printf("[%s] Connection failed: %v (retrying in %s)", c.cfg.Name, token.Error(), reconnectDelay)
		select {
		case <-ctx.Done():
			log.Printf("[%s] Connect cancelled", c.cfg.Name)
			return
		case <-time.After(reconnectDelay):
		}
	}

	// Block until context is cancelled — auto-reconnect handles drops from here
	<-ctx.Done()

	// Disconnect in a goroutine with a timeout — paho can hang during connect retry
	done := make(chan struct{})
	go func() {
		c.mqtt.Disconnect(250)
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		log.Printf("[%s] Disconnect timed out, forcing close", c.cfg.Name)
	}
	log.Printf("[%s] Disconnected", c.cfg.Name)
}

// Close stops the client.
func (c *Client) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.cancel != nil {
		c.cancel()
	}
}

func (c *Client) onMessage(_ mqtt.Client, msg mqtt.Message) {
	var bambuMsg Message
	if err := json.Unmarshal(msg.Payload(), &bambuMsg); err != nil {
		log.Printf("[%s] Failed to parse message: %v", c.cfg.Name, err)
		return
	}

	if bambuMsg.Print == nil {
		return
	}

	c.handler(c.cfg.Serial, bambuMsg.Print)
}

func (c *Client) requestPushAll(client mqtt.Client) {
	topic := fmt.Sprintf("device/%s/request", c.cfg.Serial)
	token := client.Publish(topic, 0, false, pushAllPayload)
	token.Wait()
	if token.Error() != nil {
		log.Printf("[%s] Failed to request pushall: %v", c.cfg.Name, token.Error())
	}
}
