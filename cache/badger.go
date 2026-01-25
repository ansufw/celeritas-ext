package cache

import (
	"time"

	"github.com/dgraph-io/badger/v4"
)

type Badger struct {
	Conn   *badger.DB
	Prefix string
}

func (b *Badger) Has(str string) (bool, error) {
	_, err := b.Get(str)
	if err != nil {
		return false, nil
	}
	return true, nil
}

func (b *Badger) Get(str string) (any, error) {
	var fromCache []byte
	err := b.Conn.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(str))
		if err != nil {
			return err
		}
		err = item.Value(func(val []byte) error {
			fromCache = append(fromCache, val...)
			return nil
		})
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	decoded, err := decode(string(fromCache))
	if err != nil {
		return nil, err
	}
	item := decoded[str]
	return item, nil
}

func (b *Badger) Set(str string, value any, expires ...int) error {
	entry := Entry{}
	entry[str] = value
	encoded, err := encode(entry)
	if err != nil {
		return err
	}

	if len(expires) > 0 {
		err = b.Conn.Update(func(txn *badger.Txn) error {
			e := badger.NewEntry([]byte(str), encoded).WithTTL(time.Second * time.Duration(expires[0]))
			err = txn.SetEntry(e)
			return err
		})
	} else {
		err = b.Conn.Update(func(txn *badger.Txn) error {
			e := badger.NewEntry([]byte(str), encoded)
			err = txn.SetEntry(e)
			return err
		})
	}

	return nil
}

func (b *Badger) Forget(str string) error {
	err := b.Conn.Update(func(txn *badger.Txn) error {
		return txn.Delete([]byte(str))
	})
	return err
}

func (b *Badger) EmptyByMatch(str string) error {

	return b.emptyByMatch(str)
}

func (b *Badger) Empty() error {

	return b.emptyByMatch("")
}

func (b *Badger) emptyByMatch(str string) error {

	deleteKeys := func(keyForDelete [][]byte) error {
		if err := b.Conn.Update(func(txn *badger.Txn) error {
			for _, key := range keyForDelete {
				if err := txn.Delete(key); err != nil {
					return err
				}
			}
			return nil
		}); err != nil {
			return err
		}
		return nil
	}

	collectSize := 100_000
	err := b.Conn.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.AllVersions = false
		opts.PrefetchValues = false
		iter := txn.NewIterator(opts)
		defer iter.Close()

		keysForDelete := make([][]byte, 0, collectSize)
		keysCollected := 0
		for iter.Seek([]byte(str)); iter.ValidForPrefix([]byte(str)); iter.Next() {
			key := iter.Item().KeyCopy(nil)
			keysForDelete = append(keysForDelete, key)
			keysCollected++
			if keysCollected == collectSize {
				if err := deleteKeys(keysForDelete); err != nil {
					return err
				}
				keysForDelete = make([][]byte, 0, collectSize)
				keysCollected = 0
			}

		}
		if keysCollected > 0 {
			if err := deleteKeys(keysForDelete); err != nil {
				return err
			}
		}

		return nil
	})
	return err
}
