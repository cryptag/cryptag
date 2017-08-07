import BLAKE2s from 'blake2s';
import Base58 from 'bs58';
import nacl from './crypto/nacl';
import scrypt from './crypto/scrypt';

const miniLock = {}

// -----------------------
// Settings
// -----------------------

miniLock.settings = {}

// Minimum entropy for user key
miniLock.settings.minKeyEntropy = 100

// This is where session variables are stored
miniLock.session = {
  keys: {},
  keyPairReady: false
}

// -----------------------
// Utility Functions
// -----------------------

miniLock.util = {}

// Input: none
// Result: Resets miniLock.session.currentFile
miniLock.util.resetCurrentFile = function () {
}

// Input: String
// Output: Boolean
// Notes: Validates if string is a proper miniLock ID.
miniLock.util.validateID = function (id) {
  var base58Match = new RegExp(
    '^[1-9ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz]+$'
  )
  if (
    (id.length > 55) ||
  (id.length < 40)
  ) {
    return false
  }
  if (!base58Match.test(id)) {
    return false
  }
  var bytes = Base58.decode(id)
  if (bytes.length !== 33) {
    return false
  }
  var hash = new BLAKE2s(1)
  hash.update(bytes.subarray(0, 32))
  if (hash.digest()[0] !== bytes[32]) {
    return false
  }
  return true
}

// Input: Nonce (Base64) (String), Expected nonce length in bytes (Number)
// Output: Boolean
// Notes: Validates if string is a proper nonce.
miniLock.util.validateNonce = function (nonce, expectedLength) {
  var base64Match = new RegExp(
    '^(?:[A-Za-z0-9+\/]{4})*(?:[A-Za-z0-9+\/]{2}==|[A-Za-z0-9+\/]{3}=)?$'
  )
  if (
    (nonce.length > 40) ||
  (nonce.length < 10)
  ) {
    return false
  }
  if (base64Match.test(nonce)) {
    var bytes = nacl.util.decodeBase64(nonce)
    return bytes.length === expectedLength
  }
  return false
}

// Input: String
// Output: Boolean
// Notes: Validates if string is a proper symmetric key.
miniLock.util.validateKey = function (key) {
  var base64Match = new RegExp(
    '^(?:[A-Za-z0-9+\/]{4})*(?:[A-Za-z0-9+\/]{2}==|[A-Za-z0-9+\/]{3}=)?$'
  )
  if (
    (key.length > 50) ||
  (key.length < 40)
  ) {
    return false
  }
  if (base64Match.test(key)) {
    var bytes = nacl.util.decodeBase64(key)
    return bytes.length === 32
  }
  return false
}

miniLock.util.validateEphemeral = miniLock.util.validateKey

// Input: none
// Output: Random string suitable for use as filename.
miniLock.util.getRandomFilename = function () {
  var randomBytes = nacl.randomBytes(6)
  return Base58.encode(randomBytes)
}

// Input: Filename (String)
// Output: Whether filename extension looks suspicious (Boolean)
miniLock.util.isFilenameSuspicious = function (filename) {
  var suspicious = [
    'exe', 'scr', 'url', 'com', 'pif', 'bat',
    'xht', 'htm', 'html', 'xml', 'xhtml', 'js',
    'sh', 'svg', 'gadget', 'msi', 'msp', 'hta',
    'cpl', 'msc', 'jar', 'cmd', 'vb', 'vbs',
    'jse', 'ws', 'wsf', 'wsc', 'wsh', 'ps1',
    'ps2', 'ps1xml', 'ps2xml', 'psc1', 'scf', 'lnk',
    'inf', 'reg', 'doc', 'xls', 'ppt', 'pdf',
    'swf', 'fla', 'docm', 'dotm', 'xlsm', 'xltm',
    'xlam', 'pptm', 'potm', 'ppam', 'ppsm', 'sldm',
    'dll', 'dllx', 'rar', 'zip', '7z', 'gzip',
    'gzip2', 'tar', 'fon', 'svgz', 'jnlp'
  ]
  var extension = filename.toLowerCase().match(/\.\w+$/)
  if (!extension) {
    return true
  }
  extension = extension[0].substring(1)
  return (suspicious.indexOf(extension) >= 0)
}

// Input: 4-byte little-endian Uint8Array
// Output: ByteArray converter to number
miniLock.util.byteArrayToNumber = function (byteArray) {
  var n = 0
  for (var i = 3; i >= 0; i--) {
    n += byteArray[i]
    if (i > 0) {
      n = n << 8
    }
  }
  return n
}

