# Go-miniLock
### A pure-Go reimplementation of the miniLock asymmetric encryption system.
by Cathal Garvey, Copyright Oct. 2015, proudly licensed under the GNU AGPL.

[![Support via Gratipay](https://cdn.rawgit.com/gratipay/gratipay-badge/2.3.0/dist/gratipay.png)](https://gratipay.com/~onetruecathal/)

[Or, Tip me a few bits? - 32ddsuR73CHH8igCNCLvRE3UwBqL8yU2ag](bitcoin:32ddsuR73CHH8igCNCLvRE3UwBqL8yU2ag?label=Support%20go-miniLock%20and%20other%20projects%20that%20matter)

### What
TL;DR: go-miniLock is a total Golang rewrite of miniLock, enabling native code performance,
more platform flexibility, and downstream potential for automation and novel communication
media not available to the original miniLock Chromium app.

See [miniLock.io](https://miniLock.io) for information on miniLock. It's a file
encryption system designed by [Nadim Kobaissi](https://nadim.computer/) and reviewed
for security and soundness by experts. It's pretty well-put together, but as if that
weren't enough it was released as an easy-to-use, user-focused Chrome App. In many
respects it achieves what PGP was supposed to achieve, while dodging all the nastiness
of PGP: Minilock gives:

* Tiny identities: At ~40 characters, miniLock ID keys can be shared trivially through
  any medium.
* Authenticated, Private communication between people without requiring a third-party
* Easy verification of respective key:identity matchings.
* Minimised metadata within the encrypted files; only recipients can see the identity
  of the sender and themselves, they cannot identify other valid recipients, and outsiders
  cannot determine, given a miniLock file, who sent it or who was the recipient.
* No persistent keys: miniLock is designed to use deterministic keys that are generated
  from the user's memorable, highly secure passphrase and their email address (or a fake one..)
* Transport agnostic: miniLock just encrypts files, it doesn't insist on a particular way
  of transmitting them.

The big disadvantage of miniLock has been how tied it is to Chrome; this limits platforms
to laptops and desktops only, to GUI-enabled systems only, and makes integrating miniLock
into other systems impossible. If you wanted to create a logging application that sends
encrypted reports to your email daily using miniLock, forget it. If you wanted to build a
P2P social network using miniLock for authentication and privacy, forget it.

See [deadLock](https://github.com/cathalgarvey/deadlock) for my past efforts towards creating
a shell-scriptable version of miniLock, but Python isn't much better than a Chrome app, due
to version wars (thanks, fossilised 2.X users..), lack of pre-installation on Windows, and
difficulty of C-extension compilation on WinMac. Oh, and the bug-prone-ness of Python in
general!

Golang, as a language, addresses all the needs I forsee for a more versatile miniLock:
it compiles to any platform extremely quickly, it has growing support for building native
mobile apps, it can transpile to JS, it's fast, and it offers useful tools and concurrency
primitives that facilitate the underlying, highly paralleliseable activity of miniLock.
And, for a developer, it's very good at catching common bugs at compile-time;
forgotten or renamed variables, typing errors, mismatched return types, etcetera.

So here's go-minilock; it sets out to be both an easy-to-use alternative to PGP, a native
answer to the miniLock browser extension, and a library for easily constructing tools that
go beyond manual human-to-human cryptography and extend into the automated, networked, or
decentralised sphere.

### Usage
Documentation for the library can be found at [godoc](https://godoc.org/github.com/cathalgarvey/go-minilock).
Functionality is deliberately broken into construction of the encrypted data itself and constructing
the headers that assist in decryption and obfuscation of communicating parties; this is to enable
use of the library for more than just miniLock-of-files, but also because other systems built atop
miniLock (such as [Peerio](https://peerio.com)) use detached, updateable headers as a way to
facilitate social file-sharing.

Much of the slightly-lower-level crypto stuff is in a sub-package called "taber",
which can be imported separately with `import "github.com/cathalgarvey/go-minilock/taber"`,
and documentation for which is [here on Godoc](https://godoc.org/github.com/cathalgarvey/go-minilock/taber).

For terminal usage of go-miniLock, you can install the tool with: `go get -u github.com/cathalgarvey/go-minilock/minilock-cli`.
Usage is simple enough and needs improvement:

    minilock-cli encrypt <file> <your email> <recipient1> [<recipient2>...]
    minilock-cli decrypt <file> <your email>

A number of flags modify usual behaviour. The most important is probably the "-p"
flag which allows the passphrase for the user's key to be provided directly instead
of being requested interactively; this allows shell-scripting using minilock-cli,
or simply aliasing to create a rapid way of encrypting or decrypting things using
your key. Beware, obviously, that for *personal* uses this breaks one of the security
features of minilock, namely that personal keys are not stored but *remembered*!
This feature, therefore, was more intended for server-side or scripting uses than
for individuals.

A UI would be *really* nice but isn't yet on the cards. Watch this space. Meanwhile, use [miniLock](https://minilock.io).

### Where from Here
Here are things I'd really enjoy, if you're feeling creative. I may start on some of these, also..

* Python bindings to go-miniLock, to enable a drastic refactor of [deadlock](https://github.com/cathalgarvey/deadlock).
  Current Go:Python binding solutions I've seen have involved some very ugly C shimming, but I suspect
  using FFI or Ctypes might work since Go 1.5 introduced C-ABI library compilation?
* Integration of go-miniLock with desktop mail clients.
* Transpiling *usefully* to JS using GopherJS, with a comparable library interface.
* An Android client using the new Go:Android Bindings introduced in Go 1.5. Integration of said Android client into K9 Mail.
* A self-hostable, federating [Peerio](https://peerio.com) server that
  [respects your fundamental rights](https://fsf.org).
    - Bonus: Federates with other such servers in a robust way.
    - Bonus: Generates chaff traffic.
    - Bonus: Offers option to delete correspondance in same way as shared files.
    - Bonus: Talks to email servers, receives email and stores/delivers miniLock..and vice-versa.
* A total rewrite of Peerio Client that doesn't require Chromium, could run headlessly.
    - Bonus: IMAP/SMTP adaptor 'client' for mail client alternative.
    - Bonus: IRC/XMPP adaptor 'client' for chat client alternatives.
    - Bonus: "Sync Folder Contents" option for dropbox-style crypto-extension to Peerio.

### Credits Reel
* [GNU Foundation](https://gnu.org) for Free Software and the GPL/AGPL. :)
* [Nadim Kobeissi](https://nadim.computer) for miniLock itself.
* [dchest](https://github.com/dchest) for [blake2s](https://github.com/dchest/blake2s), necessary for file hashing in miniLock.
* [alecthomas](https://github.com/alecthomas) for [kingpin](https://github.com/alecthomas/kingpin), used for CLI flag parsing.
* [howeyc](https://github.com/howeyc) for [gopass](https://github.com/howeyc/gopass), used for secure interactive password input.
