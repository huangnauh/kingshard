package database

import (
	"math/rand"
	"sync"
	"time"

	"kingshard/config"
	"kingshard/core/errors"
)

type Database struct {
	Cfg *config.DatabaseConfig

	sync.RWMutex
	LastNodeIndex int
}

func (db *Database) GenerateIndex() {
	n := len(db.Cfg.Nodes)
	if n <= 0 {
		return
	}

	rand.Seed(time.Now().UnixNano())
	db.LastNodeIndex = rand.Intn(n)
}

func (db *Database) ParseDatabase(cfg *config.DatabaseConfig) {
	db.Cfg = cfg
	db.GenerateIndex()
}

func (db *Database) GetNextNode() (string, error) {
	l := len(db.Cfg.Nodes)
	if l == 0 {
		return "", errors.ErrNoDBNode
	}

	if l == 1 {
		return db.Cfg.Nodes[0], nil
	}

	if l <= db.LastNodeIndex {
		db.GenerateIndex()
	}

	node := db.Cfg.Nodes[db.LastNodeIndex]
	db.LastNodeIndex = (db.LastNodeIndex + 1) % l
	return node, nil
}
