// ------------------------
// ⚙️ CONFIGURATION
// ------------------------
const VERSION = 1;
const CACHE_NAME = `multipass-v${VERSION}`;
const API_CACHE = "api-cache";
const RETRY_LIMIT = 3;
const RETRY_DELAY = 800;

// ------------------------
// 🔁 INSTALL & ACTIVATE
// ------------------------
self.addEventListener("install", () => {
  self.skipWaiting();
  console.log("[SW] Installed");
});

self.addEventListener("activate", (event) => {
  event.waitUntil(
    (async () => {
      const keys = await caches.keys();
      await Promise.all(
        keys.map((key) => (key !== CACHE_NAME ? caches.delete(key) : null)),
      );
      await self.clients.claim();
      console.log("[SW] Activated");
    })(),
  );
});

// ------------------------
// 🔄 RETRY FETCH (Exponential Backoff)
// ------------------------
const retryFetch = async (
  request,
  retries = RETRY_LIMIT,
  delay = RETRY_DELAY,
) => {
  for (let attempt = 0; attempt <= retries; attempt++) {
    try {
      return await fetch(request);
    } catch (err) {
      if (attempt === retries) throw err;
      await new Promise((res) => setTimeout(res, delay * 2 ** attempt));
    }
  }
};

// ------------------------
// 🌐 FETCH HANDLER
// ------------------------
self.addEventListener("fetch", (event) => {
  const { request } = event;
  const url = new URL(request.url);

  // Only GET requests are cached
  if (request.method !== "GET") return;

  // -------- DYNAMIC API (Always network-first) --------
  if (
    url.pathname.includes("random") ||
    url.pathname.includes("/account/favorites") ||
    url.pathname.includes("/account/watchlist")
  ) {
    return event.respondWith(
      retryFetch(request)
        .then((res) => res)
        .catch(() => new Response("Offline", { status: 503 })),
    );
  }

  // -------- API REQUESTS (Network first with fallback) --------
  if (url.pathname.startsWith("/api/")) {
    return event.respondWith(
      (async () => {
        const cache = await caches.open(API_CACHE);
        try {
          const res = await retryFetch(request);
          cache.put(request, res.clone());
          return res;
        } catch {
          const cached = await cache.match(request);
          return cached || new Response("Offline", { status: 503 });
        }
      })(),
    );
  }

  // -------- STATIC FILES (Stale-While-Revalidate) --------
  if (
    url.pathname.match(
      /\.(js|css|html|png|jpg|jpeg|svg|gif|woff2?|ttf|eot|ico)$/,
    )
  ) {
    return event.respondWith(
      (async () => {
        const cache = await caches.open(CACHE_NAME);
        const cached = await cache.match(request);

        const fetchAndUpdate = retryFetch(request)
          .then((res) => {
            if (res.ok) cache.put(request, res.clone());
            return res;
          })
          .catch(() => cached);

        return cached || fetchAndUpdate;
      })(),
    );
  }

  // -------- FALLBACK DEFAULT (Network first with fallback) --------
  return event.respondWith(
    (async () => {
      const cache = await caches.open(CACHE_NAME);
      try {
        const res = await retryFetch(request);
        cache.put(request, res.clone());
        return res;
      } catch {
        const cached = await cache.match(request);
        return cached || new Response("Offline", { status: 503 });
      }
    })(),
  );
});

// ------------------------
// 🔁 BACKGROUND SYNC
// ------------------------
self.addEventListener("sync", (event) => {
  if (event.tag === "refresh-auth") {
    event.waitUntil(
      fetch("/api/account/refresh", {
        method: "POST",
        credentials: "include",
      }).then(() => console.log("[SW] Background auth refresh")),
    );
  }
});
