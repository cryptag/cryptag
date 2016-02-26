# CrypTag: Encrypted, taggable, searchable cloud storage

## Vision

My goal is both to enable all internet users to access all their data
from any of their devices without trusting the party storing it, and
to enable them to conveniently, selectively share this data with
others.


## What is CrypTag?

CrypTag is free, open source software that encrypts your data, _and_
makes it searchable by tag!  This Sandstorm package lets you store
your own data and files on Sandstorm without having to trust the party
who has access to these files, since all data is encrypted.

And even if you self-host Sandstorm on your own server, your data will
still be safe even if this server is stolen.

To pull this off, all data is encrypted and decrypted on the client.


## CrypTag Client Installation

### cpass

To install a command line client, _cpass_ (short for CryptPass),
useful for storing and retrieving the following kinds of data on your
Sandstorm-hosted CrypTag instance...

1. Passwords (**cpass @elimisteve**)
2. Credit card numbers (**cpass visa digits**)
3. Quotes (**cpass nietzsche quote**)
4. Bookmarks, tagged like on Pinboard or Delicious (**cpass url snowden**)
5. Command line commands -- cross-machine shell history! (**cpass install docker**)
6. GitHub recovery codes (**cpass github recoverycode**)

...install Go (see [instructions](https://golang.org/doc/install)), then run

```
$ go get github.com/elimisteve/cryptag/cmd/cryptpass-webserver-client
$ cryptpass-webserver-client
```


### cget, cput

Interested in command line clients that let you save and retrieve
files?  Contact me (@elimisteve on Twitter,
steve@tryingtobeawesome.com via email) so I can better understand your
use case(s), and I'll happily build you one based on the [existing
client code](https://github.com/elimisteve/cryptag/tree/master/cmd)!


## More Details

For more technical details and use cases, see
<https://github.com/elimisteve/cryptag#cryptag>.