// Input: Number
// Output: Number as 4-byte little-endian Uint8Array
miniLock.util.numberToByteArray = function (n) {
  var byteArray = [0, 0, 0, 0]
  for (var i = 0; i < byteArray.length; i++) {
    byteArray[i] = n & 255
    n = n >> 8
  }
  return new Uint8Array(byteArray)
}

// -----------------------
// Cryptographic Functions
// -----------------------

miniLock.crypto = {}

// Chunk size (in bytes)
// Warning: Must not be less than 256 bytes
miniLock.crypto.chunkSize = 1024 * 1024 * 1;

// Input: User key hash (Uint8Array), Salt (Uint8Array), callback function
// Result: Calls scrypt which returns
//	32 bytes of key material in a Uint8Array,
//	which then passed to the callback.
miniLock.crypto.getScryptKey = function (key, salt, callback) {
  scrypt(key, salt, 17, 8, 32, 1000, function (keyBytes) {
    return callback(nacl.util.decodeBase64(keyBytes))
  }, 'base64');
}

/*
// Input: User key, user email
// Output: Whether key is strong enough
miniLock.crypto.checkKeyStrength = function (key, email) {
  var minEntropy = miniLock.settings.minKeyEntropy
  if (key.length < 32) { return false }
  if (key.match(email)) { return false }
  return (zxcvbn(key).entropy > minEntropy)
}
*/

// Input: User key (String), User salt (email) (String), callback (function)
// Result: Passes the following object to the callback:
// {
//	publicKey: Public encryption key (Uint8Array),
//	secretKey: Secret encryption key (Uint8Array)
// }
miniLock.crypto.getKeyPair = function (key, salt, callback) {
  var keyHash = new BLAKE2s(32)
  keyHash.update(nacl.util.decodeUTF8(key))
  salt = nacl.util.decodeUTF8(salt)
  miniLock.crypto.getScryptKey(keyHash.digest(), salt, function (keyBytes) {
    if (typeof (callback) === 'function') {
      callback(nacl.box.keyPair.fromSecretKey(keyBytes))
    }
  })
}

// Input: none
// Output: nonce for usage in encryption operations
miniLock.crypto.getNonce = function () {
  return nacl.randomBytes(24)
}

// Input: none
// Output: File key for usage in nacl.secretbox() encryption operations
miniLock.crypto.getFileKey = function () {
  return nacl.randomBytes(32)
}

// Input: Public encryption key (Uint8Array)
// Output: miniLock ID (Base58)
miniLock.crypto.getMiniLockID = function (publicKey) {
  if (publicKey.length !== 32) {
    throw new Error('miniLock.crypto.getMiniLockID: invalid public key size')
    return false
  }
  var id = new Uint8Array(33)
  for (var i = 0; i < publicKey.length; i++) {
    id[i] = publicKey[i]
  }
  var hash = new BLAKE2s(1)
  hash.update(publicKey)
  id[32] = hash.digest()[0]
  return Base58.encode(id)
}

// Input: Object:
//	{
//		name: File name,
//		size: File size,
//		data: File (ArrayBuffer),
//	}
// saveName: Name to use when saving resulting file. '.minilock' extension will be added.
// miniLockIDs: Array of (Base58) public IDs to encrypt for
// myMiniLockID: Sender's miniLock ID (String)
// mySecretKey: My secret key (Uint8Array)
// callback: Name of the callback function to which encrypted result is passed.
// Result: Sends file to be encrypted, with the result picked up
//	 and sent to the specified callback.
miniLock.crypto.encryptFile = function (
  file,
  saveName,
  miniLockIDs,
  myMiniLockID,
  mySecretKey,
  callback
) {
  var currentFile = miniLock.file.new();
  currentFile.fileName = file.name
  saveName += '.minilock'
  var fileKey = miniLock.crypto.getFileKey()
  var fileNonce = miniLock.crypto.getNonce().subarray(0, 16)
  currentFile.streamEncryptor = nacl.stream.createEncryptor(
    fileKey,
    fileNonce,
    miniLock.crypto.chunkSize
  )
  var paddedFileName = new Uint8Array(256)
  var fileNameBytes = nacl.util.decodeUTF8(file.name)
  if (fileNameBytes.length > paddedFileName.length) {
    throw new Error('miniLock: Encryption failed - file name is too long')
  }
  paddedFileName.set(fileNameBytes)
  currentFile.hashObject = new BLAKE2s(32)
  var encryptedChunk
  encryptedChunk = currentFile.streamEncryptor.encryptChunk(
    paddedFileName,
    false
  )
  if (!encryptedChunk) {
    throw new Error('miniLock: Encryption failed - general encryption error')
    return false
  }
  currentFile.hashObject.update(encryptedChunk)
  currentFile.encryptedChunks.push(encryptedChunk)
  miniLock.crypto.encryptNextChunk(
    file,
    0,
    saveName,
    fileKey,
    fileNonce,
    miniLockIDs,
    myMiniLockID,
    mySecretKey,
    currentFile,
    callback
  )
}

