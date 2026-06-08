package keys

import (
	"context"
	"crypto"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/aws/aws-sdk-go-v2/service/kms/types"
	"github.com/clineomx/trussrod/utils/encryption"
)

type KMS struct {
	client   *kms.Client
	keyARN   string
	localKEK []byte
}

func (k *KMS) Decrypt(ctx context.Context, target []byte) ([]byte, error) {
	input := &kms.DecryptInput{
		CiphertextBlob: target,
	}

	decrypted, err := k.client.Decrypt(ctx, input)
	if err != nil {
		return nil, err
	}
	return decrypted.Plaintext, nil
}

func (k *KMS) CreateDEK(ctx context.Context) ([]byte, []byte, error) {
	input := &kms.GenerateDataKeyInput{
		KeyId:   aws.String(k.keyARN),
		KeySpec: "AES_256",
	}
	out, err := k.client.GenerateDataKey(ctx, input)
	if err != nil {
		return nil, nil, err
	}
	return out.Plaintext, out.CiphertextBlob, nil
}

func (k *KMS) Wrap(ctx context.Context, input []byte) ([]byte, error) {
	block, err := aes.NewCipher(k.localKEK)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	wrapped := gcm.Seal(nonce, nonce, input, nil)
	return wrapped, nil

}

func (k *KMS) Unwrap(ctx context.Context, wrapped []byte) ([]byte, error) {
	block, err := aes.NewCipher(k.localKEK)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(wrapped) < nonceSize {
		return nil, fmt.Errorf("wrapped key too short")
	}

	nonce, ciphertext := wrapped[:nonceSize], wrapped[nonceSize:]

	dek, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to unwrap local key: %w", err)
	}

	return dek, nil
}

func (k *KMS) Key() string {
	return k.keyARN
}

type KMSSigner struct {
	key    string
	client *kms.Client
}

type SignOutput struct {
	KeyId     string
	Digest    []byte
	Signature []byte
	Algorithm string
}

func (k *KMS) CreateSigner(key string) Signer {
	return &KMSSigner{key: key, client: k.client}
}

func (k *KMSSigner) Sign(ctx context.Context, input []byte) (*SignOutput, error) {
	digest := encryption.GetSHA256(input)
	result, err := k.client.Sign(ctx, &kms.SignInput{
		KeyId:            aws.String(k.key),
		Message:          digest,
		MessageType:      types.MessageTypeDigest,
		SigningAlgorithm: types.SigningAlgorithmSpecRsassaPssSha256,
	})
	if err != nil {
		return nil, err
	}

	return &SignOutput{
		KeyId:     *result.KeyId,
		Digest:    digest,
		Signature: result.Signature,
		Algorithm: string(types.SigningAlgorithmSpecRsassaPssSha256),
	}, nil
}

func (k *KMSSigner) Verify(ctx context.Context, message, signature []byte) (bool, error) {
	pkOut, err := k.client.GetPublicKey(ctx, &kms.GetPublicKeyInput{
		KeyId: aws.String(k.key),
	})
	if err != nil {
		return false, err
	}

	pub, err := x509.ParsePKIXPublicKey(pkOut.PublicKey)
	if err != nil {
		return false, err
	}
	rsaPub, ok := pub.(*rsa.PublicKey)
	if !ok {
		return false, fmt.Errorf("public key is not RSA")
	}

	if err := rsa.VerifyPSS(
		rsaPub,
		crypto.SHA256,
		message,
		signature,
		&rsa.PSSOptions{SaltLength: rsa.PSSSaltLengthEqualsHash, Hash: crypto.SHA256}); err != nil {
		return false, fmt.Errorf("signature verification failed: %w", err)
	}

	return true, nil
}

func NewKMSClient(key string, cfg *aws.Config) (*KMS, error) {
	kek := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, kek); err != nil {
		return nil, fmt.Errorf("failed to generate local KEK: %w", err)
	}
	return &KMS{client: kms.NewFromConfig(*cfg), keyARN: key, localKEK: kek}, nil
}
