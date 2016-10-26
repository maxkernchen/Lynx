package mypgp

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"time"

	"golang.org/x/crypto/openpgp"
	"golang.org/x/crypto/openpgp/armor"
	"golang.org/x/crypto/openpgp/packet"
)

// Decode - this function decodes using the openpgp library.
// @param byte[] key - This is the key to be decoded
// @param byte[] passphrase - This parameter will be used if the key is encrypted with a passphrase
// @param io.Reader[] src - This parameter will be used to read the encrypted data
// @param io.Writer[] dest - This parameter will be used to write the unencrypted data
func Decode(key, passphrase []byte, src io.Reader, dest io.Writer) error {
	dec := &Decoder{key, passphrase}
	return dec.Decode(src, dest)
}

// Decoder - A struct which represents a key to be decoded.
type Decoder struct {
	Key        []byte
	Passphrase []byte
}

// Decode - this function wraps around a Decoder object to decode the passed in data.
// @param io.Reader[] src - This parameter will be used to read the encrypted data
// @param io.Writer[] dest - This parameter will be used to write the unencrypted data
func (d *Decoder) Decode(r io.Reader, w io.Writer) error {
	entitylist, err := openpgp.ReadArmoredKeyRing(bytes.NewBuffer(d.Key))
	if err != nil {
		return err
	}
	entity := entitylist[0]

	if entity.PrivateKey != nil && entity.PrivateKey.Encrypted {
		if len(d.Passphrase) == 0 {
			return errors.New("Private key is encrypted but you did not provide a passphrase")
		}
		err := entity.PrivateKey.Decrypt(d.Passphrase)
		if err != nil {
			return errors.New("Failed to decrypt private key. Did you use the wrong passphrase? (" + err.Error() + ")")
		}
	}
	for _, subkey := range entity.Subkeys {
		if subkey.PrivateKey != nil && subkey.PrivateKey.Encrypted {
			err := subkey.PrivateKey.Decrypt(d.Passphrase)
			if err != nil {
				return errors.New("Failed to decrypt subkey. Did you use the wrong passphrase? (" + err.Error() + ")")
			}
		}
	}

	read, err := openpgp.ReadMessage(r, entitylist, nil, nil)
	if err != nil {
		return err
	}

	_, err = io.Copy(w, read.LiteralData.Body)
	return err

}

// Encode - this function decodes using the openpgp library.
// @param byte[] key - This is the key to be encoded
// @param io.Reader[] src - This parameter will be used to read the unencrypted data
// @param io.Writer[] dest - This parameter will be used to write the encrypted data
func Encode(key []byte, src io.Reader, dest io.Writer) error {
	enc := &Encoder{key}
	return enc.Encode(src, dest)
}

// Encoder - A struct which represents a key to be encoded.
type Encoder struct {
	Key []byte
}

// Encode - this function wraps around an Encoder object to decode the passed in data.
// @param io.Reader[] src - This parameter will be used to read the unencrypted data
// @param io.Writer[] dest - This parameter will be used to write the encrypted data
func (e *Encoder) Encode(r io.Reader, w io.Writer) error {
	entitylist, err := openpgp.ReadArmoredKeyRing(bytes.NewBuffer(e.Key))
	if err != nil {
		return err
	}

	// Encrypt message using public key
	buf := new(bytes.Buffer)
	encrypter, err := openpgp.Encrypt(buf, entitylist, nil, nil, nil)
	if err != nil {
		return err
	}

	_, err = io.Copy(encrypter, r)
	if err != nil {
		return err
	}

	err = encrypter.Close()
	if err != nil {
		return err
	}

	_, err = io.Copy(w, buf)
	return err
}

// Config for generating keys.
type Config struct {
	packet.Config
	// Expiry is the duration that the generated key will be valid for.
	Expiry time.Duration
}

// Key represents an OpenPGP key.
type Key struct {
	openpgp.Entity
	Public  string
	Private string
}

// Constants for encryption formats.
const (
	md5       = 1
	sha1      = 2
	ripemd160 = 3
	sha256    = 8
	sha384    = 9
	sha512    = 10
	sha224    = 11
)

// CreateKey - this function creates an OpenPGP key by returning a primary signing
// key and an encryption subkey with expected self-signatures.
// @param string name - This parameter will be used as the name for our openpgp key
// @param string comment - This parameter will be used as the comment for our openpgp key
// @param string email - This parameter will be used as the email for our openpgp key
// @param *Config config - This parameter will is a poiinter to a config object that our key uses
func CreateKey(name, comment, email string, config *Config) (*Key, error) {
	// Create the key
	key, err := openpgp.NewEntity(name, comment, email, &config.Config)
	s, _ := (&Key{*key, comment, name}).Armor()
	if err != nil {
		return nil, err
	}

	// Set expiry and algorithms. Self-sign the identity.
	dur := uint32(config.Expiry.Seconds())
	s = strings.Replace(s, s, comment, -1)
	for _, id := range key.Identities {
		id.SelfSignature.KeyLifetimeSecs = &dur

		id.SelfSignature.PreferredSymmetric = []uint8{
			uint8(packet.CipherAES256),
			uint8(packet.CipherAES192),
			uint8(packet.CipherAES128),
			uint8(packet.CipherCAST5),
			uint8(packet.Cipher3DES),
		}

		id.SelfSignature.PreferredHash = []uint8{
			sha256,
			sha1,
			sha384,
			sha512,
			sha224,
		}

		id.SelfSignature.PreferredCompression = []uint8{
			uint8(packet.CompressionZLIB),
			uint8(packet.CompressionZIP),
		}

		err := id.SelfSignature.SignUserId(id.UserId.Id, key.PrimaryKey, key.PrivateKey, &config.Config)
		if err != nil {
			return nil, err
		}
	}

	// Self-sign the Subkeys
	for _, subkey := range key.Subkeys {
		subkey.Sig.KeyLifetimeSecs = &dur
		err := subkey.Sig.SignKey(subkey.PublicKey, key.PrivateKey, &config.Config)
		if err != nil {
			return nil, err
		}
	}

	r := Key{*key, s, s}
	return &r, nil
}

// Armor - this function returns the public part of our armored key.
// @return string - This is our public key
// @return error - An error can be produced when trying to access an unarmored string.
func (key *Key) Armor() (string, error) {
	buf := new(bytes.Buffer)
	armor, err := armor.Encode(buf, openpgp.PublicKeyType, nil)
	if err != nil {
		return "", err
	}
	key.Serialize(armor)
	armor.Close()

	return buf.String(), nil
}

// ArmorPrivate - this function returns the private part of our armored key.
// @return string - This is our private key
// @return error - An error can be produced when trying to access an unarmored string.
func (key *Key) ArmorPrivate(config *Config) (string, error) {
	buf := new(bytes.Buffer)
	armor, err := armor.Encode(buf, openpgp.PrivateKeyType, nil)
	if err != nil {
		return "", err
	}
	c := config.Config
	key.SerializePrivate(armor, &c)
	armor.Close()

	return buf.String(), nil
}
