package mdbx

import (
	"bytes"
	"fmt"

	// "sync"

	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/torquem-ch/mdbx-go/mdbx"
	// "github.com/ethereum/go-ethereum/metrics"
)

type Database struct {
	fn string // filename for reporting
	// dbi mdbx.DBI
	// tx *mdbx.Txn
	env *mdbx.Env

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
	// quitChan chan chan error // Quit channel to stop the metrics collection before closing the database

	// log log.Logger // Contextual logger tracking the database path
}

type Batch struct {
	db *Database

	dataPut    map[string][]byte
	dataDelete map[string]bool
}

func (b *Batch) Put(key []byte, value []byte) error {
	if b.ValueSize() > 100 {
		b.Write()
		b.Reset()
	}
	// fmt.Println("Item to PUT - ", key, value)
	b.dataPut[string(key)] = value
	return nil
}

func (b *Batch) Delete(key []byte) error {
	// fmt.Println('')
	if b.ValueSize() > 100 {
		b.Write()
		b.Reset()
	}
	b.dataDelete[string(key)] = true
	return nil
}

func (b *Batch) Write() error {
	// fmt.Println("ITEM TO WRITE - ")
	// fmt.Println("DB - ", b.db)
	for key, value := range b.dataPut {

		err := b.db.Put([]byte(key), value)
		// fmt.Println("DB PUT - ")
		if err != nil {
			return err
		}
	}

	for key, _ := range b.dataDelete {
		// fmt.Println("DB DELETE - ")
		err := b.db.Delete([]byte(key))
		if err != nil {
			return err
		}
	}

	// fmt.Println("WRITE FINISHED")

	return nil
}

func (b *Batch) Replay(ethdb.KeyValueWriter) error {
	return nil
}

func (b *Batch) Reset() {
	for key := range b.dataPut {
		delete(b.dataPut, key)
	}
	for key := range b.dataDelete {
		delete(b.dataDelete, key)
	}
}

func (b *Batch) ValueSize() int {
	return len(b.dataDelete) + len(b.dataPut)
}

// ////////////////// SNAPSHOT
type snapshot struct {
	db  *Database
	txn *mdbx.Txn
}

func (s *snapshot) Has(key []byte) (bool, error) {
	// fmt.Println("SNAPSHOT GET")

	// err := s.db.env.View(func(txn *mdbx.Txn) error {
	dbi, err := s.txn.OpenRoot(0)
	if err != nil {
		return false, err
	}

	_, err = s.txn.Get(dbi, key)
	if err != nil {
		return false, nil
	}

	// fmt.Println("HAS FINISHED")

	return true, nil
	// })

	// if err != nil {
	// 	return false, err
	// }

	// // fmt.Println("HAS FINISHED")

	// return true, nil
}

func (s *snapshot) Get(key []byte) ([]byte, error) {
	// fmt.Println("SNAPSHOT GET")

	var value []byte

	dbi, err := s.txn.OpenRoot(0)
	if err != nil {
		return nil, err
	}

	value, err = s.txn.Get(dbi, key)
	if err != nil {
		return nil, err
	}

	// fmt.Println("GET FINISHED")
	return value, err

}

func (s *snapshot) Release() {
	s.txn.Commit()
}

// ////////////////////// Iterator
type iterator struct {
	db     *Database
	cur    *mdbx.Cursor
	prefix []byte
	start  []byte
	length int
}

func (it *iterator) Next() bool {
	key, _, err := it.cur.Get(nil, nil, mdbx.Next)
	if len(key) < it.length {
		return false
	}
	if !bytes.Equal(key[:it.length], it.prefix) || err != nil {
		return false
	}
	return true
}

func (it *iterator) Error() error {
	_, _, err := it.cur.Get(nil, nil, mdbx.GetCurrent)
	return err
}

func (it *iterator) Key() []byte {
	key, _, err := it.cur.Get(nil, nil, mdbx.GetCurrent)
	if err != nil {
		return nil
	}
	return key
}

func (it *iterator) Value() []byte {
	_, value, err := it.cur.Get(nil, nil, mdbx.GetCurrent)
	if err != nil {
		return nil
	}
	return value
}

func (it *iterator) Release() {
	it.cur.Close()
}

// /////////////////// NORMAL

