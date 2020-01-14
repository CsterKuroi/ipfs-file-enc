package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"

	ipfsfe "github.com/Csterkuroi/ipfs-file-enc"
	senc "github.com/jbenet/go-simple-encrypt"
	mb "github.com/multiformats/go-multibase"
)

// flags
var (
	Key       string
	API       string
	RandomKey bool
)

// errors
var (
	ErrNoIPFS = errors.New("ipfs node error: not online")
)

const (
	gwayGlobal = "https://gateway.ipfs.io"
	gwayLocal  = "http://localhost:8080"
)

var Usage = `ENCRYPT AND SEND
    # will ask for a key
    ipfs-file-enc share <local-file-path>

    # encrypt with a known key. (256 bits please)
    ipfs-file-enc --key <secret-key> share <local-file-path>

    # encrypt with a randomly generated key. will be printed out.
    ipfs-file-enc --random-key share <local-file-path>


GET AND DECRYPT
    # will ask for key
    ipfs-file-enc download <ipfs-link> <local-destination-path>

    # decrypt with given key.
    ipfs-file-enc --key <secret-key> download <ipfs-link> <local-destination-path>

OPTIONS
    --h, --help              show usage
    --key <secret-key>       a 256bit secret key, encoded with multibase (no key = random key)
    --api <ipfs-api-url>     an ipfs node api to use (overrides defaults)

EXAMPLES
    > ipfs-file-enc share my_secret.jpg
`

func init() {
	flag.BoolVar(&RandomKey, "random-key", false, "use a randomly generated key (deprecated opt)")
	flag.StringVar(&Key, "key", "", "an AES encryption key in hex")
	flag.StringVar(&API, "api", "", "override IPFS node API")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, Usage)
	}
}

func decodeKey(k string) ([]byte, error) {
	_, b, err := mb.Decode(k)
	if err != nil {
		return nil, fmt.Errorf("multibase decoding error: %v", err)
	}
	if len(b) != 32 {
		return nil, fmt.Errorf("key must be exactly 256 bits. Was: %d", len(b))
	}
	return b, nil
}

func getSencKey(randomIfNone bool) (ipfsfe.Key, error) {
	NilKey := ipfsfe.Key(nil)

	var k []byte
	var err error
	if Key != "" {
		k, err = decodeKey(Key)
	} else if randomIfNone { // random key
		k, err = senc.RandomKey()
	} else {
		err = errors.New("Please enter a key with --key")
	}
	if err != nil {
		return NilKey, err
	}

	return ipfsfe.Key(k), nil
}

func cmdDownload(args []string) error {
	if RandomKey {
		return errors.New("cannot use --random-key with download")
	}
	if len(args) < 2 {
		return errors.New("not enough arguments. download requires 2. see -h")
	}

	srcLink := ipfsfe.IPFSLink(args[0])
	if len(srcLink) < 1 {
		return errors.New("invalid ipfs-link")
	}

	dstPath := args[1]
	if dstPath == "" {
		return errors.New("requires a destination path")
	}

	// check for Key, get key.
	key, err := getSencKey(false)
	if err != nil {
		return err
	}

	// fmt.Println("Initializing ipfs node...")
	n := ipfsfe.GetROIPFSNode(API)
	if !n.IsUp() {
		return ErrNoIPFS
	}

	// fmt.Println("Getting", srcLink, "...")
	err = ipfsfe.GetDecrypt(n, srcLink, dstPath, key)
	if err != nil {
		return err
	}
	fmt.Println("write to:", dstPath)
	return nil
}

func cmdShare(args []string) error {
	if len(args) < 1 {
		return errors.New("not enough arguments. share requires 1. see -h")
	}
	srcPath := args[0]
	if srcPath == "" {
		return errors.New("requires a source path")
	}

	// check for Key, get key.
	key, err := getSencKey(true)
	if err != nil {
		return err
	}

	// fmt.Println("Initializing ipfs node...")
	n, err := ipfsfe.GetRWIPFSNode(API)
	if err != nil {
		return err
	}
	if !n.IsUp() {
		return ErrNoIPFS
	}

	// fmt.Println("Sharing", srcPath, "...")
	link, err := ipfsfe.EncryptAndPut(n, srcPath, key)
	if err != nil {
		return err
	}

	l := string(link)
	if !strings.HasPrefix(l, "/ipfs/") {
		l = "/ipfs/" + l
	}

	keyStr, err := mb.Encode(mb.Base58BTC, key)
	if err != nil {
		return err
	}

	fmt.Println("Shared as: ", l)
	fmt.Println("Key: ", keyStr)
	fmt.Println("Ciphertext on global gateway: ", gwayGlobal, l)
	fmt.Println("Ciphertext on local gateway: ", gwayLocal, l)
	fmt.Println("")
	fmt.Println("Get, Decrypt with:")
	fmt.Println("    ipfs-file-enc --key", keyStr, "download", l, "<filename>")
	fmt.Println("")
	return nil
}

func errMain(args []string) error {
	// no command is not an error. it's usage.
	if len(args) == 0 {
		fmt.Println(Usage)
		return nil
	}

	cmd := args[0]
	switch cmd {
	case "download":
		return cmdDownload(args[1:])
	case "share":
		return cmdShare(args[1:])
	default:
		return errors.New("Unknown command: " + cmd)
	}
}

func main() {
	flag.Parse()
	args := flag.Args()
	if err := errMain(args); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(-1)
	}
}
