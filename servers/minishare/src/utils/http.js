export function onSuccessBlob(goodStatusCode, response) {
  if (response.status !== goodStatusCode) {
    // On error, server should always respond with {"error": "..."}
    throw response.json();
  }
  return response.blob();
}

export function onSuccessJSON(goodStatusCode, response) {
  if (response.status !== goodStatusCode) {
    // On error, server should always respond with {"error": "..."}
    throw response.json();
  }
  return response.json();
}

function _postHeaders({recipientIDs}) {
  return {
    'X-Minilock-Recipient-Ids': recipientIDs.join('; ')
  }
}

export function httpPost(urlSuffix, recipientIDs, payload) {
  return fetch(window.location.origin + urlSuffix, {
    method: 'POST',
    headers: _postHeaders({recipientIDs}),
    body: payload
  })
  .then(onSuccessBlob.bind(this, 201))
}

export function httpGet(urlSuffix, authToken) {
  return fetch(window.location.origin + urlSuffix, {
    headers: {
      Authorization: 'Bearer ' + authToken
    }
  })
  .then(onSuccessJSON.bind(this, 200))
}
