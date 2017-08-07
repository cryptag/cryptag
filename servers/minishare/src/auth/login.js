import miniLock from '../utils/miniLock';


function getAuthUrl() {
  return window.location.origin + '/api/login';
}

function getAuthHeaders(mID) {
  return {
    'X-Minilock-Id': mID
  }
}

function onLoginError(reason, displayAlert, loginErrorCallback) {
  console.log("Error logging in:", reason);
  if (reason.toString() === "TypeError: Failed to fetch") {
    console.log("Trying to log in again");
    loginErrorCallback();
    return;
  }
  displayAlert(reason, 'danger');
}

function onLoginSuccess(response) {
  if (response.status !== 200) {
    throw response.json();
  }
  return response.blob();
}

// From https://github.com/kaepora/miniLock/blob/ffea0ecb7a619d921129b8b4aed2081050ec48c1/src/js/miniLock.js#L592-L595 --
//
//    miniLock.crypto.decryptFile's callback is passed these parameters:
//      file: Decrypted file object (blob),
//      saveName: File name for saving the file (String),
//      senderID: Sender's miniLock ID (Base58 string)
function decryptAuthToken(loginCompleteCallback, fileBlob, saveName, senderID) {
  let reader = new FileReader();
  reader.addEventListener("loadend", () => {
    let authToken = reader.result;
    console.log('authToken:', authToken);
    loginCompleteCallback(authToken);
  });

  reader.readAsText(fileBlob);
}

export function decryptMessage(mID, secretKey, message, decryptFileCallback) {
  console.log("Trying to decrypt", message);

  miniLock.crypto.decryptFile(message,
    mID,
    secretKey,
    decryptFileCallback);
}

export function minishareLogin(minilockID, secretKey, loginCompleteCallback, loginErrorCallback, displayAlert, authURL=getAuthUrl()) {
  return fetch(authURL, {
    headers: getAuthHeaders(minilockID)
  })
    .then(onLoginSuccess)
    .then((body) => {
      decryptMessage(minilockID,
                     secretKey,
                     body,
                     decryptAuthToken.bind(this, loginCompleteCallback))
    })
    .catch((reason) => {
      if (reason.then) {
        reason.then((errjson) => {
          onLoginError(errjson.error, displayAlert, loginErrorCallback);
        })
        return;
      }
      onLoginError(reason, displayAlert, loginErrorCallback);
      return;
    });
}

export function base64DecodeToBlob(data, type='application/octet-stream') {
  let binStr = atob(data);
  let binStrLength = binStr.length;
  let array = new Uint8Array(binStrLength);

  for (let i = 0; i < binStrLength; i++) {
    array[i] = binStr.charCodeAt(i);
  }
  let msg = new Blob([array], { type: type });
  return msg;
}
