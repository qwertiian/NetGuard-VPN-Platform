package vpn

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestParseWgDump(t *testing.T) {
	dump := "privkey\tpubkey\t51820\toff\n" +
		"peerpubkey\tpsk\t192.168.1.1:1234\t10.0.0.2/32\t1630000000\t100\t200\t25"
	
	status, err := ParseWgDump(dump)
	assert.NoError(t, err)
	assert.Equal(t, "pubkey", status.PublicKey)
	assert.Equal(t, 51820, status.ListenPort)
	assert.Len(t, status.Peers, 1)
	
	peer := status.Peers[0]
	assert.Equal(t, "peerpubkey", peer.PublicKey)
	assert.Equal(t, "192.168.1.1:1234", peer.Endpoint)
	assert.Equal(t, "10.0.0.2/32", peer.AllowedIPs)
	assert.Equal(t, time.Unix(1630000000, 0), peer.LatestHandshake)
	assert.Equal(t, int64(100), peer.TransferRx)
	assert.Equal(t, int64(200), peer.TransferTx)
	assert.Equal(t, 25, peer.PersistentKeepalive)
}
