package main

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"os"
)

func runKeygen() error {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return fmt.Errorf("generating key pair: %w", err)
	}

	block := &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: []byte(priv),
	}
	if err := os.WriteFile("signing.key", pem.EncodeToMemory(block), 0600); err != nil {
		return fmt.Errorf("writing signing.key: %w", err)
	}

	fmt.Println("Private key written to: signing.key")
	fmt.Println("Store it in CI secrets as MANIFEST_SIGNING_KEY — never commit it.")
	fmt.Println()
	fmt.Println("Public key (paste as --key flag and MANIFEST_PUBLIC_KEY secret):")
	fmt.Println(base64.StdEncoding.EncodeToString(pub))
	return nil
}
