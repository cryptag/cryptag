## CrypTag

**_Encrypted, taggable, searchable web storage._**

CrypTag was announced at DEF CON 23 in August of 2015.  Presentation
slides:
<https://talks.stevendphillips.com/cryptag-defcon23-cryptovillage>.


### What is CrypTag?

CrypTag is an idea, a spec, an API, and a piece of software that makes
it easy to build a zero-knowledge system, which means that the server
holding user data doesn't know what it is (since it's encrypted).

It is meant as a primitive to be used to build more sophisticated
systems that would rather not re-implement the pieces necessary to
build a zero-knowledge system.

To use a command line password manager, CryptPass, see "Getting
Started with CryptPass" below.


### How is it searchable _and_ encrypted?

It's not _fully_ searchable; you can query by tag.  See the
presentation from DEF CON 23:
<https://talks.stevendphillips.com/cryptag-defcon23-cryptovillage/#/7>


### Then the server stores the tags in plaintext?

Nope!  The client stores mapping between tags ("snowden") and a random
hex string ("b6a27d9"), and the server only ever sees the random
strings.

(The client also encrypts these mappings and stores them to the
server, too.)


### Use Cases (what CrypTag is good at)

Anything in the "write once, read many times" category is ideal, such
as:

* Bookmarks
* Notes/personal writings
* Emails, text messages, instant messages


### Non-use Cases (what CrypTag is not good at)

Anything that requires rapid changes being made to data by multiple
users, such as:

* Real-time collaborative document editing

  * Real-time spreadsheet editing should work OK (as long as you're OK
    with "last write wins" to a cell), since each cell can be its own
    Row that can be changed concurrently with other Rows


## Getting Started with CryptPass

The current focus for CrypTag is creating a password manager out of it
called _CryptPass_.  CryptPass exists as a command line tool (`cpass`)
you can use to store and retrieve passwords.  Unencrypted passwords
_never_ touch disk; they are stored encrypted, read into memory, then
printed to your terminal for you to use, with the first one found
added to your clipboard.

### TL;DR version

Install + config:

```
go get github.com/elimisteve/cryptag/cmd/cpass
cpass
```

Create passwords, fetch them by tag, or delete them:

```
cpass create mytwitterp4ssw0rd twitter @myusername tag3 tag4
cpass @myusername
cpass delete twitter
```

Keep reading for more advanced options, including password sharing via
shared Dropbox folders.


### Installing cpass

Install Go ([instructions](https://golang.org/doc/install)), then run

    go get github.com/elimisteve/cryptag/cmd/cpass

That's it!  Now run

    cpass

`cpass` will generate a new encryption key to store your passwords
with, as well as create the directories it will use to store your
data, all in `~/.cryptag` (by default).


### Using cpass

**Create** a new password and associated tags with commands like:

    cpass create mycr4zyemailp4ssw0rd gmail email elimisteve@gmail

    cpass create mytwitterp4ssw0rd twitter @elimisteve

    cpass create mycr4zyAWSp4ssw0rd4myj0b work aws


**Fetch** the password you're looking for and see them printed to the
screen with commands like:

    cpass gmail

    cpass @elimisteve

    cpass aws work

For convenience, `cpass` adds the first password found to your
clipboard so you can paste it into whichever program you're using.

To **view all** your passwords, run

    cpass all

And finally, to **delete** all passwords with certain tags, run

    cpass delete aws

To only delete one specific password, not all passwords with a generic
tag (e.g., "email") that you may have used to tag multiple passwords,
use the password's tag of the form `id:...`, which is auto-generated
and guaranteed to be unique:

    cpass delete id:a91d46c7-45bb-48e4-43d1-642196df15b2


### Multiple Storage Backends

Maybe you want to store your personal data in `~/.cryptag` but have
passwords you share with colleagues at `~/Dropbox/cryptag_work`, for
example.  (With `cpass` this is secure because plaintext, unencrypted
passwords _never_ touch disk; only encrypted data will ever be stored
in `~/Dropbox/cryptag_work`.)

Create a new backend with the desired name (e.g., "work") by running

    CRYPTAG_BACKEND_NAME=work cpass

You can then change the value of `DataPath` at the end of
`~/.cryptag/backends/work.json` to wherever you want your work
passwords stored (e.g., `/home/MYUSERNAME/Dropbox/cryptag_work`),
being sure to use the absolute path.

(Dropbox note: CrypTag-based programs generally, and `cpass`
specifically, store each piece of data (e.g., each password and each
tag) in a separate file, so it _is_ safe for multiple people to create
passwords simultaneously and sove them to a shared Dropbox folder,
unlike with KeePass, KeePassX, 1Password, and some other password
managers.)

Now you can save shared work passwords with the same commands as
before, except with the `CRYPTAG_BACKEND_NAME` environment variable
set:

    CRYPTAG_BACKEND_NAME=work cpass create mycr4zyAWSp4ssw0rd4myj0b work aws

Now you should share `~/.cryptag/backends/work.json` with your
colleagues -- or at least the encryption key -- so that you can
decrypt passwords saved by each other.


### More Convenient Multiple Storage Backends

See [this issue](https://github.com/elimisteve/cryptag/issues/18) for
discussion on how to make this much better!  I currently do this:

    echo 'CRYPTAG_BACKEND_NAME=work cpass "$@"' > ~/bin/work
    chmod +x ~/bin/work

so I can simply do

    work create mycr4zyAWSp4ssw0rd4myj0b aws

to create work passwords, or

    work aws

to fetch them.
