package doctor

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestDoctorChecks(t *testing.T) {
	cfg := &Config{
		Interface: "wg0",
		ConfigDir: "/tmp",
		Subnet:    "10.0.0.0/24",
		DBPath:    "/tmp/db.sqlite",
	}

	report := RunDiagnostics(cfg)
	assert.NotNil(t, report)
	assert.Greater(t, len(report.Checks), 0)
}
