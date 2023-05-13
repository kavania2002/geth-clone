package bbolt

import (
	"errors"

	"github.com/ethereum/go-ethereum/ethdb"
	bolt "go.etcd.io/bbolt"
)

type Database struct {
	fn string
	db *bolt.DB
	tx *bolt.Tx
	b  *bolt.Bucket
}

func (db *Database) Has(key []byte) (bool, error) {
	value := db.b.Get(key)
	if value != nil {
		return true, nil
	}
	return false, errors.New("Has Error")
}

func (db *Database) Get(key []byte) ([]byte, error) {
	value := db.b.Get(key)
	if value != nil {
		return value, nil
	}
	return nil, errors.New("Get Error")
}

func (db *Database) Put(key []byte, value []byte) error {
	err := db.b.Put(key, value)
	if err != nil {
		return err
	}

	err = db.tx.Commit()
	if err != nil {
		return err
	}
	return nil
}

func (db *Database) Delete(key []byte) error {
	err := db.b.Delete(key)
	if err != nil {
		return err
	}

	err = db.tx.Commit()
	if err != nil {
		return err
	}
	return nil
}

func (db *Database) Close() error {
	err := db.db.Close()
	if err != nil {
		return err
	}
	return nil
}

func (db *Database) Stat(property string) (string, error) {
	return "", nil
}

func (db *Database) Compact(start []byte, limit []byte) error {
	return nil
}

func (db *Database) NewBatch() ethdb.Batch {
	return nil
}

func (db *Database) NewBatchWithSize(size int) ethdb.Batch {
	return nil
}

func (db *Database) NewIterator(prefix []byte, start []byte) ethdb.Iterator {
	return nil
}
func (db *Database) NewSnapshot() (ethdb.Snapshot, error) {
	return nil, nil
}

func New(file string) (*Database, error) {
	db, err := bolt.Open(file, 0666, nil)
	if err != nil {
		return nil, err
	}

	tx, err := db.Begin(true)
	if err != nil {
		return nil, nil
	}

	b, err := tx.CreateBucket([]byte("blockchain"))
	if err != nil {
		return nil, nil
	}

	return &Database{fn: file, db: db, tx: tx, b: b}, nil
} 