//	Input:
//		Entire file object,
//		data position on which to start decryption (number),
//		Name to use when saving the file (String),
//		fileKey (Uint8Array),
//		fileNonce (Uint8Array),
//		miniLock IDs for which to encrypt (Array),
//		sender ID (Base58 string),
//		sender long-term secret key (Uint8Array),
//		current file (Object; see miniLock.file.new() for its structure)
//		Callback to execute when last chunk has been decrypted.
//	Result: Will recursively encrypt until the last chunk,
//		at which point callbackOnComplete() is called.
//		Callback is passed these parameters:
//			file: Decrypted file object (blob),
//			saveName: File name for saving the file (String),
//			senderID: Sender's miniLock ID (Base58 string)
miniLock.crypto.encryptNextChunk = function (
  file,
  dataPosition,
  saveName,
  fileKey,
  fileNonce,
  miniLockIDs,
  myMiniLockID,
  mySecretKey,
  currentFile,
  callbackOnComplete
) {
  miniLock.file.read(
    file,
    dataPosition,
    dataPosition + miniLock.crypto.chunkSize,
    function (chunk) {
      chunk = chunk.data
      var isLast = false
      if (dataPosition >= (file.size - miniLock.crypto.chunkSize)) {
        isLast = true
      }
      var encryptedChunk
      encryptedChunk = currentFile.streamEncryptor.encryptChunk(
        chunk,
        isLast
      )
      if (!encryptedChunk) {
        throw new Error('miniLock: Encryption failed - general encryption error')
        return false
      }
      currentFile.hashObject.update(encryptedChunk)
      currentFile.encryptedChunks.push(encryptedChunk)
      if (isLast) {
        currentFile.streamEncryptor.clean()
        // Finish generating header so we can pass finished file to callback
        var ephemeral = nacl.box.keyPair()
        var header = {
          version: 1,
          ephemeral: nacl.util.encodeBase64(ephemeral.publicKey),
          decryptInfo: {}
        }
        var decryptInfoNonces = []
        for (var u = 0; u < miniLockIDs.length; u++) {
          decryptInfoNonces.push(
            miniLock.crypto.getNonce()
          )
        }
        for (var i = 0; i < miniLockIDs.length; i++) {
          var decryptInfo = {
            senderID: myMiniLockID,
            recipientID: miniLockIDs[i],
            fileInfo: {
              fileKey: nacl.util.encodeBase64(fileKey),
              fileNonce: nacl.util.encodeBase64(fileNonce),
              fileHash: nacl.util.encodeBase64(
                currentFile.hashObject.digest()
              )
            }
          }
          decryptInfo.fileInfo = nacl.util.encodeBase64(nacl.box(
            nacl.util.decodeUTF8(JSON.stringify(decryptInfo.fileInfo)),
            decryptInfoNonces[i],
            Base58.decode(miniLockIDs[i]).subarray(0, 32),
            mySecretKey
          ))
          decryptInfo = nacl.util.encodeBase64(nacl.box(
            nacl.util.decodeUTF8(JSON.stringify(decryptInfo)),
            decryptInfoNonces[i],
            Base58.decode(miniLockIDs[i]).subarray(0, 32),
            ephemeral.secretKey
          ))
          header.decryptInfo[
            nacl.util.encodeBase64(decryptInfoNonces[i])
          ] = decryptInfo
        }
        header = JSON.stringify(header)
        currentFile.encryptedChunks.unshift(
          'miniLock',
          miniLock.util.numberToByteArray(header.length),
          header
        )
        return callbackOnComplete(
          new Blob(currentFile.encryptedChunks),
          saveName,
          myMiniLockID
        )
      }
      else {
        dataPosition += miniLock.crypto.chunkSize
        return miniLock.crypto.encryptNextChunk(
          file,
          dataPosition,
          saveName,
          fileKey,
          fileNonce,
          miniLockIDs,
          myMiniLockID,
          mySecretKey,
          currentFile,
          callbackOnComplete
        )
      }
    }
  )
}


