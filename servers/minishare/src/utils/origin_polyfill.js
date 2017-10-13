// window.location.origin polyfill, as per
// https://stackoverflow.com/a/25495161/197160 --
if (!window.location.origin) {
  window.location.origin = window.location.protocol + "//" +
    window.location.hostname +
    (window.location.port ? ':' + window.location.port : '');
}
