package ipfssenc

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	ipfs "github.com/ipfs/go-ipfs-api"
	senc "github.com/jbenet/go-simple-encrypt"
)

type IPFSLink string
type Key []byte

var (
	ErrNotImplemented       = errors.New("ErrNotImplemented")
	ErrFailedToUseLocalNode = errors.New("Failed to use local node")
)

var (
	GlobalNodeURL    = "https://gateway.ipfs.io:4001"
	GlobalGatewayURL = "https://gateway.ipfs.io"
	LocalNodeURL     = "http://localhost:4001"
	LocalGatewayURL  = "http://localhost:8080"
)

func GetRWIPFSNode(url string) (*ipfs.Shell, error) {
	// must be a local node.
	if url != "" {
		return ipfs.NewShell(url), nil
	}

	// no url given. try local
	if n := ipfs.NewLocalShell(); n != nil {
		return n, nil
	}

	// try LocalGatewayURL
	if n := ipfs.NewShell(LocalNodeURL); n != nil {
		return n, nil
	}

	return nil, ErrFailedToUseLocalNode
}

func GetROIPFSNode(url string) *ipfs.Shell {
	rwn, err := GetRWIPFSNode(url)
	if err == nil {
		return rwn
	}

	// use global gateway.
	return ipfs.NewShell(GlobalNodeURL)
}

//
// Encrypt encrypts a given PlainText w/ given Key.
func Encrypt(pt io.Reader, secret Key) (ct io.Reader, err error) {
	return senc.Encrypt(secret, pt)
}

// Decrypt decrypts a given CipherText w/ given Key.
func Decrypt(ct io.Reader, secret Key) (pt io.Reader, err error) {
	return senc.Decrypt(secret, ct)
}

// Put adds a given Reader to the network, and gets a link for it.
func Put(n *ipfs.Shell, r io.Reader) (IPFSLink, error) {
	s, err := n.Add(r)
	if s != "" {
		s = "/ipfs/" + s
	}
	return IPFSLink(s), err
}

// Get retrieves a CipherText for given Link from the network.
func Get(n *ipfs.Shell, link IPFSLink) (io.ReadCloser, error) {
	return n.Cat(string(link))
}

func EncryptAndPut(n *ipfs.Shell, localPath string, secret Key) (IPFSLink, error) {
	if _, err := os.Stat(localPath); err != nil {
		return IPFSLink(""), fmt.Errorf("Unable to open files - %v", err.Error())
	}
	srcIsDir := false
	if fi, err := os.Stat(localPath); err != nil {
		return IPFSLink("") ,err
	} else {
		srcIsDir = fi.IsDir()
	}
	if srcIsDir{
		return IPFSLink("") ,fmt.Errorf("Unable to open dir")
	}
	f, err := os.Open(localPath)
	defer f.Close()
	c, err := Encrypt(f, secret)
	if err != nil {
		return IPFSLink(""), err
	}
	return Put(n, c)
}

func GetDecrypt(n *ipfs.Shell, link IPFSLink, localPath string, secret Key) error {
	ct, err := Get(n, link)
	if err != nil {
		return err
	}
	defer func() {
		// ipfs-api.Cat docs say to drain the reader and close it.
		io.Copy(ioutil.Discard, ct)
		ct.Close()
	}()

	pt, err := Decrypt(ct, secret)
	if err != nil {
		return err
	}
	f, err := os.OpenFile(localPath, os.O_CREATE|os.O_RDWR, os.FileMode(0644))
	// copy over contents
	if _, err := io.Copy(f, pt); err != nil {
		return err
	}
	return nil
}
