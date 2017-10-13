import React, { Component } from 'react';

import PasteForm from './PasteForm';
import LinkModal from './modals/LinkModal';
import AlertContainer from './general/AlertContainer';

import { minishareLogin, decryptMessage, base64DecodeToBlob } from '../auth/login';
import { getEmail } from '../utils/encrypter';
import miniLock from '../utils/miniLock';
import { httpGet, httpPost } from '../utils/http';
import { genPassphrase } from '../data/minishare';
import { tagsByPrefix, tagByPrefixStripped } from '../utils/tags';

import './App.css';
import '../utils/origin_polyfill';

const URL_REDIRECT_PAUSE_SECS = 4;

class App extends Component {
  constructor() {
    super();

    this.state = {
      showLinkModal: false,
      authToken: '',
      keyPair: null,
      keyPairReady: false,
      mID: '', // miniLock ID
      passphrase: '',
      alertMessage: '',
      alertStyle: 'success',

      type: 'type:text',
      title: '',
      body: '',
      isTypeURLRedirect: false,
      redirecting: false
    }
  }

  componentDidMount() {
    this.keypairFromURLHash();
  }

  displayAlert = (message, alertStyle = 'success') => {
    this.setState({
      alertMessage: message,
      alertStyle: alertStyle
    })
  }

  onError = (errStr) => {
    this.displayAlert(errStr, 'warning');
  }

  onAlertDismiss = () => {
    this.setState({
      alertMessage: ''
    })
  }

  keypairFromURLHash = (urlHash=document.location.hash) => {
    if (!urlHash || urlHash === '#') {
      return;
    }

    // User was linked here; should download and decrypt

    this.displayAlert('Downloading and decrypting...');

    const passphrase = document.location.hash.slice(1);

    console.log("Passphrase is `%s`", passphrase);

    this.setState({
      passphrase: passphrase
    }, () => {
      // TODO: Use promises
      this.setKeypair(this.login.bind(this, this.downloadShare));
    })
  }

  decryptCallback = (fileBlob, saveName, senderMinilockID) => {
    console.log(fileBlob, saveName, senderMinilockID);

    const tags = saveName.split('|||');
    console.log("Tags on received message:", tags);

    const isTypeURLRedirect = tags.indexOf('type:urlredirect') !== -1;
    const isTypeText = tags.indexOf('type:text') !== -1;
    // let isTypeFile = tags.includes('type:file');

    if (isTypeText) {
      let reader = new FileReader();
      reader.addEventListener("loadend", () => {
        const body = reader.result;
        console.log('Decrypted message:', body);

        const title = tagByPrefixStripped(tags, 'title:');

        let alertMessage = '';
        if (senderMinilockID !== this.state.mID) {
          alertMessage = 'Sent to you by ' + senderMinilockID;
        }

        this.setState({
          alertMessage,
          type: tagsByPrefix(tags, 'type:').join('|||'),
          title,
          body,
          isTypeURLRedirect
        });
      })

      reader.readAsText(fileBlob);
      return;
    }

    console.log(`decryptCallback: got unrecognized share with tags ${tags}`);
  }

  downloadShare = () => {
    const { authToken, keyPair, mID } = this.state;

    console.log("downloadShare: authToken:", authToken);

    httpGet('/shares/once', authToken)
    .then((jsonMsg) => {
      if (jsonMsg.error) {
        this.onError(jsonMsg.error);
        return;
      }

      if (jsonMsg.length > 1) {
        console.log("Ignoring all shares after the first!");
      }
      const b64b64msg = jsonMsg[0];

      const b64msg = base64DecodeToBlob(b64b64msg, 'text/plain');

      var reader = new FileReader();
      reader.addEventListener("loadend", () => {
        const b64 = reader.result;
        const msg = base64DecodeToBlob(b64);

        decryptMessage(mID, keyPair.secretKey, msg, this.decryptCallback)
      })

      reader.readAsText(b64msg);
    })
    .catch((reason) => {
      this.handleErr(reason);
    })
  }

  setKeypair = (callback=function(){}) => {
    let { passphrase } = this.state;
    if (!passphrase) {
      passphrase = genPassphrase(10);
    }

    console.log("passphrase:", passphrase);
    const email = getEmail(passphrase);

    miniLock.crypto.getKeyPair(passphrase, email, (keyPair) => {
      const mID = miniLock.crypto.getMiniLockID(keyPair.publicKey);
      console.log("mID ==", mID);

      this.setState({
        keyPair: keyPair,
        keyPairReady: true,
        mID: mID,
        passphrase: passphrase
      }, callback);
    })
  }

  login = (callback=function(){}) => {
    const loginCompleteCallback = (authToken) => {
      console.log("loginCompleteCallback: setting authToken to", authToken);
      this.setState({
        authToken: authToken
      }, callback)
    }

    const loginErrorCallback = () => {
      setTimeout(this.login, 2000);
    }

    console.log("minishareLogin about to run...");

    return minishareLogin(this.state.mID,
                          this.state.keyPair.secretKey,
                          loginCompleteCallback,
                          loginErrorCallback,
                          this.displayAlert,
                          "/login")
  }

  genPasteURL = () => {
    if (!this.state.passphrase) {
      return '';
    }
    return window.location.origin + '/#' + this.state.passphrase;
  }

