/*Package miniLock is a modern, authenticated, asymmetric encryption protocol that conceals metadata.

go-miniLock is a total Golang rewrite of miniLock.io, enabling native code performance,
more platform flexibility, and downstream potential for automation and novel communication
media not available to the original miniLock Chromium app.

go-minilock is copyright 2015, and proudly licensed under the
GNU Affero General Public License, version 3 or later at your option.

To understand the code layout here, it may help to understand how minilock
organises ciphertext. When creating ciphertext, minilock uses a symmetric
cipher to encrypt chunks of the plaintext file, and indexes/length-prefixes each
encrypted chunk. So, the raw ciphertext is a sliced and boxed version of the
plaintext, using NaCL-style box encryption.

Minilock then puts some file metadata in a fileInfo datastructure including filename,
file hash, and the symmetric cipher key, and encrypts this datastructure to an
ephemeral key generated just for this ciphertext. This ephemeral key's private
key component is then encrypted to each intended recipient, and each recipient's
"decryptInfo" datastructure is added to the file header.

The resulting header, consisting of the fileInfo object and the various decryptInfo
objects, is prepended to the raw ciphertext, and is in turn prepended with a Header
length indicator and a magic 8-byte leader ("miniLock" in ascii). The result is a
finished miniLock file.

This means that boxing or unwrapping a message has three stages of encryption/decryption,
so the API of this library may at first seem daunting. One may well ask why I didn't
hide all the messy details; firstly, I don't believe in hiding API details merely
for aesthetics because I constantly encounter libraries that hide functions I genuinely
need access to. Secondly, because there are other ways to use the fileInfo, decryptInfo
and raw ciphertext objects that would require individual access to each level.

If you just want to encrypt a thing to a key, the function EncryptFileContentsWithStrings
does everything at once; feed it your filename/contents, your chosen GUID (eg. your email),
your passphrase, whether to include you as a recipient, and the miniLock IDs of all intended
recipients, and you'll get back a finished miniLock file.

Likewise, to decrypt a miniLock file, just pass the miniLock file contents, your
GUID/email, and your passphrase to DecryptFileContentsWithStrings, and you'll get
back the sender's ID, filename, and file contents.

Please use responsibly; crypto is easy to do wrong and can do real harm if it
exposes private information. I make no guarantees that my code is safe to use, in
this sense.
*/
package minilock