func (db *Database) Has(key []byte) (bool, error) {
	// fmt.Println("ITEM TO HAS ", key)

	// tx, err := db.env.BeginTxn(nil, 0)
	// if err != nil {
	// 	return false, err
	// }

	// // Open the database within the transaction
	// dbi, err := tx.OpenRoot(0)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// _, err = tx.Get(dbi, key)
	// if err != nil {
	// 	return false, err
	// }

	// _, commitError := tx.Commit()
	// if commitError != nil {
	// 	return false, commitError
	// }

	err := db.env.View(func(txn *mdbx.Txn) error {
		// fmt.Println("HAS INITIATED")
		dbi, err := txn.OpenRoot(0)
		if err != nil {
			return err
		}

		// fmt.Println("HAS VALUE HERE")
		_, err = txn.Get(dbi, key)
		if err != nil {
			return err
		}
		// fmt.Println("HAS YOHO")

		return nil
	})

	if err != nil {
		return false, nil
	}

	// fmt.Println("HAS FINISHED")

	return true, nil
}

func (db *Database) Get(key []byte) ([]byte, error) {
	// fmt.Println("ITEM TO GET ", key)

	// tx, err := db.env.BeginTxn(nil, 0)
	// if err != nil {
	// 	return nil, err
	// }

	// // Open the database within the transaction
	// dbi, err := tx.OpenRoot(0)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// value, err := tx.Get(dbi, key)
	// if err != nil {
	// 	return []byte(""), err
	// }

	// _, commitError := tx.Commit()
	// if commitError != nil {
	// 	return nil, commitError
	// }

	var value []byte

	err := db.env.View(func(txn *mdbx.Txn) error {
		// fmt.Println("GET INITIATED")
		dbi, err := txn.OpenRoot(0)
		if err != nil {
			return err
		}

		// fmt.Println("GET VALUE HERE")
		value, err = txn.Get(dbi, key)
		if err != nil {
			return err
		}
		// fmt.Println("YOHO")

		return nil
	})

	if err != nil {
		return []byte(""), err
	}

	// fmt.Println("GET FINISHED")

	return value, nil
}

func (db *Database) Put(key []byte, value []byte) error {
	// fmt.Println("ITEM TO PUT ", key)

	// tx, err := db.env.BeginTxn(nil, 0)
	// if err != nil {
	// 	return err
	// }

	// // Open the database within the transaction
	// dbi, err := tx.OpenRoot(0)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// err = tx.Put(dbi, key, value, 0)
	// if err != nil {
	// 	return err
	// }

	// _, commitError := tx.Commit()
	// if commitError != nil {
	// 	return commitError
	// }

	err := db.env.Update(func(txn *mdbx.Txn) error {
		dbi, err := txn.OpenRoot(0)
		if err != nil {
			return err
		}

		err = txn.Put(dbi, key, value, 0)
		if err != nil {
			return err
		}

		// _, commitError := txn.Commit()
		// if commitError != nil {
		// 	return commitError
		// }

		return nil
	})

	if err != nil {
		return err
	}

	// fmt.Println("PUT FINISHED")

	return nil
}

func (db *Database) Delete(key []byte) error {
	// fmt.Println("ITEM TO DELETE ", key)

	// tx, err := db.env.BeginTxn(nil, 0)
	// if err != nil {
	// 	return err
	// }

	// // Open the database within the transaction
	// dbi, err := tx.OpenRoot(0)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// value, err := db.Get(key)
	// if err != nil {
	// 	return err
	// }
	// err = tx.Del(dbi, key, value)
	// if err != nil {
	// 	return err
	// }

	// _, commitError := tx.Commit()
	// if commitError != nil {
	// 	return commitError
	// }

	err := db.env.Update(func(txn *mdbx.Txn) error {
		// fmt.Println("DELETE INITIATED")
		dbi, err := txn.OpenRoot(0)
		if err != nil {
			return err
		}

		// fmt.Println("VALUE HERE")

		value, err := txn.Get(dbi, key)
		if err != nil {
			return nil
		}
		// fmt.Println("GET GOT")

		err = txn.Del(dbi, key, value)
		if err != nil {
			return err
		}

		// _, commitError := txn.Commit()
		// if commitError != nil {
		// 	return commitEraror
		// }

		return nil
	})

	if err != nil {
		return err
	}

	// fmt.Println("DELETE FINISHED")

	return nil
}

// func (db *Database) Commit() error {
// 	_, err := db.tx.Commit()
// 	if err != nil {
// 		return err
// 	}

// 	return nil
// }

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
	return &Batch{
		db:         db,
		dataPut:    make(map[string][]byte),
		dataDelete: make(map[string]bool),
	}
}

func (db *Database) NewBatchWithSize(size int) ethdb.Batch {
	return &Batch{
		db:         db,
		dataPut:    make(map[string][]byte),
		dataDelete: make(map[string]bool),
	}
}

