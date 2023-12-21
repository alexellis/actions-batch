package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/google/go-github/v57/github"
	"golang.org/x/crypto/nacl/box"
)

func createSecrets(ctx context.Context, client *github.Client, owner, repoName, secretsFrom string) (map[string]string, error) {
	mapped := map[string]string{}

	key, _, err := client.Actions.GetRepoPublicKey(ctx, owner, repoName)
	if err != nil {
		return nil, err
	}

	dir, err := os.ReadDir(secretsFrom)
	if err != nil {
		return nil, err
	}

	for _, f := range dir {
		if f.IsDir() {
			continue
		}

		secretName := strings.ToUpper(strings.ReplaceAll(f.Name(), "-", "_"))
		secretData, _ := os.ReadFile(path.Join(secretsFrom, f.Name()))

		encryptedVal, err := encryptSecret(key, strings.TrimSpace(string(secretData)))
		if err != nil {
			return nil, err
		}

		if _, err := client.Actions.CreateOrUpdateRepoSecret(ctx, owner, repoName, &github.EncryptedSecret{
			Name:           secretName,
			EncryptedValue: encryptedVal,
			KeyID:          key.GetKeyID(),
		}); err != nil {
			return nil, err
		} else {
			fmt.Printf("Created secret: %s (%s)\n", secretName, path.Join(secretsFrom, f.Name()))
			mapped[secretName] = string(secretName)
		}

	}

	return mapped, nil
}

func encryptSecret(publicKey *github.PublicKey, secret string) (string, error) {
	publicKeyBytes, err := base64.StdEncoding.DecodeString(publicKey.GetKey())
	if err != nil {
		return "", err
	}
	// Decode the public key
	var publicKeyDecoded [32]byte
	copy(publicKeyDecoded[:], publicKeyBytes)

	encrypted, err := box.SealAnonymous(nil, []byte(secret), (*[32]byte)(publicKeyBytes), rand.Reader)
	if err != nil {
		return "", err
	}

	encryptedBase64 := base64.StdEncoding.EncodeToString(encrypted)
	return encryptedBase64, nil
}
