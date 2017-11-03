package database

import (
	"math/rand"

	"fmt"
	"kingshard/backend"
	"kingshard/config"
	"kingshard/core/errors"
	"kingshard/core/golog"
)

type Database struct {
	Cfg   *config.DatabaseConfig
	nodes []*backend.Node
}

func New(cfg *config.DatabaseConfig, nodes []*backend.Node) *Database {
	db := new(Database)
	db.Cfg = cfg
	db.nodes = nodes
	return db
}

func (db *Database) GetUser() (string, string) {
	return db.Cfg.User, db.Cfg.Password
}

func (db *Database) GetNode() (*backend.Node, error) {
	l := len(db.nodes)
	if l == 0 {
		golog.Error("database", "GetNode", fmt.Sprintf("node not exist in db %s", db), 0)
		return nil, errors.ErrNoDBNode
	}

	if l == 1 {
		return db.nodes[0], nil
	}

	start := rand.Intn(l)
	i := start
	var node *backend.Node
	for {
		node = db.nodes[i]
		i = (i + 1) % l
		if !node.Master.IsDown() || i == start {
			break
		}
		golog.Error("database", "GetNode", fmt.Sprintf("node %s is Down", node), 0)
	}

	return node, nil
}