func (db *Database) NewIterator(prefix []byte, start []byte) ethdb.Iterator {
	it := &iterator{}
	err := db.env.View(func(txn *mdbx.Txn) error {

		dbi, err := txn.OpenRoot(0)
		if err != nil {
			// fmt.Println("DBI Error Occurred - ", err)
			return nil
		}

		cur := mdbx.CreateCursor()
		err = cur.Bind(txn, dbi)
		if err != nil {
			// fmt.Println("CUR Error Occurred - ", err)
			return nil
		}

		fullString := append(prefix, start...)

		_, _, err = cur.Get(fullString, nil, mdbx.SetRange)
		if err != nil {
			// fmt.Println("SET Error Occured - ", err)
			return nil
		}

		it = &iterator{db: db,
			cur:    cur,
			prefix: prefix,
			start:  start,
			length: len(prefix)}

		return err
	})

	if err != nil {
		// fmt.Println("VIEW Error ", err)
		return nil
	}
	return it
}
func (db *Database) NewSnapshot() (ethdb.Snapshot, error) {
	// destinationEnv, err := mdbx.NewEnv()
	// if err != nil {
	// 	// fmt.Println("Cannot Open Environment")
	// 	return nil, nil
	// }
	// destinationEnv.SetGeometry(-1, -1, 1024*1024*1024, -1, -1, -1)

	// // fmt.Println("Environment Created ", destinationEnv, err)

	// var file string = "/temp"
	// // fmt.Println("File - ", )

	// err = destinationEnv.Open(file, 0, 0664)
	// if err != nil {
	// 	// fmt.Println("Cannot Use Open function")
	// 	return nil, err
	// }

	// err = db.env.View(func(sourceTxn *mdbx.Txn) error {
	// 	sourceDB, err := sourceTxn.OpenRoot(0)
	// 	if err != nil {
	// 		return err
	// 	}

	// 	destinationTxn, err := destinationEnv.BeginTxn(nil, 0)
	// 	if err != nil {
	// 		return err
	// 	}

	// 	destinationDB, err := destinationTxn.OpenRoot(0)
	// 	if err != nil {
	// 		return err
	// 	}

	// 	// Cursor to iterate over the records in the source database
	// 	sourceCursor := mdbx.CreateCursor()

	// 	// Loop through each record in the source database and copy it to the destination database
	// 	for {
	// 		key, value, err := sourceCursor.Get(key, value)
	// 		if err != nil {
	// 			if err == mdbx.NotFound {
	// 				break // Reached the end of the source database
	// 			}
	// 			return err
	// 		}

	// 		err = destinationTxn.Put(destinationDB, key, value, 0)
	// 		if err != nil {
	// 			return err
	// 		}
	// 	}

	// 	return destinationTxn.Commit()
	// })

	tx, err := db.env.BeginTxn(nil, 0)
	if err != nil {
		return nil, err
	}

	return &snapshot{
		db:  db,
		txn: tx,
	}, nil
	// return nil, nil
}

// func Reset(env *mdbx.Env) (*Database, error) {
// 	txn, err := env.BeginTxn(nil, 0)
// 	if err != nil {
// 		log.Fatal(err)
// 		return nil, err
// 	}
// 	// defer txn.Abort()

// 	// Open the database within the transaction
// 	dbi, err := txn.OpenRoot(0)
// 	if err != nil {
// 		log.Fatal(err)
// 		return nil, err
// 	}

// 	db := &Database{tx: txn, dbi: dbi}
// 	// fmt.Println("Created!!")
// 	return db, nil
// }

func New(file string) (*Database, error) {
	env, err := mdbx.NewEnv()
	if err != nil {
		// fmt.Println("Cannot Open Environment")
		return nil, nil
	}
	env.SetGeometry(-1, -1, 1024*1024*1024, -1, -1, -1)

	fmt.Println("Environment Created ", env, err)
	fmt.Println("File - ", file)

	err = env.Open(file, 0, 0664)
	if err != nil {
		// fmt.Println("Cannot Use Open function")
		return nil, err
	}

	// txn, err := env.BeginTxn(nil, 0)
	// if err != nil {
	// 	log.Fatal(err)
	// 	return nil, err
	// }
	// // defer txn.Abort()

	// // Open the database within the transaction
	// dbi, err := txn.OpenRoot(0)
	// if err != nil {
	// 	log.Fatal(err)
	// 	return nil, err
	// }
	// fmt.Println("Created!!")

	// db := &Database{fn: file, dbi: dbi, tx: txn, env: env}
	db := &Database{fn: file, env: env}

	fmt.Println("db - ", db)
	return db, nil
}
