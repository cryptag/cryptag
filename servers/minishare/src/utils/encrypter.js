import { genPassphrase } from '../data/minishare';

const sha384 = require('js-sha512').sha384;

const emailDomain = '@cryptag.org';

export function getEmail(passphrase){
  return sha384(passphrase) + emailDomain;
}

export function getPassphrase(documentHash){
  let isNewPassphrase = false;

  let passphrase = documentHash || '#';
  passphrase = passphrase.slice(1);

  // Generate new passphrase for user if none specified (that is, if the
  // URL hash is blank)
  if (!passphrase){
    passphrase = genPassphrase();
    isNewPassphrase = true;
  }

  return {
    passphrase,
    isNewPassphrase
  };
}

// TODO: Do smarter msgKey creation
export function generateMessageKey(i){
  let date = new Date();
  return date.toGMTString() + ' - ' + date.getSeconds() + '.' + date.getMilliseconds() + '.' + i;
}
