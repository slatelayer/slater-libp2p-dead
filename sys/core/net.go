package core

import (
	"context"
	"sync"
	"time"

	ds "github.com/ipfs/go-datastore"
	logging "github.com/ipfs/go-log/v2"
	libp2p "github.com/libp2p/go-libp2p"
	connmgr "github.com/libp2p/go-libp2p-connmgr"
	"github.com/libp2p/go-libp2p-core/crypto"
	host "github.com/libp2p/go-libp2p-core/host"
	peer "github.com/libp2p/go-libp2p-core/peer"
	routing "github.com/libp2p/go-libp2p-core/routing"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	dual "github.com/libp2p/go-libp2p-kad-dht/dual"
	noise "github.com/libp2p/go-libp2p-noise"
	pstore "github.com/libp2p/go-libp2p-peerstore/pstoreds"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	record "github.com/libp2p/go-libp2p-record"
	"github.com/libp2p/go-libp2p/p2p/discovery/mdns"
	discoveryRouting "github.com/libp2p/go-libp2p/p2p/discovery/routing"
	quic "github.com/libp2p/go-libp2p/p2p/transport/quic"
	tcp "github.com/libp2p/go-libp2p/p2p/transport/tcp"
)

const DiscoveryServiceTag = "slater" // TODO version

type node struct {
	host     host.Host
	dht      *dual.DHT
	psub     *pubsub.PubSub
	ctx      context.Context
	channels map[string]channel
	output   chan message

	discoveryKey string
	signet       []byte
}

type channel struct {
	topic *pubsub.Topic
	sub   *pubsub.Subscription
}

func (n node) close() error {
	return n.host.Close()
}

func startNet(key crypto.PrivKey, store datastore) (*node, error) {
	background := context.Background()
	var ddht *dual.DHT
	connectionManager, err := connmgr.NewConnManager(
		100,                                  // Lowwater
		400,                                  // HighWater,
		connmgr.WithGracePeriod(time.Minute), // GracePeriod
	)

	if err != nil {
		return nil, err
	}

	logging.SetAllLoggers(logging.LevelWarn)
	logging.SetLogLevel("slater:core", "debug")
	logging.SetLogLevel("mdns", "debug")

	bootstrapNodes, _ := defaultBootstrapPeers()

	pStore, err := pstore.NewPeerstore(background, store.store, pstore.DefaultOpts())
	if err != nil {
		log.Panic(err)
	}

	host, err := libp2p.New(
		libp2p.UserAgent("slater"), // implicit v0; TODO add version
		libp2p.Peerstore(pStore),
		libp2p.Identity(key),
		libp2p.ListenAddrStrings(
			"/ip4/0.0.0.0/tcp/0",
			"/ip4/0.0.0.0/udp/0/quic",
		),
		libp2p.Security(noise.ID, noise.New),
		libp2p.ChainOptions(
			libp2p.NoTransports,
			libp2p.Transport(tcp.NewTCPTransport),
			libp2p.Transport(quic.NewTransport),
		),
		libp2p.Routing(func(host host.Host) (routing.PeerRouting, error) {
			ddht, err = newDHT(background, host, store.store, bootstrapNodes)
			return ddht, err
		}),
		libp2p.ConnectionManager(connectionManager),
		libp2p.NATPortMap(),
		libp2p.EnableAutoRelay(),
		libp2p.EnableNATService(),
		libp2p.EnableHolePunching(),
	)
	if err != nil {
		return nil, err
	}

	psub, err := pubsub.NewGossipSub(
		background,
		host,
		pubsub.WithMessageSigning(true),
		pubsub.WithStrictSignatureVerification(true),
		pubsub.WithDiscovery(discoveryRouting.NewRoutingDiscovery(ddht)),
	)
	if err != nil {
		return nil, err
	}

	n := node{
		host:     host,
		dht:      ddht,
		psub:     psub,
		ctx:      background,
		channels: make(map[string]channel),
		output:   make(chan message),
	}

	n.bootstrap(bootstrapNodes)

	if err := runMdns(&n); err != nil {
		return nil, err
	}

	return &n, nil
}

func newDHT(ctx context.Context, host host.Host, store ds.Batching, bootstrapNodes []peer.AddrInfo) (*dual.DHT, error) {
	dhtOpts := []dual.Option{
		dual.DHTOption(dht.Datastore(store)),
		dual.DHTOption(dht.NamespacedValidator("pk", record.PublicKeyValidator{})),
		dual.DHTOption(dht.Concurrency(10)),
		dual.DHTOption(dht.BootstrapPeers(bootstrapNodes...)),
	}

	return dual.New(ctx, host, dhtOpts...)
}

func (n *node) bootstrap(peers []peer.AddrInfo) {
	connected := make(chan struct{})

	var wg sync.WaitGroup
	for _, pi := range peers {
		//host.Peerstore().AddAddrs(pi.ID, pi.Addrs, peerstore.PermanentAddrTTL)
		wg.Add(1)
		go func(pi peer.AddrInfo) {
			defer wg.Done()
			err := n.host.Connect(n.ctx, pi)
			if err != nil {
				log.Warnf("error connecting to peer %s: %s", pi.ID.Pretty(), err)
				return
			}
			log.Infof("Connected to %s", pi.ID.Pretty())
			connected <- struct{}{}
		}(pi)
	}

	go func() {
		wg.Wait()
		close(connected)
	}()

	i := 0
	for range connected {
		i++
	}
	if nPeers := len(peers); i < nPeers/2 {
		log.Warnf("only connected to %d bootstrap peers out of %d", i, nPeers)
	}

	err := n.dht.Bootstrap(n.ctx)
	if err != nil {
		log.Error(err)
		return
	}
}

type discoveryNotifee struct {
	h host.Host
}

func (n *discoveryNotifee) HandlePeerFound(pi peer.AddrInfo) {
	if pi.ID.String() == n.h.ID().String() {
		return
	}
	log.Infof("discovered new peer %s\n", pi.ID.Pretty())
	err := n.h.Connect(context.Background(), pi)
	if err != nil {
		log.Warnf("error connecting to peer %s: %s", pi.ID.Pretty(), err)
		return
	}
	log.Infof("Connected to %s", pi.ID.Pretty())
}

func runMdns(n *node) error {
	s := mdns.NewMdnsService(n.host, DiscoveryServiceTag, &discoveryNotifee{n.host})
	return s.Start()
}

func (n *node) join(k string, f pubsub.ValidatorEx) {
	n.psub.RegisterTopicValidator(k, f)

	topic, err := n.psub.Join(k)
	if err != nil {
		log.Error(err)
	}

	sub, err := topic.Subscribe()
	if err != nil {
		log.Error(err)
	}

	n.channels[k] = channel{topic, sub}

	go run(n, sub)
}

func (n *node) send(topic string, msg message) {
	bytes, err := encode(msg)
	if err != nil {
		log.Error(err)
		return
	}

	n.channels[topic].topic.Publish(n.ctx, bytes)
}

func run(n *node, sub *pubsub.Subscription) {
	for {
		msg, err := sub.Next(n.ctx)

		if err != nil {
			log.Error("NET shutting down, cuz", err)
			delete(n.channels, sub.Topic())
			return
		}

		if msg.ReceivedFrom == n.host.ID() {
			continue
		}

		m, err := decode(msg.Data)
		if err != nil {
			log.Error("NET", err)
			continue
		}

		m["device"] = msg.ReceivedFrom.Pretty()

		n.output <- m
	}
}
