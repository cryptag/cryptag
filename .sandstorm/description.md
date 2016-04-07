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


## CrypTag Grains

You can use CrypTag Sandstorm grains for storing passwords or other
secrets. The three best things about CrypTag:

- You access all your data from a secure client on your own computer.

- Data never goes out of sync. (All data is stored on the server.)

- Searches are efficient. This is CrypTag's key idea: store secret information on a server,
  with labels ("tags") that the server can't understand, that can still be used for search!


Overview of daily use
---------------------

    $ cpass-sandstorm create mytw1tt3rp4ssword twitter @myusername login:myusername

This stores your Twitter password (in encrypted form, of course) to Sandstorm.


    $ cpass-sandstorm @myusername

This adds the Twitter password for @myusername to your clipboard automatically!


    $ cpass-sandstorm all

This will list all passwords (and, actually, all other textual data;
see below) you've stored.


More tips/use cases
-------------------

cpass-sandstorm is for more than just passwords, though.  You may also
want to store and access:

1. Credit card numbers (cpass-sandstorm visa digits)
2. Quotes (cpass-sandstorm nietzsche quote)
3. Bookmarks, tagged like on Pinboard or Delicious (cpass-sandstorm url snowden)
4. Command line commands -- cross-machine shell history! (cpass-sandstorm install docker)
5. GitHub recovery codes (cpass-sandstorm github recoverycode)


Getting started
---------------

## Linux and Mac OS X

Run this to download the cpass-sandstorm command line program:

    $ mkdir ~/bin; cd ~/bin && C="cpass-sandstorm" && curl -SL https://github.com/elimisteve/cryptag/blob/v1-beta/bin/cpass-sandstorm$(if [ "$(uname)" != "Linux" ]; then echo -n "-osx"; fi)?raw=true -o ./$C && chmod +x ./$C

Then click the key icon above this web page (on Sandstorm) and
generate a Sandstorm API key to give to cpass-sandstorm like so:

    $ ./cpass-sandstorm init <sandstorm_key>

To see the remaining valid commands (such as "create", seen above), run

    $ ./cpass-sandstorm

Enjoy!


## Windows

Run this in PowerShell:

    (New-Object Net.WebClient).DownloadFile("https://github.com/elimisteve/cryptag/blob/v1-beta/bin/cpass-sandstorm$(If ([IntPtr]::size -eq 4) { '-32' }).exe?raw=true", "cpass-sandstorm.exe"); icacls.exe .\cpass-sandstorm.exe /grant everyone:rx

Then click the key icon above this web page (on Sandstorm) and
generate a Sandstorm API key to give to cpass-sandstorm.exe like so:

    .\cpass-sandstorm.exe init <sandstorm_key>

To see the remaining valid subcommands (such as "create", seen above), run

    .\cpass-sandstorm.exe

Enjoy!


Help and feedback
-----------------

If you have questions or feedback (which is always welcome!), feel
free to send a message the CrypTag mailing list:
<https://groups.google.com/forum/#!forum/cryptag>

If you experience a bug, you can report it here:
<https://github.com/elimisteve/cryptag/issues>


Learn more
----------

You'll find more details at:

- Conceptual overview in these slides from my DEFCON talk introducing CrypTag:
<https://talks.stevendphillips.com/cryptag-defcon23-cryptovillage/>

- GitHub repo: <https://github.com/elimisteve/cryptag>


### cget, cput

Interested in command line clients that let you save and retrieve
files?  Contact me (@elimisteve on Twitter) so I can better understand
your use case(s), and I'll happily build you one based on the
[existing client
code](https://github.com/elimisteve/cryptag/tree/master/cmd)!
