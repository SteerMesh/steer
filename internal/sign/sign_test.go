package sign

import (
	"crypto/ed25519"
	"crypto/x509"
	"encoding/pem"
	"os"
	"path/filepath"
	"testing"

	"github.com/SteerMesh/steer/internal/bundle"
)

func TestSignVerify(t *testing.T) {
	dir := t.TempDir()
	privPath := filepath.Join(dir, "priv.pem")
	pubPath := filepath.Join(dir, "pub.pem")

	pub, priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatal(err)
	}
	privDER, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(privPath, pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: privDER}), 0644); err != nil {
		t.Fatal(err)
	}
	pubDER, err := x509.MarshalPKIXPublicKey(pub)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(pubPath, pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubDER}), 0644); err != nil {
		t.Fatal(err)
	}

	canonical := []byte(`{"version":"1.0","packs":[{"name":"p","version":"1.0.0"}],"files":[]}`)
	algo, keyID, value, err := Sign(canonical, privPath)
	if err != nil {
		t.Fatal(err)
	}
	if algo != AlgorithmEd25519 || keyID != "default" || value == "" {
		t.Errorf("Sign: algo=%q keyID=%q value empty=%v", algo, keyID, value == "")
	}

	// Verify with public key
	sb := &bundle.Signature{Algorithm: algo, KeyID: keyID, Value: value}
	if err := Verify(canonical, sb, pubPath); err != nil {
		t.Fatal(err)
	}

	// Tampered canonical should fail
	if err := Verify([]byte("tampered"), sb, pubPath); err == nil {
		t.Error("expected verification to fail for tampered data")
	}
}
