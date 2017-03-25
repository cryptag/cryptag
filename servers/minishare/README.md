# miniShare

miniShare is a web app that users can securely share secrets through.
Behind the scenes, miniShare uses [miniLock](https://minilock.io) for
challenge/response-based authentication. The
[cmd/cryptag](https://github.com/cryptag/cryptag/tree/master/cmd/cryptag)
CLI tool knows how to create then store miniLock-encrypted data in
miniShare, generate a shareable URL, then share that file or text with
anyone.

Future versions will allow users to determine when that data is
deleted from the server, but for now it's deleted after 1 successful
download/view so that even if the share link is intercepted by
adversary, as long as they click on it after the intended recipient,
no data is compromised.

## Instances

There is currently one public instance running at
<https://minishare.cryptag.org> and <https://minishare.io>, and that
can also be accessed as a Tor hidden service at
<http://ptga662wjtg2cie3.onion>.


## Tor Hidden Service

<https://www.torproject.org/docs/tor-hidden-service.html.en>

To run miniShare as a Tor Hidden Service (aka "Onion Service" they're
now called) on Ubuntu, add something like this to your
`/etc/tor/torrc` file, and manually create the mentioned
`minishare_hidden_service` directory:

```
# minishare
HiddenServiceDir /var/lib/tor/minishare_hidden_service/
HiddenServicePort 80 127.0.0.1:8000
```


# Development / Running

```
npm install
mkdir src
mkdir build
browserify -t [ babelify --presets [ react es2015 ] ] src/index.js -o build/app.js
go build
./minishare
```

Then view <http://localhost:8000>.