// Input: Object:
//	{
//		name: File name,
//		size: File size,
//		data: Encrypted file (ArrayBuffer),
//	}
// myMiniLockID: Sender's miniLock ID (String)
// mySecretKey: Sender's secret key (Uint8Array)
// callback: Name of the callback function to which decrypted result is passed.
// Result: Sends file to be decrypted, with the result picked up
//	and sent to the specified callback.
miniLock.crypto.decryptFile = function (
  file,
  myMiniLockID,
  mySecretKey,
  callback
) {
  var currentFile = miniLock.file.new();
  miniLock.file.read(file, 8, 12, function (headerLength) {
    headerLength = miniLock.util.byteArrayToNumber(
      headerLength.data
    )
    miniLock.file.read(file, 12, headerLength + 12, function (header) {
      try {
        header = nacl.util.encodeUTF8(header.data)
        header = JSON.parse(header)
      }
      catch (error) {
        throw new Error('miniLock: Decryption failed - could not parse header')
        return false
      }
      if (
        !header.hasOwnProperty('version')
    || header.version !== 1
      ) {
        throw new Error('miniLock: Decryption failed - invalid header version')
        return false
      }
      if (
        !header.hasOwnProperty('ephemeral')
    || !miniLock.util.validateEphemeral(header.ephemeral)
      ) {
        throw new Error('miniLock: Decryption failed - could not parse header')
        return false
      }
      // Attempt decryptInfo decryptions until one succeeds
      var actualDecryptInfo = null
      var actualDecryptInfoNonce = null
      var actualFileInfo = null
      for (var i in header.decryptInfo) {
        if (
          ({}).hasOwnProperty.call(header.decryptInfo, i)
     && miniLock.util.validateNonce(i, 24)
        ) {
          actualDecryptInfo = nacl.box.open(
            nacl.util.decodeBase64(header.decryptInfo[i]),
            nacl.util.decodeBase64(i),
            nacl.util.decodeBase64(header.ephemeral),
            mySecretKey
          )
          if (actualDecryptInfo) {
            actualDecryptInfo = JSON.parse(
              nacl.util.encodeUTF8(actualDecryptInfo)
            )
            actualDecryptInfoNonce = nacl.util.decodeBase64(i)
            break
          }
        }
      }
      if (
        !actualDecryptInfo
    || !({}).hasOwnProperty.call(actualDecryptInfo, 'recipientID')
    || actualDecryptInfo.recipientID !== myMiniLockID
      ) {
        throw new Error('miniLock: Decryption failed - File is not encrypted for this recipient')
        return false
      }
      if (
        !({}).hasOwnProperty.call(actualDecryptInfo, 'fileInfo')
    || !({}).hasOwnProperty.call(actualDecryptInfo, 'senderID')
    || !miniLock.util.validateID(actualDecryptInfo.senderID)
      ) {
        throw new Error('miniLock: Decryption failed - could not validate sender ID')
        return false
      }
      try {
        actualFileInfo = nacl.box.open(
          nacl.util.decodeBase64(actualDecryptInfo.fileInfo),
          actualDecryptInfoNonce,
          Base58.decode(actualDecryptInfo.senderID).subarray(0, 32),
          mySecretKey
        )
        actualFileInfo = JSON.parse(
          nacl.util.encodeUTF8(actualFileInfo)
        )
      }
      catch (err) {
        throw new Error('miniLock: Decryption failed - could not parse header')
        return false
      }
      // Begin actual ciphertext decryption
      var dataPosition = 12 + headerLength
      currentFile.streamDecryptor = nacl.stream.createDecryptor(
        nacl.util.decodeBase64(actualFileInfo.fileKey),
        nacl.util.decodeBase64(actualFileInfo.fileNonce),
        miniLock.crypto.chunkSize
      )
      miniLock.crypto.decryptNextChunk(
        file,
        dataPosition,
        actualFileInfo,
        actualDecryptInfo.senderID,
        headerLength,
        currentFile,
        callback
      )
    })
  })
}

