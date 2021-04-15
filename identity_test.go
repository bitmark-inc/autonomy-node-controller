package main

import (
	"encoding/hex"
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/suite"
)

type PodIdentityTestSuite struct {
	suite.Suite
	KeyFile string
}

func (suite *PodIdentityTestSuite) TestNewPodIdentity() {
	i, err := NewPodIdentity()
	suite.NoError(err)

	suite.NoError(i.SaveKey(suite.KeyFile))

	keyBytes, err := os.ReadFile(suite.KeyFile)
	suite.NoError(err)

	var key KeyFile
	suite.NoError(json.Unmarshal(keyBytes, &key))

	privateKey, err := hex.DecodeString(key.PrivateKey)
	suite.NoError(err)
	suite.Equal(i.PrivateKey, privateKey)
}

func (suite *PodIdentityTestSuite) TestLoadPodIdentity() {
	privateKeyHex := "6bca03671bd851dd55f61e8ec867cbe49831aa3e3914b22a2990444a4b62c113"
	keyObject := map[string]string{
		"private_key": privateKeyHex,
	}
	b, err := json.Marshal(keyObject)
	suite.NoError(err)

	suite.NoError(os.WriteFile(suite.KeyFile, b, 0600))
	defer os.Remove(suite.KeyFile)

	i, err := new(PodIdentity).LoadKey(suite.KeyFile)
	suite.NoError(err)

	privateKey, err := hex.DecodeString(privateKeyHex)
	suite.NoError(err)
	suite.Equal(i.PrivateKey, privateKey)
}

func (suite *PodIdentityTestSuite) TestCreateOrLoadPodIdentityWithExistingKey() {
	privateKeyHex := "6bca03671bd851dd55f61e8ec867cbe49831aa3e3914b22a2990444a4b62c113"
	keyObject := map[string]string{
		"private_key": privateKeyHex,
	}
	b, err := json.Marshal(keyObject)
	suite.NoError(err)

	suite.NoError(os.WriteFile(suite.KeyFile, b, 0600))
	defer os.Remove(suite.KeyFile)

	i, created, err := CreateOrLoadPodIdentityFromKey(suite.KeyFile)
	suite.False(created)
	suite.NoError(err)

	privateKey, err := hex.DecodeString(privateKeyHex)
	suite.NoError(err)
	suite.Equal(i.PrivateKey, privateKey)
}

func (suite *PodIdentityTestSuite) TestCreateOrLoadPodIdentityWithoutExistingKey() {
	i, created, err := CreateOrLoadPodIdentityFromKey(suite.KeyFile)
	suite.False(created)
	suite.NoError(err)

	keyBytes, err := os.ReadFile(suite.KeyFile)
	suite.NoError(err)

	var key KeyFile
	suite.NoError(json.Unmarshal(keyBytes, &key))

	privateKey, err := hex.DecodeString(key.PrivateKey)
	suite.NoError(err)
	suite.Equal(i.PrivateKey, privateKey)
}

func TestPodIdentityTestSuite(t *testing.T) {
	suite.Run(t, &PodIdentityTestSuite{
		KeyFile: "keyfile_test.json",
	})
}
