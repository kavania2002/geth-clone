package mdbx

import (
	"bytes"
	"fmt"

	// "sync"

	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/torquem-ch/mdbx-go/mdbx"
)

type Database struct {
	fn string // filename for reporting
	env *mdbx.Env // Environment to open the database
}

type Batch struct {
	db *Database // the database instance
	dataPut    map[string][]byte // map to store the data to be put
	dataDelete map[string]bool // map to store the data to be deleted
}

// Put inserts the given value into the batch for later committing.
func (b *Batch) Put(key []byte, value []byte) error {
	if b.ValueSize() > 100 {
		b.Write()
		b.Reset()
	}
	// fmt.Println("Item to PUT - ", key, value)
	b.dataPut[string(key)] = value
	return nil
}

// Delete inserts the a key removal into the batch for later committing.
func (b *Batch) Delete(key []byte) error {
	// fmt.Println('')
	if b.ValueSize() > 100 {
		b.Write()
		b.Reset()
	}
	b.dataDelete[string(key)] = true
	return nil
}

// Write flushes any accumulated data to disk.
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

// Replay replays the batch contents.
func (b *Batch) Replay(ethdb.KeyValueWriter) error {
	return nil
}

// Reset resets the batch for reuse.
func (b *Batch) Reset() {
	for key := range b.dataPut {
		delete(b.dataPut, key)
	}
	for key := range b.dataDelete {
		delete(b.dataDelete, key)
	}
}

// ValueSize retrieves the amount of data queued up for writing.
func (b *Batch) ValueSize() int {
	return len(b.dataDelete) + len(b.dataPut)
}

// ////////////////// SNAPSHOT
// snapshot wraps a mdbx snapshot for implementing the Snapshot interface.
type snapshot struct {
	db  *Database // the database instance
	txn *mdbx.Txn // the transaction to create a snapshot
}

// Has retrieves if a key is present in the snapshot backing by a key-value
// data store.
func (s *snapshot) Has(key []byte) (bool, error) {
	// fmt.Println("SNAPSHOT GET")

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

}


// Get retrieves the given key if it's present in the snapshot backing by
// key-value data store.
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

// Release releases associated resources. Release should always succeed and can
// be called multiple times without causing error.
func (s *snapshot) Release() {
	s.txn.Commit()
}

// ////////////////////// Iterator
type iterator struct {
	db     *Database // the database instance
	cur    *mdbx.Cursor // Cursor to hold a position
	prefix []byte // key prefix
	start  []byte // start key
	length int // length of the prefix
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


		return nil
	})

	if err != nil {
		return err
	}

	// fmt.Println("DELETE FINISHED")

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

// NewBatch creates a write-only key-value store that buffers changes to its host
// database until a final write is called.
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

// NewIterator creates a binary-alphabetical iterator over a subset
// of database content with a particular key prefix, starting at a particular
// initial key (or after, if it does not exist).
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

// NewSnapshot creates a database snapshot based on the current state.
// The created snapshot will not be affected by all following mutations
// happened on the database.
// Note don't forget to release the snapshot once it's used up, otherwise
// the stale data will never be cleaned up by the underlying compactor.
func (db *Database) NewSnapshot() (ethdb.Snapshot, error) {

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

	
	db := &Database{fn: file, env: env}

	fmt.Println("db - ", db)
	return db, nil
}