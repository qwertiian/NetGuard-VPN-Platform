package firewall

type Manager interface {
	Setup(wgInterface, extInterface, subnet string) error
	Teardown(wgInterface, extInterface, subnet string) error
	IsConfigured(wgInterface string) (bool, error)
	EnsureSSHAccess() error
	BackupRules(path string) error
	RestoreRules(path string) error
}
