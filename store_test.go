package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/suite"
)

type StoreTestSuite struct {
	suite.Suite
	dbFile string
}

func (s *StoreTestSuite) TearDownSuite() {
	os.RemoveAll(s.dbFile)
}

func (s *StoreTestSuite) TestMember() {
	memberDID := "did:key:family-member"

	store := NewBoltStore(s.dbFile)

	// add the member with limited access
	err := store.UpdateMemberAccessMode(memberDID, AccessModeLimited)
	s.NoError(err)
	mode := store.AccessMode(memberDID)
	s.Equal(AccessModeLimited, mode)

	// update its access to minimal
	err = store.UpdateMemberAccessMode(memberDID, AccessModeMinimal)
	s.NoError(err)
	mode = store.AccessMode(memberDID)
	s.Equal(AccessModeMinimal, mode)

	// remove its access
	err = store.RemoveMember(memberDID)
	s.NoError(err)
	mode = store.AccessMode(memberDID)
	s.Equal(AccessModeNotApplicant, mode)
}

func TestStoreTestSuite(t *testing.T) {
	suite.Run(t, &StoreTestSuite{
		dbFile: "test.db",
	})
}
