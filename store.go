// SPDX-License-Identifier: ISC
// Copyright (c) 2019-2021 Bitmark Inc.
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"encoding/binary"

	bolt "go.etcd.io/bbolt"
)

var (
	bucketBinding = []byte("bindings")
	bucketMember  = []byte("members")

	valueTrue  = []byte("true")
	valueFalse = []byte("false")
)

type Store interface {
	SetBinding(did, nonce string) error
	BindingNonce(did string) string
	CompleteBinding(did string) error
	HasBinding(did string) bool
	UpdateMemberAccessMode(memberDID string, accessMode AccessMode) error
	RemoveMember(memberDID string) error
	MemberAccessMode(memberDID string) AccessMode
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
		if _, err := tx.CreateBucketIfNotExists(bucketBinding); err != nil {
			return err
		}
		if _, err := tx.CreateBucketIfNotExists(bucketMember); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		panic(err)
	}

	return &BoltStore{db}
}

func (s *BoltStore) SetBinding(did, nonce string) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketBinding)
		return b.Put([]byte(did), []byte(nonce))
	})
}

func (s *BoltStore) BindingNonce(did string) string {
	var nonce string
	s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketBinding)
		nonce = string(b.Get([]byte(did)))
		return nil
	})
	return nonce
}

func (s *BoltStore) CompleteBinding(did string) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketBinding)
		return b.Put([]byte(did), valueTrue)
	})
}

func (s *BoltStore) HasBinding(did string) bool {
	var bound bool
	s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketBinding)
		v := b.Get([]byte(did))
		bound = bytes.Equal(v, valueTrue)
		return nil
	})
	return bound
}

func (s *BoltStore) UpdateMemberAccessMode(memberDID string, accessMode AccessMode) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketMember)
		v := make([]byte, 8)
		binary.BigEndian.PutUint64(v, uint64(accessMode))
		return b.Put([]byte(memberDID), v)
	})
}

func (s *BoltStore) RemoveMember(memberDID string) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketMember)
		return b.Delete([]byte(memberDID))
	})
}

func (s *BoltStore) MemberAccessMode(memberDID string) AccessMode {
	var mode AccessMode
	s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketMember)
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
