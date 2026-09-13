package storage

import (
	"crypto/rand"
	"database/sql"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

type Database struct {
	db *sql.DB
}

func newUUID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

func Open(path string) (*Database, error) {
	connStr := path
	if path != ":memory:" {
		connStr = path + "?_pragma=busy_timeout(5000)&_pragma=foreign_keys(ON)&_pragma=journal_mode(WAL)"
	}
	db, err := sql.Open("sqlite", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &Database{db: db}, nil
}

func (d *Database) Initialize() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS peers (
			id TEXT PRIMARY KEY,
			name TEXT UNIQUE NOT NULL,
			public_key TEXT NOT NULL,
			private_key_path TEXT NOT NULL,
			preshared_key_path TEXT DEFAULT '',
			vpn_ipv4 TEXT UNIQUE NOT NULL,
			dns_servers TEXT DEFAULT '',
			allowed_ips TEXT NOT NULL DEFAULT '0.0.0.0/0',
			keepalive INTEGER DEFAULT 25,
			status TEXT NOT NULL DEFAULT 'active',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);`,
		`CREATE TABLE IF NOT EXISTS server_state (
			key TEXT PRIMARY KEY,
			value TEXT NOT NULL,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);`,
	}

	for _, q := range queries {
		if _, err := d.db.Exec(q); err != nil {
			return fmt.Errorf("failed to execute query %q: %w", q, err)
		}
	}
	return nil
}

func (d *Database) Close() error {
	return d.db.Close()
}

func (d *Database) CreatePeer(peer *Peer) error {
	if peer.ID == "" {
		peer.ID = newUUID()
	}
	now := time.Now().UTC()
	if peer.CreatedAt.IsZero() {
		peer.CreatedAt = now
	}
	if peer.UpdatedAt.IsZero() {
		peer.UpdatedAt = now
	}

	query := `INSERT INTO peers (id, name, public_key, private_key_path, preshared_key_path, vpn_ipv4, dns_servers, allowed_ips, keepalive, status, created_at, updated_at) 
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := d.db.Exec(query, peer.ID, peer.Name, peer.PublicKey, peer.PrivateKeyPath, peer.PresharedKeyPath, peer.VPNIPv4, peer.DNSServers, peer.AllowedIPs, peer.Keepalive, peer.Status, peer.CreatedAt.Format(time.RFC3339), peer.UpdatedAt.Format(time.RFC3339))
	return err
}

func scanPeer(scanner interface {
	Scan(dest ...any) error
}) (*Peer, error) {
	var p Peer
	var cTime, uTime string
	err := scanner.Scan(&p.ID, &p.Name, &p.PublicKey, &p.PrivateKeyPath, &p.PresharedKeyPath, &p.VPNIPv4, &p.DNSServers, &p.AllowedIPs, &p.Keepalive, &p.Status, &cTime, &uTime)
	if err != nil {
		return nil, err
	}
	p.CreatedAt, _ = time.Parse(time.RFC3339, cTime)
	p.UpdatedAt, _ = time.Parse(time.RFC3339, uTime)
	return &p, nil
}

func (d *Database) GetPeer(name string) (*Peer, error) {
	row := d.db.QueryRow(`SELECT id, name, public_key, private_key_path, preshared_key_path, vpn_ipv4, dns_servers, allowed_ips, keepalive, status, created_at, updated_at FROM peers WHERE name = ?`, name)
	return scanPeer(row)
}

func (d *Database) GetPeerByIP(ip string) (*Peer, error) {
	row := d.db.QueryRow(`SELECT id, name, public_key, private_key_path, preshared_key_path, vpn_ipv4, dns_servers, allowed_ips, keepalive, status, created_at, updated_at FROM peers WHERE vpn_ipv4 = ?`, ip)
	return scanPeer(row)
}

func (d *Database) ListPeers() ([]*Peer, error) {
	rows, err := d.db.Query(`SELECT id, name, public_key, private_key_path, preshared_key_path, vpn_ipv4, dns_servers, allowed_ips, keepalive, status, created_at, updated_at FROM peers`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var peers []*Peer
	for rows.Next() {
		p, err := scanPeer(rows)
		if err != nil {
			return nil, err
		}
		peers = append(peers, p)
	}
	return peers, rows.Err()
}

func (d *Database) ListActivePeers() ([]*Peer, error) {
	rows, err := d.db.Query(`SELECT id, name, public_key, private_key_path, preshared_key_path, vpn_ipv4, dns_servers, allowed_ips, keepalive, status, created_at, updated_at FROM peers WHERE status = 'active'`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var peers []*Peer
	for rows.Next() {
		p, err := scanPeer(rows)
		if err != nil {
			return nil, err
		}
		peers = append(peers, p)
	}
	return peers, rows.Err()
}

func (d *Database) UpdatePeer(peer *Peer) error {
	peer.UpdatedAt = time.Now().UTC()
	query := `UPDATE peers SET public_key = ?, private_key_path = ?, preshared_key_path = ?, vpn_ipv4 = ?, dns_servers = ?, allowed_ips = ?, keepalive = ?, status = ?, updated_at = ? WHERE name = ?`
	_, err := d.db.Exec(query, peer.PublicKey, peer.PrivateKeyPath, peer.PresharedKeyPath, peer.VPNIPv4, peer.DNSServers, peer.AllowedIPs, peer.Keepalive, peer.Status, peer.UpdatedAt.Format(time.RFC3339), peer.Name)
	return err
}

func (d *Database) UpdatePeerStatus(name, status string) error {
	query := `UPDATE peers SET status = ?, updated_at = ? WHERE name = ?`
	_, err := d.db.Exec(query, status, time.Now().UTC().Format(time.RFC3339), name)
	return err
}

func (d *Database) DeletePeer(name string) error {
	_, err := d.db.Exec(`DELETE FROM peers WHERE name = ?`, name)
	return err
}

func (d *Database) PeerExists(name string) (bool, error) {
	var count int
	err := d.db.QueryRow(`SELECT COUNT(*) FROM peers WHERE name = ?`, name).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (d *Database) SetState(key, value string) error {
	query := `INSERT INTO server_state (key, value, updated_at) VALUES (?, ?, ?)
	ON CONFLICT(key) DO UPDATE SET value = excluded.value, updated_at = excluded.updated_at`
	_, err := d.db.Exec(query, key, value, time.Now().UTC().Format(time.RFC3339))
	return err
}

func (d *Database) GetState(key string) (string, error) {
	var value string
	err := d.db.QueryRow(`SELECT value FROM server_state WHERE key = ?`, key).Scan(&value)
	if err == sql.ErrNoRows {
		return "", fmt.Errorf("state not found")
	}
	return value, err
}

func (d *Database) DeleteState(key string) error {
	_, err := d.db.Exec(`DELETE FROM server_state WHERE key = ?`, key)
	return err
}
