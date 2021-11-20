package peernet

import (
	"context"
	"fmt"
	"github.com/iotexproject/go-p2p"
	"github.com/l3uddz/nabarr/logger"
	"github.com/multiformats/go-multiaddr"
	"github.com/rs/zerolog"
)

type Config struct {
	ExternalHost  string `yaml:"external_host"`
	ExternalPort  int    `yaml:"external_port"`
	IdentityKey   string `yaml:"identity_key"`
	NetworkKey    string `yaml:"network_key"`
	BootstrapNode string `yaml:"bootstrap_node"`

	Verbosity string `yaml:"verbosity,omitempty"`
}

type Node struct {
	cfg  Config
	ctx  context.Context
	log  zerolog.Logger
	host *p2p.Host
}

func New(c Config) (*Node, error) {
	// validate config
	if c.ExternalHost == "" {
		return nil, fmt.Errorf("external_host must be provided")
	}

	if c.NetworkKey == "" {
		return nil, fmt.Errorf("network_key must be provided")
	}

	// set config defaults
	if c.ExternalPort == 0 {
		c.ExternalPort = 9157
	}

	// init node
	opts := []p2p.Option{
		p2p.HostName("0.0.0.0"),
		p2p.Port(c.ExternalPort),
		p2p.ExternalHostName(c.ExternalHost),
		p2p.ExternalPort(c.ExternalPort),
		p2p.DHTProtocolID(1337),
		p2p.MasterKey(c.IdentityKey),
		p2p.PrivateNetworkPSK(c.NetworkKey),

		// feats
		p2p.Gossip(),
		p2p.SecureIO(),
	}

	ctx := context.Background()
	host, err := p2p.NewHost(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("new host: %w", err)
	}

	l := logger.New(c.Verbosity).With().
		Str("identity", host.HostIdentity()).
		Logger()

	n := &Node{
		cfg:  c,
		ctx:  ctx,
		host: host,

		log: l,
	}

	return n, nil
}

func (n *Node) Subscribe(topic string, callback p2p.HandleBroadcast) error {
	return n.host.AddBroadcastPubSub(n.ctx, topic, callback)
}

func (n *Node) Broadcast(topic string, data []byte) error {
	return n.host.Broadcast(n.ctx, topic, data)
}

func (n *Node) Bootstrap() error {
	if n.cfg.BootstrapNode == "" {
		return nil
	}

	n.log.Debug().
		Str("bootstrap_node", n.cfg.BootstrapNode).
		Msg("Connecting to peer network")

	ma, err := multiaddr.NewMultiaddr(n.cfg.BootstrapNode)
	if err != nil {
		return fmt.Errorf("parse bootstrap_node: %w", err)
	}

	if err := n.host.ConnectWithMultiaddr(n.ctx, ma); err != nil {
		return fmt.Errorf("connect bootstrap_node: %w", err)
	}

	n.host.JoinOverlay(n.ctx)

	n.log.Info().Msg("Connected to peer network")
	return nil
}

func (n *Node) Close() error {
	return n.host.Close()
}
