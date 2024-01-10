# webpki-issue-223

https://github.com/rustls/webpki/pull/223

## Run

Go 1.17 is required, you can get the version using [gvm](https://github.com/moovweb/gvm): `gvm install go1.17.13` and `gvm use go1.17.13`

```bash
# run go server first
$ cd go_quic_server
$ go run main.go
```

```bash
# run rust quinn client in another terminal
$ cd quinn_client
$ cargo run
```

You should see:

```
ERROR quinn_client: connecting: the cryptographic handshake failed: error 42: invalid peer certificate: BadEncoding
```

# Credits

The `quinn_client` copies and modifies code from the `quinn` project, which is an MIT and Apache-2 dual-licensed project.

This is only a poc project of an `rustls-webpki` issue.

# License

MIT or Apache-2
