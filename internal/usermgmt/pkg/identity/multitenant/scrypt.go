package multitenant

import (
	"crypto/aes"
	"crypto/cipher"
	"github.com/pkg/errors"
	"golang.org/x/crypto/scrypt"
)

//Key derives a key from the password, salt, and cost parameters, returning a byte slice of length keyLen that can
//be used as cryptographic key.
//N is a CPU/memory cost parameter, which must be a power of two greater than 1. r and p must satisfy r * p < 2³⁰. If the
//parameters do not satisfy the limits, the function returns a nil byte slice and an error.

//The recommended parameters for interactive logins as of 2017 are N=32768, r=8 and p=1. The parameters N, r, and p should
//be increased as memory latency and CPU parallelism increases; consider setting N to the highest power of 2 you can derive
//within 100 milliseconds. Remember to get a good random salt.
const (
	P      = 1
	KeyLen = 32
)

var (
	ErrNilScryptHash       = errors.New("scrypt hash is nil")
	ErrInvalidScryptKey    = errors.New("scrypt signer key not specified")
	ErrInvalidScryptRounds = errors.New("scrypt rounds must be between 1 and 8")
	ErrInvalidMemoryCost   = errors.New("scrypt memory cost must be between 1 and 14")
)

type ScryptHash interface {
	Key() []byte
	SaltSeparator() []byte
	Rounds() int
	MemoryCost() int
}

func IsScryptHashValid(scryptHash ScryptHash) error {
	if scryptHash == nil {
		return ErrNilScryptHash
	}
	if len(scryptHash.Key()) == 0 {
		return ErrInvalidScryptKey
	}
	if scryptHash.Rounds() < 1 || scryptHash.Rounds() > 8 {
		return ErrInvalidScryptRounds
	}
	if scryptHash.MemoryCost() < 1 || scryptHash.MemoryCost() > 14 {
		return ErrInvalidMemoryCost
	}
	return nil
}

func HashedPassword(hash ScryptHash, password []byte, salt []byte) ([]byte, error) {
	err := IsScryptHashValid(hash)
	if err != nil {
		return nil, err
	}

	ck, err := scrypt.Key(password, append(salt, hash.SaltSeparator()...), 1<<hash.MemoryCost(), hash.Rounds(), P, KeyLen)
	if err != nil {
		return nil, err
	}

	var block cipher.Block
	if block, err = aes.NewCipher(ck); err != nil {
		return nil, err
	}

	cipherText := make([]byte, aes.BlockSize+len(hash.Key()))
	stream := cipher.NewCTR(block, cipherText[:aes.BlockSize])
	stream.XORKeyStream(cipherText[aes.BlockSize:], hash.Key())
	return cipherText[aes.BlockSize:], nil
}

//experiment
/*type App struct {
	SignerKey     []byte
	SaltSeparator []byte
	Rounds        int
	MemCost       int
	P             int
	KeyLen        int
}

func New(signerKey, saltSeparator string, rounds, memCost int) (*App, error) {
	var (
		app = &App{
			Rounds:  rounds,
			MemCost: memCost,
			P:       P,
			KeyLen:  KeyLen,
		}
		err error
	)

	if app.SignerKey, err = base64.StdEncoding.DecodeString(signerKey); err != nil {
		return nil, err
	}
	if app.SaltSeparator, err = base64.StdEncoding.DecodeString(saltSeparator); err != nil {
		return nil, err
	}
	return app, nil
}

func (a *App) EncodeToString(password, salt []byte) (string, error) {
	res, err := a.Encode(password, salt)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(res), nil
}

func (a *App) Encode(password, salt []byte) ([]byte, error) {
	return key(password, salt, a.SignerKey, a.SaltSeparator, a.Rounds, a.MemCost, a.P, a.KeyLen)
}

func (a *App) Verify(password, salt []byte, passwordHash string) bool {
	h, err := a.EncodeToString(password, salt)
	if err != nil {
		return false
	}

	return h == passwordHash
}

func (a *App) FirebaseVerify(password, salt, passwordHash string) bool {
	_salt, err := base64.StdEncoding.DecodeString(salt)
	if err != nil {
		return false
	}

	var hs string
	if hs, err = a.EncodeToString([]byte(password), _salt); err != nil {
		return false
	}
	return hs == passwordHash
}

func Key(password, salt []byte, signerKey, saltSeparator string, rounds, memCost, p, keyLen int) ([]byte, error) {
	var (
		sk, ss []byte
		err    error
	)

	if sk, err = base64.StdEncoding.DecodeString(signerKey); err != nil {
		return nil, err
	}
	if ss, err = base64.StdEncoding.DecodeString(saltSeparator); err != nil {
		return nil, err
	}

	return key(password, salt, sk, ss, rounds, memCost, p, keyLen)
}

func key(password, salt, signerKey, saltSeparator []byte, rounds, memCost, p, keyLen int) ([]byte, error) {
	ck, err := scrypt.Key(password, append(salt, saltSeparator...), 1<<memCost, rounds, p, keyLen)
	if err != nil {
		return nil, err
	}

	var block cipher.Block
	if block, err = aes.NewCipher(ck); err != nil {
		return nil, err
	}

	cipherText := make([]byte, aes.BlockSize+len(signerKey))
	stream := cipher.NewCTR(block, cipherText[:aes.BlockSize])
	stream.XORKeyStream(cipherText[aes.BlockSize:], signerKey)
	return cipherText[aes.BlockSize:], nil
}*/
