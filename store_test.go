// SPDX-License-Identifier: ISC
// Copyright (c) 2019-2021 Bitmark Inc.
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/suite"
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

func TestStoreTestSuite(t *testing.T) {
	suite.Run(t, &StoreTestSuite{
		dbFile: "test.db",
		store:  NewBoltStore("test.db"),
	})
}
