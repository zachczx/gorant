import htmx from 'htmx.org';

// See https://github.com/bigskysoftware/htmx/issues/1690
//
// "At the moment you need some workarounds to use an extension with Vite, Rollup or other build tools because of a missing UMD wrapper.
// Furthermore, the extensions are not minified like htmx.js."
window.htmx = htmx;
