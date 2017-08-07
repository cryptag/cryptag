export function tagByPrefix(plaintags, ...prefixes) {
  let prefix = '';
  for (let i = 0; i < prefixes.length; i++) {
    prefix = prefixes[i];
    for (let j = 0; j < plaintags.length; j++) {
      if (plaintags[j].startsWith(prefix)) {
        return plaintags[j];
      }
    }
  }
  return '';
}

export function tagByPrefixStripped(plaintags, ...prefixes) {
  let tag = '';
  for (let i = 0; i < prefixes.length; i++) {
    tag = tagByPrefix(plaintags, prefixes[i]);
    if (tag !== '') {
      return tag.slice(prefixes[i].length);
    }
  }

  return '';
}

export function tagsByPrefix(plaintags, prefix) {
  let tags = [];
  for (let i = 0; i < plaintags.length; i++) {
    if (plaintags[i].startsWith(prefix)) {
      tags.push(plaintags[i]);
    }
  }
  return tags;
}

export function tagsByPrefixStripped(plaintags, prefix) {
  let stripped = [];
  for (let i = 0; i < plaintags.length; i++) {
    if (plaintags[i].startsWith(prefix)) {
      // Strip off prefix
      stripped.push(plaintags[i].slice(prefix.length));
    }
  }
  return stripped;
}
