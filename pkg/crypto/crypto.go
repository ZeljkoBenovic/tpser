package crypto

import (
	"crypto/ecdsa"
	"fmt"

	hdwallet "github.com/miguelmota/go-ethereum-hdwallet"
)

// GenerateFromMnemonic generates a number of private keys from a mnemonic
func GenerateFromMnemonic(mnemonic string, count uint64) ([]*ecdsa.PrivateKey, error) {
	wallet, err := hdwallet.NewFromMnemonic(mnemonic)
	if err != nil {
		return nil, fmt.Errorf("could not process mnemonic: %w", err)
	}

	keys := make([]*ecdsa.PrivateKey, 0, count)
	for i := uint64(0); i < count; i++ {
		path, err := hdwallet.ParseDerivationPath(fmt.Sprintf("m/44'/60'/0'/0/%d", i))
		if err != nil {
			return nil, fmt.Errorf("could not derive account path: %w", err)
		}

		account, err := wallet.Derive(path, false)
		if err != nil {
			return nil, fmt.Errorf("could not derive account from mnemonic: %w", err)
		}

		key, err := wallet.PrivateKey(account)
		if err != nil {
			return nil, fmt.Errorf("unable to get private key: %w", err)
		}

		keys = append(keys, key)
	}

	return keys, nil
}
