package pkgmirror

import (
	"errors"
	"fmt"
	"os"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/boltdb/bolt"
)

func OpenDatabaseWithBucket(basePath string, bucket []byte) (db *bolt.DB, err error) {
	if err = os.MkdirAll(basePath, 0755); err != nil {
		return
	}

	path := fmt.Sprintf("%s/%s.db", basePath, bucket)

	db, err = bolt.Open(path, 0600, &bolt.Options{
		Timeout:  1 * time.Second,
		ReadOnly: false,
	})

	if err != nil {
		return
	}

	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(bucket)

		return err
	})

	return
}

// adapted from : https://github.com/boltdb/bolt/blob/master/cmd/bolt/main.go

var (

	// ErrPathRequired is returned when the path to a Bolt database is not specified.
	ErrPathRequired = errors.New("path required")

	// ErrFileNotFound is returned when a Bolt database does not exist.
	ErrFileNotFound = errors.New("file not found")

	ErrUnableToCloseDatabase = errors.New("Unable to close the database")
)

type BoltCompacter struct {
	TxMaxSize int64
	Logger    *log.Logger
}

// Run executes the command.
func (bc *BoltCompacter) Compact(srcPath string) (err error) {
	now := time.Now()
	bckPath := fmt.Sprintf("%s.%d-%d-%d-%d.backup", srcPath, now.Year(), now.Month(), now.Day(), now.Unix())
	dstPath := fmt.Sprintf("%s.%d-%d-%d-%d.compacted", srcPath, now.Year(), now.Month(), now.Day(), now.Unix())

	logger := bc.Logger.WithFields(log.Fields{
		"method":  "compacter",
		"srcPath": srcPath,
		"dstPath": dstPath,
	})

	// Require database paths.
	if srcPath == "" {
		logger.Error("srcPath is not defined")

		return ErrPathRequired
	}

	// Ensure source file exists.
	fi, err := os.Stat(srcPath)
	if os.IsNotExist(err) {
		logger.Error("srcPath does not exist")

		return ErrFileNotFound
	} else if err != nil {
		return err
	}
	initialSize := fi.Size()

	// Open source database.
	src, err := bolt.Open(srcPath, 0444, nil)
	if err != nil {
		logger.Error("unable to open the src database")
		return err
	}
	defer src.Close()

	// Open destination database.
	dst, err := bolt.Open(dstPath, fi.Mode(), nil)
	if err != nil {
		logger.Error("unable to open the dst database")
		return err
	}
	defer dst.Close()

	// Run compaction.
	if err := bc.compact(dst, src); err != nil {
		logger.Error("unable to compact the database")
		return err
	}

	// Report stats on new size.
	fi, err = os.Stat(dstPath)
	if err != nil {
		logger.Error("unable to get the stat on the compacted database")
		return err
	} else if fi.Size() == 0 {
		return fmt.Errorf("zero db size")
	}

	logger.WithFields(log.Fields{
		"stat": fmt.Sprintf("%d -> %d bytes (gain=%.2fx)\n", initialSize, fi.Size(), float64(initialSize)/float64(fi.Size())),
	}).Info("Compact action ended!")

	// mv srcPath => bckPath
	if err := os.Rename(srcPath, bckPath); err != nil {
		return err
	}

	if err := os.Rename(dstPath, srcPath); err != nil {
		return err
	}

	// delete backup
	os.Remove(bckPath)

	return nil
}

func (bc *BoltCompacter) compact(dst, src *bolt.DB) error {
	// commit regularly, or we'll run out of memory for large datasets if using one transaction.
	var size int64
	tx, err := dst.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := bc.walk(src, func(keys [][]byte, k, v []byte, seq uint64) error {
		// On each key/value, check if we have exceeded tx size.
		sz := int64(len(k) + len(v))
		if size+sz > bc.TxMaxSize && bc.TxMaxSize != 0 {
			// Commit previous transaction.
			if err := tx.Commit(); err != nil {
				return err
			}

			// Start new transaction.
			tx, err = dst.Begin(true)
			if err != nil {
				return err
			}
			size = 0
		}
		size += sz

		// Create bucket on the root transaction if this is the first level.
		nk := len(keys)
		if nk == 0 {
			bkt, err := tx.CreateBucket(k)
			if err != nil {
				return err
			}
			if err := bkt.SetSequence(seq); err != nil {
				return err
			}
			return nil
		}

		// Create buckets on subsequent levels, if necessary.
		b := tx.Bucket(keys[0])
		if nk > 1 {
			for _, k := range keys[1:] {
				b = b.Bucket(k)
			}
		}

		// If there is no value then this is a bucket call.
		if v == nil {
			bkt, err := b.CreateBucket(k)
			if err != nil {
				return err
			}
			if err := bkt.SetSequence(seq); err != nil {
				return err
			}
			return nil
		}

		// Otherwise treat it as a key/value pair.
		return b.Put(k, v)
	}); err != nil {
		return err
	}

	return tx.Commit()
}

// walkFunc is the type of the function called for keys (buckets and "normal"
// values) discovered by Walk. keys is the list of keys to descend to the bucket
// owning the discovered key/value pair k/v.
type walkFunc func(keys [][]byte, k, v []byte, seq uint64) error

// walk walks recursively the bolt database db, calling walkFn for each key it finds.
func (bc *BoltCompacter) walk(db *bolt.DB, walkFn walkFunc) error {
	return db.View(func(tx *bolt.Tx) error {
		return tx.ForEach(func(name []byte, b *bolt.Bucket) error {
			return bc.walkBucket(b, nil, name, nil, b.Sequence(), walkFn)
		})
	})
}

func (bc *BoltCompacter) walkBucket(b *bolt.Bucket, keypath [][]byte, k, v []byte, seq uint64, fn walkFunc) error {
	// Execute callback.
	if err := fn(keypath, k, v, seq); err != nil {
		return err
	}

	// If this is not a bucket then stop.
	if v != nil {
		return nil
	}

	// Iterate over each child key/value.
	keypath = append(keypath, k)
	return b.ForEach(func(k, v []byte) error {
		if v == nil {
			bkt := b.Bucket(k)
			return bc.walkBucket(bkt, keypath, k, nil, bkt.Sequence(), fn)
		}
		return bc.walkBucket(b, keypath, k, v, b.Sequence(), fn)
	})
}
