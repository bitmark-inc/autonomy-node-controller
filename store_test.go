// SPDX-License-Identifier: ISC
// Copyright (c) 2019-2021 Bitmark Inc.
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package main

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/suite"
	"go.etcd.io/bbolt"
)

type StoreTestSuite struct {
	suite.Suite
	dbFile string
	store  *BoltStore
}

func (s *StoreTestSuite) TearDownSuite() {
	os.RemoveAll(s.dbFile)
}

func (s *StoreTestSuite) TestBinding() {
	did := "did:key:au-user"

	bound := s.store.HasBinding(did)
	s.False(bound)

	err := s.store.SetBinding(did, "123")
	s.NoError(err)

	nonce := s.store.BindingNonce(did)
	s.Equal("123", nonce)

	err = s.store.CompleteBinding(did)
	s.NoError(err)

	bound = s.store.HasBinding(did)
	s.True(bound)
}

func (s *StoreTestSuite) TestMember() {
	memberDID := "did:key:family-member"

	// add the member with limited access
	err := s.store.UpdateMemberAccessMode(memberDID, AccessModeLimited)
	s.NoError(err)
	mode := s.store.MemberAccessMode(memberDID)
	s.Equal(AccessModeLimited, mode)

	// update its access to minimal
	err = s.store.UpdateMemberAccessMode(memberDID, AccessModeMinimal)
	s.NoError(err)
	mode = s.store.MemberAccessMode(memberDID)
	s.Equal(AccessModeMinimal, mode)

	// remove its access
	err = s.store.RemoveMember(memberDID)
	s.NoError(err)
	mode = s.store.MemberAccessMode(memberDID)
	s.Equal(AccessModeNotApplicant, mode)
}

// TestSaveAndLoadRequestsUsage tests saving and loading usage data
func (s *StoreTestSuite) TestSaveAndLoadRequestsUsage() {

	// test `loadRequest` with a fresh db (no usage data)
	{
		_, err := s.store.LoadRequestsUsage()
		s.Error(err, ErrKeyNotFound.Error())
	}

	u := NewUsage()
	u.CountRequests(1)

	s.NoError(s.store.SaveRequestsUsage(u))

	testUsageBytes, err := json.Marshal(u)
	s.NoError(err)

	s.store.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket(bucketUsage)
		s.NotNil(b)

		usageBytes := b.Get(keyRequest)
		s.NotNil(usageBytes)

		s.EqualValues(testUsageBytes, usageBytes)
		return nil
	})

	loadedUsage, err := s.store.LoadRequestsUsage()
	s.NoError(err)

	loadedUsageBytes, err := json.Marshal(loadedUsage)
	s.NoError(err)
	s.EqualValues(testUsageBytes, loadedUsageBytes)
}

func TestStoreTestSuite(t *testing.T) {
	suite.Run(t, &StoreTestSuite{
		dbFile: "test.db",
		store:  NewBoltStore("test.db"),
	})
}
