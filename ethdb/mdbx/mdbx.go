package mdbx

import (
	"fmt"
	"log"

	// "sync"

	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/torquem-ch/mdbx-go/mdbx"
	// "github.com/ethereum/go-ethereum/metrics"
)

type Database struct {
	fn  string // filename for reporting
	dbi mdbx.DBI
	env *mdbx.Env
	tx  *mdbx.Txn

	// compTimeMeter       metrics.Meter // Meter for measuring the total time spent in database compaction
	// compReadMeter       metrics.Meter // Meter for measuring the data read during compaction
	// compWriteMeter      metrics.Meter // Meter for measuring the data written during compaction
	// writeDelayNMeter    metrics.Meter // Meter for measuring the write delay number due to database compaction
	// writeDelayMeter     metrics.Meter // Meter for measuring the write delay duration due to database compaction
	// diskSizeGauge       metrics.Gauge // Gauge for tracking the size of all the levels in the database
	// diskReadMeter       metrics.Meter // Meter for measuring the effective amount of data read
	// diskWriteMeter      metrics.Meter // Meter for measuring the effective amount of data written
	// memCompGauge        metrics.Gauge // Gauge for tracking the number of memory compaction
	// level0CompGauge     metrics.Gauge // Gauge for tracking the number of table compaction in level0
	// nonlevel0CompGauge  metrics.Gauge // Gauge for tracking the number of table compaction in non0 level
	// seekCompGauge       metrics.Gauge // Gauge for tracking the number of table compaction caused by read opt
	// manualMemAllocGauge metrics.Gauge // Gauge to track the amount of memory that has been manually allocated (not a part of runtime/GC)

	// quitLock sync.Mutex      // Mutex protecting the quit channel access
	quitChan chan chan error // Quit channel to stop the metrics collection before closing the database

	// log log.Logger // Contextual logger tracking the database path
}

func (db *Database) Has(key []byte) (bool, error) {
	_, err := db.tx.Get(db.dbi, key)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (db *Database) Get(key []byte) ([]byte, error) {
	fmt.Println("Item to Get: ", key)
	value, err := db.tx.Get(db.dbi, key)
	if err != nil {
		return []byte(""), err
	}
	return value, nil
}

func (db *Database) Put(key []byte, value []byte) error {
	err := db.tx.Put(db.dbi, key, value, 0)
	if err != nil {
		return err
	}
	// commitError := db.Commit()
	// if commitError != nil {
	// return err
	// }
	return nil
}

func (db *Database) Delete(key []byte) error {
	fmt.Println("Item to delete ", key)
	value, err := db.Get(key)
	if err != nil {
		return err
	}
	err = db.tx.Del(db.dbi, key, value)
	if err != nil {
		return err
	}

	// commitError := db.Commit()
	// if commitError != nil {
	// return err
	// }
	return nil
}

func (db *Database) Commit() error {
	_, err := db.tx.Commit()
	if err != nil {
		return err
	}

	return nil
}
func (db *Database) Close() error {
	db.env.Close()
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

func Reset(env *mdbx.Env) (*Database, error) {
	txn, err := env.BeginTxn(nil, 0)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	// defer txn.Abort()

	// Open the database within the transaction
	dbi, err := txn.OpenRoot(0)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	db := &Database{tx: txn, dbi: dbi}
	fmt.Println("Created!!")
	return db, nil
}

func New(file string) (*Database, error) {
	env, err := mdbx.NewEnv()
	if err != nil {
		fmt.Println("Cannot Open Environment")
		return nil, nil
	}
	fmt.Println("Environment Created ", env, err)
	fmt.Println("File - ", file)

	err = env.Open(file, 0, 0664)
	if err != nil {
		fmt.Println("Cannot Use Open function")
		return nil, err
	}
	

	txn, err := env.BeginTxn(nil, 0)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	// defer txn.Abort()

	// Open the database within the transaction
	dbi, err := txn.OpenRoot(0)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	fmt.Println("Created!!")

	db := &Database{fn: file, dbi:dbi, tx: txn, env: env}
	fmt.Println("db - ", db)
	return db, nil
}
