package key

import (
	"crypto/elliptic"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"strings"

	"github.com/bitmark-inc/secp256k1-go"
	"github.com/multiformats/go-multicodec"

	"github.com/bitmark-inc/autonomy-pod-controller/utils"
)

const (
	didKeyPrefix = "did:key:z"
)

// Sign returns the signature of a message using the given private key
func Sign(privateKey []byte, message string) (string, error) {
	hash := sha256.Sum256([]byte(message))
	s, err := secp256k1.Sign(hash[:], privateKey)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(s), nil
}

// VerifySignature validates the signature of a message with the given public key
func VerifySignature(did, message, signature string) bool {
	pub, err := PublicKeyFromDID(did)
	if err != nil {
		return false
	}

	sig, err := hex.DecodeString(signature)
	if err != nil {
		return false
	}

	hash := sha256.Sum256([]byte(message))

	return secp256k1.VerifySignature(pub, hash[:], sig)
}

// PublicKey return public key bytes of a private key
func PublicKey(privateKey []byte) []byte {
	x, y := secp256k1.S256().ScalarBaseMult(privateKey)
	return elliptic.MarshalCompressed(secp256k1.S256(), x, y)
}

// DID returns autonomy DID key of a private key
func DID(privateKey []byte) string {
	publicKeyBytes := PublicKey(privateKey)

	buf := make([]byte, binary.MaxVarintLen64)
	size := binary.PutUvarint(buf, uint64(multicodec.Secp256k1Pub))
	prefix := buf[:size]

	return didKeyPrefix + utils.ToBase58(append(prefix, publicKeyBytes...))
}

// PublicKeyFromDID return public key bytes of a give DID key.
func PublicKeyFromDID(did string) ([]byte, error) {
	if !strings.HasPrefix(did, didKeyPrefix) {
		return nil, errors.New("invalid did")
	}

	encodedBytes := utils.FromBase58(strings.TrimPrefix(did, didKeyPrefix))
	publicKeyType, publicKeyBytesStart := binary.Uvarint(encodedBytes)
	if publicKeyType != uint64(multicodec.Secp256k1Pub) {
		return nil, errors.New("invalid key type")
	}

	return encodedBytes[publicKeyBytesStart:], nil
}
