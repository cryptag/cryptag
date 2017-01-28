# Taber - NaCl at Higher Elevation
by Cathal Garvey, Copyright Oct 2015, Proudly [AGPL licensed](https://gnu.org/licenses/agpl.html).

### How do I use this?
Get thee to [Taber's Godoc](https://godoc.org/github.com/cathalgarvey/go-minilock/taber).

### What is this?
This is the technical heavy lifting behind encryption in [go-minilock](), separated from the main repository to keep things tidy and make it easier to re-use for other stuff.

[miniLock](https://minilock.io), a high-level file encryption system, offers an extremely user-friendly way to encrypt and decrypt things, designed for normal people, using NaCl along with some other great pieces:

* [zxcvbn](), a passphrase-strength checking library that ensures safe passwords and forbids the stupid ones we're all prone to using.
* [blake2s](), a fast and variable length cryptographic hash function.
* [scrypt](), a password hardening function that helps combat computers' natural advantage in brute-force-guessing passwords.
* [Networking and Cryptography Library]() ("NaCl"), a high-level cryptographic library designed for application developers to make crypto harder to screw up.

With the exception of zxcvbn, which hasn't yet been ported to Go, this library offers the constructions used by miniLock in pure Go, built using Blake2s, Scrypt and NaCl.

### What's in a name?
Taber-NaCl. You put your precious stuff in it. :)
