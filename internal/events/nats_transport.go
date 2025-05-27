package events

import (
	"fmt"
	"log"
	"time"

	"github.com/nats-io/nats.go"
)

// NATSTransport implements the EventTransport interface using NATS
type NATSTransport struct {
	conn      *nats.Conn
	url       string
	subs      []*nats.Subscription
	connected bool
	options   []nats.Option
}

// NATSConfig represents configuration options for NATS transport
type NATSConfig struct {
	URL            string
	ConnectTimeout time.Duration
	MaxReconnects  int
	ReconnectWait  time.Duration
}

// DefaultNATSConfig provides sensible defaults for NATS
func DefaultNATSConfig() NATSConfig {
	return NATSConfig{
		URL:            nats.DefaultURL,
		ConnectTimeout: 5 * time.Second,
		MaxReconnects:  10,
		ReconnectWait:  1 * time.Second,
	}
}

// NewNATSTransport creates a new NATS transport
func NewNATSTransport(config NATSConfig) (*NATSTransport, error) {
	options := []nats.Option{
		nats.Timeout(config.ConnectTimeout),
		nats.MaxReconnects(config.MaxReconnects),
		nats.ReconnectWait(config.ReconnectWait),
		nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {
			log.Printf("NATS disconnected: %v", err)
		}),
		nats.ReconnectHandler(func(nc *nats.Conn) {
			log.Printf("NATS reconnected to %s", nc.ConnectedUrl())
		}),
		nats.ClosedHandler(func(nc *nats.Conn) {
			log.Printf("NATS connection closed")
		}),
	}

	conn, err := nats.Connect(config.URL, options...)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to NATS: %w", err)
	}

	return &NATSTransport{
		conn:      conn,
		url:       config.URL,
		subs:      make([]*nats.Subscription, 0),
		connected: true,
		options:   options,
	}, nil
}

// Publish sends data to NATS for a specific topic
func (n *NATSTransport) Publish(topic string, data []byte) error {
	if !n.connected {
		return fmt.Errorf("not connected to NATS")
	}
	return n.conn.Publish(topic, data)
}

// Subscribe registers a handler for a NATS topic
func (n *NATSTransport) Subscribe(topic string, handler func([]byte)) error {
	if !n.connected {
		return fmt.Errorf("not connected to NATS")
	}

	sub, err := n.conn.Subscribe(topic, func(msg *nats.Msg) {
		handler(msg.Data)
	})

	if err != nil {
		return fmt.Errorf("failed to subscribe to NATS topic %s: %w", topic, err)
	}

	n.subs = append(n.subs, sub)
	return nil
}

// Close cleans up NATS resources
func (n *NATSTransport) Close() error {
	if !n.connected {
		return nil
	}

	for _, sub := range n.subs {
		if err := sub.Unsubscribe(); err != nil {
			log.Printf("Error unsubscribing from NATS: %v", err)
		}
	}

	n.conn.Close()
	n.connected = false
	return nil
}
