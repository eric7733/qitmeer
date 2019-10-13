package qx

import (
	"encoding/hex"
	"fmt"
	"github.com/Qitmeer/qitmeer-lib/crypto/ecc"
	"github.com/Qitmeer/qitmeer/crypto/bip32"
	"github.com/Qitmeer/qitmeer/crypto/seed"
)

func NewEntropy(size uint) (string, error) {
	s, err := seed.GenerateSeed(uint16(size))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", s), nil
}

func EcNew(curve string, entropyStr string) (string, error) {
	entropy, err := hex.DecodeString(entropyStr)
	if err != nil {
		return "", err
	}
	switch curve {
	case "secp256k1":
		masterKey, err := bip32.NewMasterKey(entropy)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%x", masterKey.Key[:]), nil
	default:
		return "", fmt.Errorf("unknown curve : %s", curve)
	}
}

func EcPrivateKeyToEcPublicKey(uncompressed bool, privateKeyStr string) (string, error) {
	data, err := hex.DecodeString(privateKeyStr)
	if err != nil {
		return "", err
	}
	_, pubKey := ecc.Secp256k1.PrivKeyFromBytes(data)
	var key []byte
	if uncompressed {
		key = pubKey.SerializeUncompressed()
	} else {
		key = pubKey.SerializeCompressed()
	}
	return fmt.Sprintf("%x", key[:]), nil
}