  createBlob = (type, title, pasteBody) => {
    const fileBlob = new Blob([pasteBody], {type: 'text/plain'});
    const saveName = type + '|||' + 'title:'+title;
    fileBlob.name = saveName;

    return {
      fileBlob,
      saveName
    };
  }

  handleErr = (reason, errCallback=this.onError) => {
    if (reason.then) {
      // Replace confusing error with clearer one
      reason.then((errjson) => {
        if (errjson.error === 'Shares not found for that user') {
          errCallback('Not found. Either it expired, has already been' +
                      ' viewed, or never existed.');
          return;
        }

        errCallback(errjson.error);
      })
      return;
    }

    errCallback(reason);
  }

  upload = (errCallback, encryptedFileBlob, saveName, senderMinilockID) => {
    console.log("upload");
    let reader = new FileReader();
    reader.addEventListener("loadend", () => {
      // From https://stackoverflow.com/questions/9267899/arraybuffer-to-base64-encoded-string#comment55137593_11562550
      let b64encMinilockFile = btoa([].reduce.call(
        new Uint8Array(reader.result),
        function (p, c) {
          return p + String.fromCharCode(c)
        }, ''));

      // Assumes sender and recipient both use this keypair, which is
      // true right now, but may not be if we offer optional sign-in
      const recipientIDs = [senderMinilockID];

      httpPost('/shares/once', recipientIDs, b64encMinilockFile)
      .then(() => {
        errCallback('');
      })
      .catch((reason) => {
        this.handleErr(reason, errCallback);
      });
    })

    reader.readAsArrayBuffer(encryptedFileBlob);
  }

  encryptAndUpload = (fileBlob, saveName, errCallback) => {
    const onKeypairReadyCallback = () => {
      console.log("onKeypairReadyCallback");
      const { mID } = this.state;
      miniLock.crypto.encryptFile(fileBlob, saveName, [mID],
                                  mID, this.state.keyPair.secretKey,
                                  this.upload.bind(this, errCallback));
    }

    this.setKeypair(onKeypairReadyCallback);
  }

  onPasteChange = (fieldName, event) => {
    event.preventDefault();

    this.setState({
      [fieldName]: event.target.value
    })
  }

  onPasteSubmit = (event) => {
    event.preventDefault();

    this.displayAlert('Encrypting, uploading, and generating a share link...');

    const { type, title, body } = this.state;

    console.log("onPasteSubmit: %s, %s, %s", type, title, body);

    // Create and encrypt
    let { fileBlob, saveName } = this.createBlob(type, title, body);
    this.encryptAndUpload(fileBlob, saveName, (err) => {
      console.log("encryptAndUpload's callback");

      if (err) {
        this.onError(err);
        return;
      }

      this.setState({
        showLinkModal: true,
        // authToken: '', // TODO: Check if this is correct
        keyPair: null,
        keyPairReady: false,
        mID: '',
        alertMessage: 'Done!'
      })

      window.location.hash = '';
    })
  }

  onCloseModal = () => {
    this.setState({
      showLinkModal: false,
      passphrase: ''
    })
  }

  redirect = () => {
    this.setState({
      redirecting: true
    })

    let { body } = this.state;

    let secs_left = URL_REDIRECT_PAUSE_SECS + 1;

    const interval = setInterval(() => {
      let units = 'seconds';

      --secs_left;
      if (secs_left === 1) {
        units = 'second';
      }

      this.setState({
        alertMessage: `Redirecting you in ${secs_left} ${units}: ${body}`
      })

      if (secs_left === 0) {
        if (!body.startsWith('http://') && !body.startsWith('https://')) {
          body = 'http://' + body;
        }

        // TODO: Consider regex check to make sure this is actually a URL
        window.location = body;

        // Will only execute if a paste/share URL redirected to
        // another share on this same domain; normally, setting
        // `window.location` will send the user off elsewhere.
        if (body.startsWith(window.location.origin)) {
          window.location.reload();
        }

        clearInterval(interval);
      }
    }, 1000)
  }

  render() {
    const { showLinkModal } = this.state;
    const { alertMessage, alertStyle } = this.state;
    const { type, title, body, isTypeURLRedirect, redirecting } = this.state;

    if (isTypeURLRedirect && !redirecting) {
      this.redirect();
    }

    const url = this.genPasteURL();

    return (
      <div className="App">
        <div className="App-header">
          <h2>miniShare</h2>
        </div>

        <br />
        <h4>
          <strong>
            Securely share self-destructing data: text, URLs, and (soon!) files
          </strong>
        </h4>
        <br />

        {showLinkModal && <LinkModal
                            showModal={showLinkModal && url}
                            url={url}
                            onCloseModal={this.onCloseModal} />}

        {alertMessage && <div>
            <AlertContainer
              style={{maxWidth: '80vw'}}
              message={alertMessage}
              alertStyle={alertStyle}
              onAlertDismiss={this.onAlertDismiss} />
            <br />
          </div>
        }

        <PasteForm
          type={type}
          title={title}
          body={body}
          onChange={this.onPasteChange}
          onSubmit={this.onPasteSubmit} />

        <div id="footer">
          Also check out <a href="https://www.leapchat.org/">
           <strong>LeapChat</strong>
          </a>: end-to-end encrypted in-browser chat
          {" | "}
          <a href="https://www.patreon.com/cryptag">Support me on Patreon!</a>
        </div>
      </div>
    );
  }
}

export default App;
