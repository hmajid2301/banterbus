
const CACHE_NAME = "banterbus-cache-v1";
const urlsToCache = [
  "/",
  "/static/css/styles.css",
  "/static/js/htmx.min.js",
  "/static/js/htmx.ws.js",
  "/static/js/alpine.min.js",
  "/static/images/favicon.ico",
  "/static/images/favicon.svg",
  "/static/images/apple-touch-icon.png",
  "/static/images/favicon-48x48.png",
  "/static/site.webmanifest"
];

self.addEventListener("install", (event) => {
  event.waitUntil(
    caches.open(CACHE_NAME).then((cache) => {
      return cache.addAll(urlsToCache);
    }),
  );
});

self.addEventListener("fetch", (event) => {
  event.respondWith(
    caches.match(event.request).then((response) => {
      return response || fetch(event.request);
    }),
  );
});
