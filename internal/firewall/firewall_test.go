package firewall

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateInterface(t *testing.T) {
	m := NewIPTablesManager()

	assert.NoError(t, m.validateInterface("eth0"))
	assert.NoError(t, m.validateInterface("wg0"))
	assert.Error(t, m.validateInterface("eth0; rm -rf /"))
	assert.Error(t, m.validateInterface("wg0 && echo bad"))
}

func TestDetectDefaultInterfaceParsing(t *testing.T) {
	iface, err := DetectDefaultInterface()
	if err == nil {
		assert.NotEmpty(t, iface)
	}
}
