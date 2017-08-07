const joinOn = '|||';

export const options = [
  ["Text",           ["type:text"].join(joinOn)],
  ["URL (redirect)", ["type:text", "type:url", "type:urlredirect"].join(joinOn)],
  // ["Markdown",       ["type:text", "type:md"].join(joinOn)],
  // ["Password",       ["type:text", "type:password"].join(joinOn)],
  // ["URL",            ["type:text", "type:url"].join(joinOn)],
  // ["File",           ["type:file"].join(joinOn)],
];
