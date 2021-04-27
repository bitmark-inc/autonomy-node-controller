package main

import (
	"encoding/binary"

	bolt "go.etcd.io/bbolt"
)

var (
	BucketMember = []byte("members")
)

type Store interface {
	UpdateMemberAccessMode(memberDID string, accessMode AccessMode) error
	RemoveMember(memberDID string) error
	AccessMode(memberDID string) AccessMode
}

type BoltStore struct {
	db *bolt.DB
}

func NewBoltStore(path string) *BoltStore {
	db, err := bolt.Open(path, 0666, nil)
	if err != nil {
		panic(err)
	}

	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(BucketMember)
		return err
	})
	if err != nil {
		panic(err)
	}

	return &BoltStore{db}
}

func (s *BoltStore) UpdateMemberAccessMode(memberDID string, accessMode AccessMode) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(BucketMember)
		v := make([]byte, 8)
		binary.BigEndian.PutUint64(v, uint64(accessMode))
		return b.Put([]byte(memberDID), v)
	})
}

func (s *BoltStore) RemoveMember(memberDID string) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(BucketMember)
		return b.Delete([]byte(memberDID))
	})
}

func (s *BoltStore) AccessMode(memberDID string) AccessMode {
	var mode AccessMode
	s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(BucketMember)
		v := b.Get([]byte(memberDID))
		if v == nil {
			mode = AccessModeNotApplicant
		} else {
			mode = AccessMode(binary.BigEndian.Uint64(v))
		}
		return nil
	})
	return mode
}
