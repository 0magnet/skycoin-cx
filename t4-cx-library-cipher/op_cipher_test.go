//go:build !nocipher

package cipher

import (
	"testing"

	"github.com/skycoin/skycoin-cx/t0-cx-core-digest"
	"github.com/skycoin/skycoin/src/cipher"
)

// Signatures made by cipher.SignHash of Skycoin verify, and a changed hash,
// signature or key, or one of the wrong length, does not.
func TestCipherSignatureVerifySkycoinSignatures(t *testing.T) {
	for i := range 8 {
		pubKey, secKey := cipher.GenerateKeyPair()
		otherPubKey, _ := cipher.GenerateKeyPair()
		digest := cipher.SHA256(digest.CxHashOfBytes([]byte{byte(i), 'c', 'x'}))
		signature := cipher.MustSignHash(digest, secKey)

		if !SignatureVerify(digest[:], signature[:], pubKey[:]) {
			t.Fatalf("pair %d: a Skycoin signature does not verify", i)
		}
		if SignatureVerify(digest[:], signature[:], otherPubKey[:]) {
			t.Errorf("pair %d: verifies with another key", i)
		}
		changedDigest := digest
		changedDigest[0] ^= 1
		if SignatureVerify(changedDigest[:], signature[:], pubKey[:]) {
			t.Errorf("pair %d: verifies another hash", i)
		}
		changedSignature := signature
		changedSignature[5] ^= 1
		if SignatureVerify(digest[:], changedSignature[:], pubKey[:]) {
			t.Errorf("pair %d: a changed signature verifies", i)
		}
		if SignatureVerify(digest[:31], signature[:], pubKey[:]) || SignatureVerify(digest[:], signature[:64], pubKey[:]) || SignatureVerify(digest[:], signature[:], pubKey[:32]) {
			t.Errorf("pair %d: an input of the wrong length verifies", i)
		}
		if SignatureVerify(nil, nil, nil) {
			t.Errorf("empty inputs verify")
		}
	}
}

// SignatureVerify allocates only what the Skycoin call allocates.
func TestCipherSignatureVerifyAllocs(t *testing.T) {
	pubKey, secKey := cipher.GenerateKeyPair()
	digest := cipher.SHA256(digest.CxHashOfBytes([]byte("abc")))
	signature := cipher.MustSignHash(digest, secKey)
	skycoin := testing.AllocsPerRun(10, func() { cipher.VerifyPubKeySignedHash(pubKey, signature, digest) })
	views := testing.AllocsPerRun(10, func() { SignatureVerify(digest[:], signature[:], pubKey[:]) })
	if views != skycoin {
		t.Errorf("%.0f allocations per verify, the Skycoin call makes %.0f", views, skycoin)
	}
}

func BenchmarkCipherSignatureVerify(b *testing.B) {
	pubKey, secKey := cipher.GenerateKeyPair()
	digest := cipher.SHA256(digest.CxHashOfBytes([]byte("abc")))
	signature := cipher.MustSignHash(digest, secKey)
	b.ReportAllocs()
	for b.Loop() {
		SignatureVerify(digest[:], signature[:], pubKey[:])
	}
}
