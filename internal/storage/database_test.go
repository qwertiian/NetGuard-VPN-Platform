package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDatabase(t *testing.T) {
	db, err := Open(":memory:")
	require.NoError(t, err)
	defer db.Close()

	err = db.Initialize()
	require.NoError(t, err)

	peer := &Peer{
		Name:           "testpeer",
		PublicKey:      "pubkey123",
		PrivateKeyPath: "/etc/ngvpn/testpeer.key",
		VPNIPv4:        "10.0.0.2",
		Status:         "active",
		AllowedIPs:     "0.0.0.0/0",
	}

	err = db.CreatePeer(peer)
	require.NoError(t, err)

	exists, err := db.PeerExists("testpeer")
	require.NoError(t, err)
	assert.True(t, exists)

	p, err := db.GetPeer("testpeer")
	require.NoError(t, err)
	assert.Equal(t, "10.0.0.2", p.VPNIPv4)
	assert.NotEmpty(t, p.ID)

	p2, err := db.GetPeerByIP("10.0.0.2")
	require.NoError(t, err)
	assert.Equal(t, "testpeer", p2.Name)

	peers, err := db.ListPeers()
	require.NoError(t, err)
	assert.Len(t, peers, 1)

	active, err := db.ListActivePeers()
	require.NoError(t, err)
	assert.Len(t, active, 1)

	err = db.UpdatePeerStatus("testpeer", "revoked")
	require.NoError(t, err)

	active2, err := db.ListActivePeers()
	require.NoError(t, err)
	assert.Len(t, active2, 0)

	err = db.DeletePeer("testpeer")
	require.NoError(t, err)

	exists, err = db.PeerExists("testpeer")
	require.NoError(t, err)
	assert.False(t, exists)

	err = db.SetState("testkey", "testvalue")
	require.NoError(t, err)

	val, err := db.GetState("testkey")
	require.NoError(t, err)
	assert.Equal(t, "testvalue", val)

	err = db.DeleteState("testkey")
	require.NoError(t, err)

	_, err = db.GetState("testkey")
	assert.Error(t, err)
}
