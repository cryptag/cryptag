import effWordlist from './effWordlist';

const crypto = window.crypto || window.msCrypto;

// Excluding max
//
// Mostly from https://github.com/chancejs/chancejs/issues/232#issuecomment-182500222
function randomIntsInRange(min, max, numInts){
  let rand = new Uint32Array(numInts);
  crypto.getRandomValues(rand);

  let ints = new Uint32Array(numInts);
  var zeroToOne = 0.0;

  for(let i = 0; i < numInts; i++){
    zeroToOne = rand[i]/(0xffffffff + 1);
    // TODO: Do security audit of this for timing attacks
    ints[i] = Math.floor(zeroToOne * (max - min)) + min;
  }

  return ints;
}

export function genPassphrase(numWords=25){
  let words = new Array(numWords);
  let ndxs = randomIntsInRange(0, effWordlist.length, numWords);

  for(let i = 0; i < numWords; i++){
    words[i] = effWordlist[ndxs[i]];
  }

  return words.join("");
}
