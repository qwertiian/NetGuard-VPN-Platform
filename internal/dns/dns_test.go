package dns

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateServers(t *testing.T) {
	assert.NoError(t, ValidateServers([]string{"1.1.1.1", "8.8.8.8"}))
	assert.Error(t, ValidateServers([]string{"1.1.1.1", "invalid"}))
}

func TestFormatForWireGuard(t *testing.T) {
	assert.Equal(t, "1.1.1.1, 9.9.9.9", FormatForWireGuard([]string{"1.1.1.1", "9.9.9.9"}))
}

func TestDefaultServers(t *testing.T) {
	assert.Equal(t, []string{"1.1.1.1", "9.9.9.9"}, DefaultServers())
}

func TestGetServersForClient(t *testing.T) {
	c := &Config{
		Servers: []string{"8.8.8.8"},
		Mode:    "custom",
	}

	assert.Equal(t, []string{"8.8.8.8"}, c.GetServersForClient(nil))
	assert.Equal(t, []string{"4.4.4.4"}, c.GetServersForClient([]string{"4.4.4.4"}))

	cEmpty := &Config{}
	assert.Equal(t, []string{"1.1.1.1", "9.9.9.9"}, cEmpty.GetServersForClient(nil))
}
