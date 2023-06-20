package wireguard

import (
	"crypto/sha1"
	"fmt"
	"net"

	"github.com/gravitl/netclient/cache"
	"github.com/gravitl/netclient/config"
	"github.com/gravitl/netclient/ncutils"
	proxy_config "github.com/gravitl/netclient/nmproxy/config"
	"github.com/gravitl/netclient/nmproxy/peer"
	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

// SetPeers - sets peers on netmaker WireGuard interface
func SetPeers(replace bool) error {

	peers := config.GetHostPeerList()
	for i := range peers {
		peer := peers[i]
		if !peer.Remove && checkForBetterEndpoint(&peer) {
			peers[i] = peer
		}
	}
	GetInterface().Config.Peers = peers
	peers = peer.SetPeersEndpointToProxy(peers)
	config := wgtypes.Config{
		ReplacePeers: replace,
		Peers:        peers,
	}
	return apply(&config)
}

// RemovePeers - removes the peers in the list given from the interface
func RemovePeers(peers []wgtypes.PeerConfig) error {
	for i := range peers {
		proxy_config.GetCfg().RemovePeer(peers[i].PublicKey.String())
		peers[i].Remove = true
	}

	config := wgtypes.Config{
		Peers: peers,
	}
	return apply(&config)
}

// == private ==

// UpdatePeer replaces a wireguard peer
// temporarily making public func to pass staticchecks
// this function will be required in future when update node on server is refactored
func UpdatePeer(p *wgtypes.PeerConfig) error {
	config := wgtypes.Config{
		Peers:        []wgtypes.PeerConfig{*p},
		ReplacePeers: false,
	}
	return apply(&config)
}

func apply(c *wgtypes.Config) error {
	wg, err := wgctrl.New()
	if err != nil {
		return err
	}
	defer wg.Close()

	return wg.ConfigureDevice(ncutils.GetInterfaceName(), *c)
}

// returns if better endpoint has been calculated for this peer already
// if so sets it and returns true
func checkForBetterEndpoint(peer *wgtypes.PeerConfig) bool {
	if peer.Endpoint == nil {
		return false
	}
	if endpoint, ok := cache.EndpointCache.Load(fmt.Sprintf("%v", sha1.Sum([]byte(peer.PublicKey.String())))); ok && endpoint != nil {
		var cacheEndpoint cache.EndpointCacheValue
		cacheEndpoint, ok = endpoint.(cache.EndpointCacheValue)
		if ok {

			peer.Endpoint.IP = net.ParseIP(cacheEndpoint.Endpoint.String())
		}
		return ok
	}
	return false
}
