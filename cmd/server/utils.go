package main

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"slices"
	"strings"

	"github.com/zanz1n/blog/internal/utils/errutils"
)

func jwtKeyPair() (ed25519.PublicKey, ed25519.PrivateKey, error) {
	priv := os.Getenv("JWT_PRIVATE_KEY")
	pub := os.Getenv("JWT_PUBLIC_KEY")

	privFunc := func() (ed25519.PrivateKey, error) {
		_, priv, err := ed25519.GenerateKey(rand.Reader)
		return priv, err
	}

	privKeyRaw, err := envKeyPair(priv, ed25519.PrivateKeySize, true, privFunc)
	if err != nil {
		return nil, nil, fmt.Errorf("jwt private key: %w", err)
	}
	privKey := ed25519.PrivateKey(privKeyRaw)

	pubFunc := func() (ed25519.PublicKey, error) {
		return privKey.Public().(ed25519.PublicKey), nil
	}

	pubKeyRaw, err := envKeyPair(pub, ed25519.PublicKeySize, false, pubFunc)
	if err != nil {
		return nil, nil, fmt.Errorf("jwt public key: %w", err)
	}
	pubKey := ed25519.PublicKey(pubKeyRaw)

	if !slices.Equal(privKey.Public().(ed25519.PublicKey), pubKey) {
		return nil, nil, errors.New(
			"jwt ed25519 key: private key and public key doesn't match",
		)
	}

	return pubKey, privKey, nil
}

func envKeyPair[T ~[]byte](
	value string,
	exlen int,
	private bool,
	f func() (T, error),
) (T, error) {
	if strings.HasPrefix(value, "file:") {
		value = value[len("file:"):]

		data, err := parseKeyFile[T](value, private)
		if errors.Is(err, os.ErrNotExist) {
			data, err = f()
			if err == nil {
				err = marshalKeyFile(value, data, private)
			}
		} else if err == nil {
			if len(data) != exlen {
				return nil, fmt.Errorf(
					"invalid length: expected %d, but got %d",
					exlen, len(data),
				)
			}
		}

		return data, err
	}

	data, err := base64.StdEncoding.DecodeString(value)
	if err != nil {
		return nil, err
	}

	if len(data) != exlen {
		return nil, fmt.Errorf(
			"invalid length: expected %d, but got %d",
			exlen, len(data),
		)
	}

	return data, err
}

func parseKeyFile[T any](name string, private bool) (res T, err error) {
	file, err := os.ReadFile(name)
	if err != nil {
		return
	}

	block, _ := pem.Decode(file)

	var key any
	if private {
		key, err = x509.ParsePKCS8PrivateKey(block.Bytes)
	} else {
		key, err = x509.ParsePKIXPublicKey(block.Bytes)
	}

	if err != nil {
		return
	}
	res = key.(T)

	return
}

func marshalKeyFile(name string, key any, private bool) (err error) {
	var data []byte
	if private {
		data, err = x509.MarshalPKCS8PrivateKey(key)
	} else {
		data, err = x509.MarshalPKIXPublicKey(key)
	}

	if err != nil {
		return
	}

	file, err := os.Create(name)
	if err != nil {
		return
	}

	mode := "PUBLIC"
	if private {
		mode = "PRIVATE"
	}

	err = pem.Encode(file, &pem.Block{
		Type:  fmt.Sprintf("BEGIN %s KEY", mode),
		Bytes: data,
	})
	return
}

func setenv(key, value string) {
	if err := os.Setenv(key, value); err != nil {
		fatal(err)
	}
}

func fatal(err any) {
	exitCode := 1
	if err, ok := err.(error); ok {
		exitCode = errutils.Os(err).OsStatus()
	}

	slog.Error(fmt.Sprint("FATAL: ", err))
	os.Exit(exitCode)
}