//	Input:
//		Entire file object,
//		data position on which to start decryption (number),
//		fileInfo object (From header),
//		sender ID (Base58 string),
//		header length (in bytes) (number),
//		current file (Object; see miniLock.file.new() for its structure)
//		Callback to execute when last chunk has been decrypted.
//	Result: Will recursively decrypt until the last chunk,
//		at which point callbackOnComplete() is called.
//		Callback is passed these parameters:
//			file: Decrypted file object (blob),
//			saveName: File name for saving the file (String),
//			senderID: Sender's miniLock ID (Base58 string)
miniLock.crypto.decryptNextChunk = function (
  file,
  dataPosition,
  fileInfo,
  senderID,
  headerLength,
  currentFile,
  callbackOnComplete
) {
  miniLock.file.read(
    file,
    dataPosition,
    dataPosition + 4 + 16 + miniLock.crypto.chunkSize,
    function (chunk) {
      chunk = chunk.data
      var actualChunkLength = miniLock.util.byteArrayToNumber(
        chunk.subarray(0, 4)
      )
      if (actualChunkLength > chunk.length) {
        throw new Error('miniLock: Decryption failed - general decryption error')
        return false
      }
      chunk = chunk.subarray(
        0, actualChunkLength + 4 + 16
      )
      var decryptedChunk
      var isLast = false
      if (
        dataPosition >= ((file.size) - (4 + 16 + actualChunkLength))
      ) {
        isLast = true
      }
      if (dataPosition === (12 + headerLength)) {
        // This is the first chunk, containing the filename
        decryptedChunk = currentFile.streamDecryptor.decryptChunk(
          chunk,
          isLast
        )
        if (!decryptedChunk) {
          throw new Error('miniLock: Decryption failed - general decryption error')
          return false
        }
        var fileName = nacl.util.encodeUTF8(decryptedChunk.subarray(0, 256))
        while (
          fileName[fileName.length - 1] === String.fromCharCode(0x00)
        ) {
          fileName = fileName.slice(0, -1)
        }
        currentFile.fileName = fileName
        currentFile.hashObject.update(chunk.subarray(0, 256 + 4 + 16))
      }
      else {
        decryptedChunk = currentFile.streamDecryptor.decryptChunk(
          chunk,
          isLast
        )
        if (!decryptedChunk) {
          throw new Error('miniLock: Decryption failed - general decryption error')
          return false
        }
        currentFile.decryptedChunks.push(decryptedChunk)
        currentFile.hashObject.update(chunk)
      }
      dataPosition += chunk.length
      if (isLast) {
        if (
          !nacl.verify(
            new Uint8Array(currentFile.hashObject.digest()),
            nacl.util.decodeBase64(fileInfo.fileHash)
          )
        ) {
          throw new Error('miniLock: Decryption failed - could not validate file contents after decryption')
          return false
        }
        else {
          currentFile.streamDecryptor.clean()
          return callbackOnComplete(
            new Blob(currentFile.decryptedChunks),
            currentFile.fileName,
            senderID
          )
        }
      }
      else {
        return miniLock.crypto.decryptNextChunk(
          file,
          dataPosition,
          fileInfo,
          senderID,
          headerLength,
          currentFile,
          callbackOnComplete
        )
      }
    }
  )
}

// -----------------------
// File Processing
// -----------------------

miniLock.file = {}

// Input: File object, bounds within which to read, and callbacks
// Output: Callback function executed with object:
//	{
//		name: File name,
//		size: File size (bytes),
//		data: File data within specified bounds (Uint8Array)
//	}
// Error callback which is called in case of error (no parameters)
miniLock.file.read = function (file, start, end, callback, errorCallback) {
  var reader = new FileReader()
  reader.onload = function (readerEvent) {
    return callback({
      name: file.name,
      size: file.size,
      data: new Uint8Array(readerEvent.target.result)
    })
  }
  reader.onerror = function () {
    if (typeof (errorCallback) === 'function') {
      return errorCallback()
    }
    else {
      throw new Error('miniLock: File read error')
      return false
    }
  }
  reader.readAsArrayBuffer(file.slice(start, end))
}

miniLock.file.new = function () {
  return {
    fileObject: null,
    fileName: '',
    encryptedChunks: [],
    decryptedChunks: [],
    hashObject: new BLAKE2s(32),
    streamEncryptor: null,
    streamDecryptor: null
  }
}

export default miniLock;
