# ipfs-file-enc - encrypt file and add to IPFS.

Currently, IPFS does not have an inbuilt content encryption system. Many solutions exist on top. I wanted something easy. This builds on [senc](https://github.com/jbenet/go-simple-encrypt).

## On the commandline

This tool is command-line based.

### Install

Using go get:

```
go get github.com/Csterkuroi/ipfs-file-enc/ipfs-file-enc
```

### How to encrypt & share

```
# encrypt with a known key. (256bits)
ipfs-file-enc --key <secret-key> share <path-to-file>

# encrypt with a randomly generated key. will be printed out.
ipfs-file-enc share <path-to-file>
```

Leave your IPFS node running, or pin this somewhere.

### How to download & decrypt

```
# will ask for key
ipfs-senc download <ipfs-link> <local-destination-dir>

# decrypt with given key.
ipfs-senc --key <secret-key> download <ipfs-link> [<local-destination-dir>]
```

Will use your local ipfs node, or the ipfs-gateway if no local node is available.

## Example
```
>ipfs-file-enc share index.js
Shared as:  /ipfs/Qmanj686x6XgSb685iDmBpenfVFb396ymbEMAegBAp3e7a
Key:  zEKiSDNK8vcL5qew7kH2VznEgQaQSU3vJoEnLjXXKkf8j
Ciphertext on global gateway:  https://gateway.ipfs.io /ipfs/Qmanj686x6XgSb685iDmBpenfVFb396ymbEMAegBAp3e7a
Ciphertext on local gateway:  http://localhost:8080 /ipfs/Qmanj686x6XgSb685iDmBpenfVFb396ymbEMAegBAp3e7a

Get, Decrypt with:
    ipfs-file-enc --key zEKiSDNK8vcL5qew7kH2VznEgQaQSU3vJoEnLjXXKkf8j download /ipfs/Qmanj686x6XgSb685iDmBpenfVFb396ymbEMAegBAp3e7a <filename>

>ipfs-file-enc --key zEKiSDNK8vcL5qew7kH2VznEgQaQSU3vJoEnLjXXKkf8j download /ipfs/Qmanj686x6XgSb685iDmBpenfVFb396ymbEMAegBAp3e7a new.js
write to: new.js
```

## TODO or problem

1. cipher in memory
2. need to save file_name, file_mode, key
3. 16 bytes IV

## License

MIT, copyright Csterkuroi.
