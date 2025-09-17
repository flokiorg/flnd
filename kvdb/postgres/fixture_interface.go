package postgres

import "github.com/flokiorg/walletd/walletdb"

type Fixture interface {
	DB() walletdb.DB
	Dump() (map[string]interface{}, error)
}
