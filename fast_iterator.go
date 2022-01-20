package iavl

import (
	"errors"

	dbm "github.com/tendermint/tm-db"
)

// Iterator is a dbm.Iterator for ImmutableTree
type FastIterator struct {
	start, end []byte

	valid bool

	ascending bool

	err error

	ndb *nodeDB

	nextFastNode *FastNode

	fastIterator dbm.Iterator
}

var _ dbm.Iterator = &FastIterator{}

func NewFastIterator(start, end []byte, ascending bool, ndb *nodeDB) *FastIterator {
	iter := &FastIterator{
		start:       start,
		end:         end,
		err:         nil,
		ascending:   ascending,
		ndb:         ndb,
		nextFastNode: nil,
		fastIterator: nil,
	}
	iter.Next()
	return iter
}

// Domain implements dbm.Iterator.
func (iter *FastIterator) Domain() ([]byte, []byte) {
	if iter.fastIterator == nil {
		return iter.start, iter.end
	}

	start, end := iter.fastIterator.Domain()

	if start != nil {
		start = start[1:]
		if len(start) == 0 {
			start = nil
		}
	}

	if end != nil {
		end = end[1:]
		if len(end) == 0 {
			end = nil
		}
	}

	return start, end
}

// Valid implements dbm.Iterator.
func (iter *FastIterator) Valid() bool {
	if iter.fastIterator == nil || !iter.fastIterator.Valid() {
		return false
	}

	return iter.valid
}

// Key implements dbm.Iterator
func (iter *FastIterator) Key() []byte {
	if iter.valid {
		return iter.nextFastNode.key
	}
	return nil
}

// Value implements dbm.Iterator
func (iter *FastIterator) Value() []byte {
	if iter.valid {
		return iter.nextFastNode.value
	}
	return nil
}

// Next implements dbm.Iterator
func (iter *FastIterator) Next() {
	if iter.ndb == nil {
		iter.err = errors.New("nodeDB is nil")
		iter.valid = false
		return
	}

	if iter.fastIterator == nil {
		iter.fastIterator, iter.err = iter.ndb.getFastIterator(iter.start, iter.end, iter.ascending)
		iter.valid = true
	} else {
		iter.fastIterator.Next()
	}

	if iter.err == nil {
		iter.err = iter.fastIterator.Error()
	}

	iter.valid = iter.valid && iter.fastIterator.Valid()
	if iter.valid {
		iter.nextFastNode, iter.err = DeserializeFastNode(iter.fastIterator.Key()[1:], iter.fastIterator.Value())
		iter.valid = iter.err == nil
	}
}

// Close implements dbm.Iterator
func (iter *FastIterator) Close() error {
	iter.fastIterator = nil
	iter.valid = false

	if iter.fastIterator != nil {
		iter.err = iter.fastIterator.Close()
	}
	return iter.err
}

// Error implements dbm.Iterator
func (iter *FastIterator) Error() error {
	return iter.err
}
