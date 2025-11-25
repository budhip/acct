package localstorage

import (
	"encoding/json"
	"fmt"
	"os"
	"path"

	"github.com/dgraph-io/badger/v4"
)

type BadgerStorage[T any] struct {
	db     *badger.DB
	bucket string
	pathDB string
}

var ErrKeyNotFound = badger.ErrKeyNotFound

func NewBadgerStorage[T any](bucket string) (*BadgerStorage[T], error) {
	pathDB := path.Join(os.TempDir(), bucket)

	// v2
	// opts := badger.DefaultOptions(pathDB).
	// 	WithBlockCacheSize(64 << 20).
	// 	WithIndexCacheSize(64 << 20).
	// 	WithMaxTableSize(64 << 20).
	// 	WithTableLoadingMode(options.FileIO).
	// 	WithValueLogLoadingMode(options.FileIO).
	// 	WithKeepL0InMemory(true)

	// v4
	// opts := badger.DefaultOptions(pathDB).
	// 	WithValueLogFileSize(1 << 30). // 1GB
	// 	WithBlockCacheSize(128 << 20).
	// 	WithIndexCacheSize(64 << 20)
	// opts.Logger = nil
	// opts.Compression = options.None

	opts := badger.DefaultOptions(pathDB).
		WithValueLogFileSize(1 << 30). // 1GB to reduce vLog GC frequency
		WithBlockCacheSize(512 << 20). // 512MB for read performance
		WithIndexCacheSize(256 << 20). // 256MB index cache for faster lookups
		// WithMaxTableSize(64 << 20).   // 64MB table size (default is 64MB)
		WithNumMemtables(5).             // More memtables reduce write stalls
		WithNumLevelZeroTables(10).      // Increase before L0 compaction kicks in
		WithNumLevelZeroTablesStall(20). // Allow more L0 tables before stalling writes
		// WithLevelOneSize(512 << 20).     // 512MB in L1 before compaction
		// WithMaxLevels(7).                // Default level structure
		WithLogger(nil)

	db, err := badger.Open(opts)
	if err != nil {
		return nil, err
	}

	return &BadgerStorage[T]{
		db:     db,
		bucket: bucket,
		pathDB: pathDB,
	}, nil
}

func (b BadgerStorage[T]) Set(key string, value T) error {
	rawVal, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	if err = b.db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(key), rawVal)
	}); err != nil {
		return fmt.Errorf("failed to set value to localstorage: %w", err)
	}

	return nil
}

func (b BadgerStorage[T]) Delete(key string) error {
	if err := b.db.Update(func(txn *badger.Txn) error {
		return txn.Delete([]byte(key))
	}); err != nil {
		return fmt.Errorf("failed to delete value from localstorage: %w", err)
	}

	return nil
}

func (b BadgerStorage[T]) Get(key string) (T, error) {
	var val T
	var rawVal []byte
	err := b.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			return err
		}

		rawVal, err = item.ValueCopy(nil)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		if err == badger.ErrKeyNotFound {
			return val, ErrKeyNotFound
		}
		return val, fmt.Errorf("failed to get value from localstorage: %w", err)
	}

	if rawVal != nil {
		if err = json.Unmarshal(rawVal, &val); err != nil {
			return val, fmt.Errorf("failed to unmarshal value from localstorage: %w", err)
		}
	}

	return val, nil
}

func (b BadgerStorage[T]) ForEach(f func(key string, value T) error) error {
	err := b.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			k := item.Key()
			v, err := item.ValueCopy(nil)
			if err != nil {
				return err
			}

			var val T
			err = json.Unmarshal(v, &val)
			if err != nil {
				return fmt.Errorf("failed to unmarshal value: %w", err)
			}

			if err = f(string(k), val); err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to iterate over localstorage: %w", err)
	}

	return nil
}

func (b BadgerStorage[T]) Clean() error {
	return os.RemoveAll(b.pathDB)
}

func (b BadgerStorage[T]) Close() error {
	// if err := b.db.Sync(); err != nil {
	// 	return fmt.Errorf("sync before close failed: %w", err)
	// }
	return b.db.Close()
}
