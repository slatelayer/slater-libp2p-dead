package core

import (
	"crypto/ed25519"
	"encoding/hex"
	"errors"
	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/blake2b"
	"golang.org/x/crypto/hkdf"
	"hash"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/kscarlett/humid"
	"github.com/kscarlett/humid/wordlist"
	"github.com/sethvargo/go-diceware/diceware"
	"lukechampine.com/frand"
)

//
// Slater uses Diceware for passphrase generation.
// https://theworld.com/~reinhold/diceware.html
//
// For now, it uses the EFF long list. (Considering short list #2 with autocomplete...)
// https://www.eff.org/deeplinks/2016/07/new-wordlists-random-passphrases
//
// TODO include the other avilable lists.
// The availability of these lists will define the initial localization effort:
// if there is already a diceware list for a language, then localize for it.
//

// Rationalle and assumptions:
//
// Approximate entropy of each diceware word: log2(7776) = 12.9 bits.
// Approximate entropy of each pin digit: log2(10) = 3.3 bits.
//
// Using a six-word phrase + four-digit PIN, we get:
//
//		(6 * math.log2(7776)) + (4 * math.log2(10)) = 90.8
//
// And then for the master key, we stretch the combo with Argon2,
// and for the discovery key, we hash the combo with blake2b.
//
// Please holler if you disagree or find this unsuitable for your use case.
//

const (
	WORDS            = 6
	DIGITS           = 4
	DISCOVERY_PREFIX = "slater"
	SALT             = "salt"
	HASH             = "hash"

	ARGON_TIME   = 1
	ARGON_MEM    = 64 * 1024
	ARGON_KEYLEN = 32
)

func generateSessionName() string {
	return humid.GenerateWithOptions(&humid.Options{
		List:           wordlist.Animals,
		AdjectiveCount: 2,
		Separator:      "-",
		Capitalize:     false,
	})
}

func generatePassphrase() string {
	words, err := diceware.Generate(WORDS)
	if err != nil {
		log.Panic("could not generate passphrase!\n", err)
	}
	return strings.Join(words, " ")
}

func generatePin() (pin string) {
	for range [DIGITS]int{} {
		pin = pin + strconv.Itoa(frand.Intn(10))
	}
	return
}

func discoveryKey(parts ...string) string {
	s := strings.Join(append([]string{DISCOVERY_PREFIX}, parts...), "-")
	b := []byte(s)
	h := blake2b.Sum256(b)
	return hex.EncodeToString(h[:])
}

func createMasterKey(rootPath, sessionID, phrase, pin string) string {
	salt, _ := createSalt(rootPath, sessionID)
	k := stretch(salt, sessionID, phrase, pin)
	hash := blake2b.Sum256(k)
	s := hex.EncodeToString(hash[:])
	path := filepath.Join(rootPath, sessionID, HASH)
	ioutil.WriteFile(path, []byte(s), 0600)
	return string(k)
}

func getMasterKey(rootPath, sessionID, phrase, pin string) (string, error) {
	salt, err := getSalt(rootPath, sessionID)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) { // TODO any other case to consider?
			//☠️ Oh, look: you deleted your salt file.
			//☠️ Dear user, you have destroyed all of your data.
			//☠️ I hope you have a full copy of it on at least one other device.
			//☠️ That sucks. Please don't do that again.
			return "", errLostSalt
		} else {
			return "", err
		}
	}
	k := stretch(salt, sessionID, phrase, pin)
	hash := blake2b.Sum256(k)
	path := filepath.Join(rootPath, sessionID, HASH)
	s := hex.EncodeToString(hash[:])
	savedHash, err := ioutil.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) { // TODO any other case to consider?
			//☠️ Oh, look: you deleted your hash file.
			//☠️ Now we can't use it to check your creds before trying to open the database.
			//☠️ So I hope you have at least one other device online,
			//☠️ so we can try to connect to it first, before opening the db and then effing reconnecting again.
			//☠️ I should just crash without saying anything!
			//☠️ Don't mess around with my files, or bad stuff may happen to your data.
			return "", errLostHash
		} else {
			return "", err
		}
	}
	if s != string(savedHash) {
		return "", errAuthFail
	}
	return string(k), nil
}

func deriveSignatureKey(sessionID, phrase, pin string) (ed25519.PrivateKey, error) {
	seed := make([]byte, ed25519.SeedSize)
	secret := []byte(sessionID + phrase + pin)
	info := []byte("yeah, slater!")
	hashFunc := func() hash.Hash { h, _ := blake2b.New256(nil); return h }
	// I guess we could pass a salt along in the message instead of omitting it...
	hkdf := hkdf.New(hashFunc, secret, nil, info)
	if _, err := io.ReadFull(hkdf, seed); err != nil {
		return nil, err
	}
	key := ed25519.NewKeyFromSeed(seed)
	return key, nil
}

var errLostSalt = errors.New("☠️ salt file missing")
var errLostHash = errors.New("☠️ hash file missing")
var errAuthFail = errors.New("passphrase and pin verification failed")

func stretch(salt []byte, things ...string) []byte {
	threads := uint8(runtime.NumCPU())
	s := strings.Join(things, "")
	return argon2.IDKey([]byte(s), salt, ARGON_TIME, ARGON_MEM, threads, ARGON_KEYLEN)
}

func createSalt(rootPath, sessionID string) (salt []byte, err error) {
	dpath := filepath.Join(rootPath, sessionID)
	fpath := filepath.Join(dpath, SALT)
	if err = os.MkdirAll(dpath, 0700); err != nil {
		return nil, err
	}
	salt = frand.Bytes(16)
	err = ioutil.WriteFile(fpath, salt, 0600)
	if err != nil {
		return nil, err
	}
	return
}

func getSalt(rootPath, sessionID string) (salt []byte, err error) {
	path := filepath.Join(rootPath, sessionID, SALT)
	salt, err = ioutil.ReadFile(path)
	return
}
