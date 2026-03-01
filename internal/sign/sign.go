package sign

import (
	"crypto/ed25519"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"os"

	"github.com/SteerMesh/steer/internal/bundle"
)

const AlgorithmEd25519 = "Ed25519"

// Sign produces a signature over canonical manifest bytes using an Ed25519 private key (PEM PKCS#8).
// Returns algorithm, keyID (derived from key or "default"), and base64 signature value.
func Sign(canonical []byte, privateKeyPath string) (algorithm, keyID, value string, err error) {
	pemData, err := os.ReadFile(privateKeyPath)
	if err != nil {
		return "", "", "", fmt.Errorf("read private key: %w", err)
	}
	block, _ := pem.Decode(pemData)
	if block == nil {
		return "", "", "", fmt.Errorf("no PEM block in private key file")
	}
	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return "", "", "", fmt.Errorf("parse private key: %w", err)
	}
	priv, ok := key.(ed25519.PrivateKey)
	if !ok {
		return "", "", "", fmt.Errorf("key is not Ed25519")
	}
	sig := ed25519.Sign(priv, canonical)
	keyID = "default"
	if block.Headers["Key-ID"] != "" {
		keyID = block.Headers["Key-ID"]
	}
	return AlgorithmEd25519, keyID, base64.StdEncoding.EncodeToString(sig), nil
}

// Verify checks the signature over canonical manifest bytes using an Ed25519 public key (PEM PKCS#8 or raw).
func Verify(canonical []byte, sig *bundle.Signature, publicKeyPath string) error {
	if sig == nil {
		return nil
	}
	if sig.Algorithm != AlgorithmEd25519 {
		return fmt.Errorf("unsupported signature algorithm: %s", sig.Algorithm)
	}
	sigBytes, err := base64.StdEncoding.DecodeString(sig.Value)
	if err != nil {
		return fmt.Errorf("decode signature: %w", err)
	}
	pemData, err := os.ReadFile(publicKeyPath)
	if err != nil {
		return fmt.Errorf("read public key: %w", err)
	}
	block, _ := pem.Decode(pemData)
	if block == nil {
		return fmt.Errorf("no PEM block in public key file")
	}
	key, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return fmt.Errorf("parse public key: %w", err)
	}
	pub, ok := key.(ed25519.PublicKey)
	if !ok {
		return fmt.Errorf("key is not Ed25519")
	}
	if !ed25519.Verify(pub, canonical, sigBytes) {
		return fmt.Errorf("signature verification failed")
	}
	return nil
}
