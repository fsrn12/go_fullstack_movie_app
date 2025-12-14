import API from "../services/API.js";

import "./YoutubeEmbed.js";

// Dynamic CSS Module Import
const cssModule = await import("../MovieDetails.css", {
  with: { type: "css" },
});
document.adoptedStyleSheets = [
  ...document.adoptedStyleSheets,
  cssModule.default,
];

export default class MovieDetailsPage extends HTMLElement {
  // Private fields (ES2022+)
  #movieId = null;
  #movieData = null;
  #abortController = null;
  #intersectionObserver = null;
  // #loadingStates = new Map();
  #animationTimeline = null;

  // Static observedAttributes for better performance
  static observedAttributes = ["movie-id", "theme"];
  // Static configuration
  static #CACHE_DURATION = 5 * 60 * 1000; // 5 minutes
  static #cache = new Map();

  constructor() {
    super();
    this.attachShadow({ mode: "open", delegatesFocus: true });
  }

  // Modern getter/setter with private field
  get movieId() {
    return this.#movieId;
  }
  set movieId(value) {
    const oldValue = this.#movieId;
    this.#movieId = value;
    if (oldValue !== value) {
      this.#fetchAndRender();
    }
  }

  async connectedCallback() {
    this.#movieId = this.params?.[0] ?? this.getAttribute("movie-id") ?? null;

    if (!this.#movieId) {
      this.#renderError("No movie ID provided");
      return;
    }

    // Initialize intersection observer for lazy loading
    this.#setupIntersectionObserver();

    // Fetch and render movie data
    this.#fetchAndRender();
    // await this.#fetchAndRender();
  }

  disconnectedCallback() {
    // Cleanup resources
    this.#abortController?.abort();
    this.#intersectionObserver?.disconnect();
    this.#animationTimeline?.cancel?.();

    const handleScroll = (event) => this.#handleScroll(event);
    window.removeEventListener("scroll", handleScroll);
    const handleKeyboard = (event) => this.#handleKeyboard(event);
    this.removeEventListener("keydown", handleKeyboard);
  }

  attributeChangedCallback(name, oldValue, newValue) {
    if (oldValue === newValue) return;

    switch (name) {
      case "movie-id":
        this.movieId = newValue;
        break;
      case "theme":
        this.#updateTheme(newValue);
        break;
    }
  }

  // ================================================
  // FETCHING AND RENDERING BUSINESS LOGIC
  // ================================================

  async #fetchAndRender() {
    // Cancel previous requests
    this.#abortController?.abort();
    this.#abortController = new AbortController();

    try {
      this.#renderLoadingSkeleton();

      // Check cache first
      const cacheKey = `movie_${this.#movieId}`;
      const cached = MovieDetailsPage.#cache.get(cacheKey);

      if (
        cached &&
        Date.now() - cached.timestamp < MovieDetailsPage.#CACHE_DURATION
      ) {
        this.#movieData = cached.data;
        await this.#renderMovieContent();
        return;
      }

      // Fetch movie data with timeout and retry logic
      const movieData = await this.#fetchMovieWithRetry();
      this.#movieData = movieData;

      // Cache the result
      MovieDetailsPage.#cache.set(cacheKey, {
        data: this.#movieData,
        timestamp: Date.now(),
      });

      // Pre-load critical resources
      this.#preloadCriticalResources();

      // Render main content with smooth transitions
      await this.#renderMovieContent();

      // Initialize interactive features
      this.#initializeInteractivity();
    } catch (error) {
      if (error.name !== "AbortError") {
        console.error("Failed to load movie:", error);
        this.#renderError(`Failed to load movie: ${error.message}`);
      }
    }
  }

  async #fetchMovieWithRetry(maxRetries = 3) {
    for (let attempt = 1; attempt <= maxRetries; attempt++) {
      try {
        const response = await API.getMovieByID(this.#movieId, {
          signal: this.#abortController.signal,
          timeout: 10000, // 10s timeout
        });
        return response.data;
      } catch (error) {
        if (attempt === maxRetries || error.name === "AbortError") {
          throw error;
        }
        // Exponential backoff
        await new Promise((resolve) =>
          setTimeout(resolve, Math.pow(2, attempt) * 1000),
        );
      }
    }
  }

  async #preloadCriticalResources() {
    const { poster_url, casting = [] } = this.#movieData;

    // Preload poster and first few cast images
    const imageUrls = [
      poster_url,
      ...casting
        .slice(0, 6)
        .map((actor) => actor.image_url)
        .filter(Boolean),
    ].filter(Boolean);

    // Use modern fetch with high priority for critical images
    await Promise.allSettled(
      imageUrls.map((url) =>
        fetch(url, {
          priority: "high",
          signal: this.#abortController.signal,
        }),
      ),
    );
  }

  #renderLoadingSkeleton() {
    this.shadowRoot.innerHTML = `
      <style>
        @import url('./MovieDetails.css');

        .skeleton {
          background: linear-gradient(90deg,
            var(--skeleton-base) 25%,
            var(--skeleton-highlight) 50%,
            var(--skeleton-base) 75%
          );
          background-size: 200% 100%;
          animation: skeleton-loading 1.5s infinite;
        }

        @keyframes skeleton-loading {
          0% { background-position: 200% 0; }
          100% { background-position: -200% 0; }
        }
      </style>

      <article class="movie-details loading" part="container">
        <div class="hero-section">
          <div class="skeleton poster-skeleton"></div>
          <div class="content-overlay">
            <div class="skeleton title-skeleton"></div>
            <div class="skeleton tagline-skeleton"></div>
            <div class="skeleton metadata-skeleton"></div>
          </div>
        </div>
      </article>
    `;
  }

  async #renderMovieContent() {
    const {
      title,
      tagline,
      poster_url,
      trailer_url,
      overview,
      release_year,
      score,
      popularity,
      genres = [],
      casting = [],
      id: movieId,
      backdrop_url,
    } = this.#movieData;

    // Create main template with modern CSS features
    this.shadowRoot.innerHTML = `
      <style>
        @import url('./MovieDetails.css');
      </style>

      <article class="movie-details" part="container" style="view-transition-name: movie-${movieId}">

        <!-- Hero Section with Backdrop -->
        <section class="hero-section" part="hero">
          <div class="backdrop-container">
            <img
              class="backdrop-image"
              src="${backdrop_url || poster_url}"
              alt=""
              loading="eager"
              decoding="sync"
              fetchpriority="high"
            />
            <div class="backdrop-overlay"></div>
          </div>

          <!-- Main Content -->
          <div class="hero-content">
            <div class="poster-section">

            <div class="poster-container">
             <div class="poster-glow"></div>
              <img
                class="movie-poster"
                src="${poster_url || "/images/placeholder-poster.jpg"}"
                alt="${title} Poster"
                loading="lazy"
                decoding="async"
                fetchpriority="high"
                part="poster"
              />
              <div class="poster-overlay"></div>
                <div class="floating-score">
                    ${score ? score.toFixed(1) : "?"}
                </div>
              <div class="poster-reflection"></div>
            </div>
                  <!--ACTION BUTTONS-->
                <div class="action-buttons" part="actions">
                <button
                  id="btn-favorites"
                  class="action-btn primary btn-primary"
                  data-movie-id="${movieId}"
                  data-action="favorite"
                  aria-label="Add to favorites"
                >
                  <svg class="btn-icon" viewBox="0 0 24 24">
                    <path d="M12 21.35l-1.45-1.32C5.4 15.36 2 12.28 2 8.5 2 5.42 4.42 3 7.5 3c1.74 0 3.41.81 4.5 2.09C13.09 3.81 14.76 3 16.5 3 19.58 3 22 5.42 22 8.5c0 3.78-3.4 6.86-8.55 11.54L12 21.35z"/>
                  </svg>
                  <span class="btn-text">Favorite</span>
                  <div class="btn-ripple"></div>
                </button>

                <button
                  id="btn-watchlist"
                  class="action-btn secondary btn-secondary"
                  data-movie-id="${movieId}"
                  ata-action="watchlist"
                  aria-label="Add to watchlist"
                >
                  <svg class="btn-icon" viewBox="0 0 24 24">
                    <path d="M17 3H5c-1.11 0-2 .9-2 2v16l7-3 7 3V5c0-1.1-.89-2-2-2z"/>
                  </svg>
                  <span class="btn-text">Watchlist</span>
                  <div class="btn-ripple"></div>
                </button>
                </div>
            </div>

            <div class="movie-info">
              <header class="movie-header">
                <h1 class="movie-title" part="title">${title}</h1>
                <p class="movie-tagline" part="tagline">${tagline || ""}</p>

                <div class="metadata-grid" part="metadata">
                  <div class="metadata-item">
                    <span class="metadata-label">Year</span>
                    <span class="metadata-value">${release_year || "N/A"}</span>
                  </div>
                  <div class="metadata-item">
                    <span class="metadata-label">Rating</span>
                    <div class="rating-display">
                      <span class="rating-value">${score || "?"}</span>
                      <span class="rating-max">/10</span>
                      <div class="rating-visual" style="--rating: ${
                        (score || 0) * 10
                      }%"></div>
                    </div>
                  </div>
                  <div class="metadata-item">
                    <span class="metadata-label">Popularity</span>
                    <span class="metadata-value">${
                      // this.#formatPopularity(popularity)
                      popularity.toFixed(1)
                    }
                    </span>
                  </div>
                </div>
              </header>

               <!-- Overview -->
            <div class="section-block">
              <div class="container">
                   <h3 class="section-title">Overview</h3>
              <p class="movie-overview" part="overview">${
                overview || "No overview available."
              }</p>
              </div>
            </div>
            </div>
          </div>
        </section>

        <!-- Details Section -->
        <section class="details-section" part="details">
          <div class="container">

            <!-- Genres -->
            ${
              genres.length
                ? `
            <div class="section-block">
              <h3 class="section-title">Genres</h3>
              <div class="genres-grid" part="genres">
                ${genres
                  .map(
                    (genre) => `
                  <span class="genre-tag" data-genre-id="${genre.id}">
                    ${genre.name}
                  </span>
                `,
                  )
                  .join("")}
              </div>
            </div>
              `
                : ""
            }

            <!-- TRAILER -->
       ${
         trailer_url
           ? `
                    <div class="trailer-section">
                        <h2 class="section-title">Trailer</h2>
                        <youtube-embed
                            data-url="${trailer_url}"
                            style="width: 100%; max-width: 800px; height:100%; margin: 0 auto; display: block;">
                        </youtube-embed>
                    </div>
                `
           : ""
       }


            <!-- Cast -->
            ${
              casting.length
                ? `
        <div class="section-block">
          <h3 class="section-title">Cast</h3>
          <div class="cast-grid" part="cast">
           ${casting
             .map(
               (actor, index) => `
              <div class="cast-member" data-actor-id="${
                actor.id
              }" style="--delay: ${index * 0.1}s">
                <div class="actor-image-container">
                  <img
                    class="actor-image lazy-image"
                    data-src="${actor.image_url || "/images/generic_actor.jpg"}"
                    alt="${actor.first_name} ${actor.last_name}"
                    loading="lazy"
                    decoding="async"
                  />
                  <div class="image-placeholder"></div>
                </div>
                <div class="actor-info">
                  <p class="actor-name">${actor.first_name} ${
                 actor.last_name
               }</p>
                </div>
              </div>
            `,
             )
             .join("")}

          </div>
        </div>
        `
                : ""
            }
          </div>
        </section>

        <!-- Trailer Modal -->
        <div id="trailer-modal" class="modal" part="modal">
          <div class="modal-backdrop"></div>
          <div class="modal-content">
            <button class="modal-close" aria-label="Close trailer">
              <svg viewBox="0 0 24 24">
                <path d="M19 6.41L17.59 5 12 10.59 6.41 5 5 6.41 10.59 12 5 17.59 6.41 19 12 13.41 17.59 19 19 17.59 13.41 12z"/>
              </svg>
            </button>
            <div class="trailer-container">
              <iframe
                id="trailer-iframe"
                title="Movie Trailer"
                allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture; web-share"
                allowfullscreen
              ></iframe>
            </div>
          </div>
        </div>
      </article>
    `;

    // Initialize lazy loading for images
    this.#initializeLazyLoading();

    // Attach event listeners
    this.#attachEventListeners();

    // Trigger entrance animations
    await this.#animateEntrance();

    // Update document title
    document.title = `${title || "Movie"} - Epic Movies`;

    // Setup SEO meta tags
    this.#updateMetaTags();
  }

  #initializeInteractivity() {
    const shadowRoot = this.shadowRoot;

    // Modern event delegation with better performance
    shadowRoot.addEventListener("click", this.#handleClick.bind(this), {
      passive: false,
    });
    shadowRoot.addEventListener("keydown", this.#handleModalKeydown.bind(this));

    // Initialize ripple effects for buttons
    this.#initializeRippleEffects();

    // Setup scroll-driven animations
    this.#setupScrollAnimations();
  }

  #initializeLazyLoading() {
    const lazyImages = this.shadowRoot.querySelectorAll(".lazy-image");
    for (let img of lazyImages) {
      this.#intersectionObserver?.observe(img);
    }
  }

  #setupIntersectionObserver() {
    this.#intersectionObserver = new IntersectionObserver(
      (entries) => {
        for (let entry of entries) {
          if (entry.isIntersecting) {
            const img = entry.target;
            const src = img.dataset.src;

            if (src) {
              // Modern image loading with error handling
              const imageLoader = new Image();
              imageLoader.onload = () => {
                img.src = src;
                img.classList.add("loaded");
              };
              imageLoader.onerror = () => {
                img.src = "/images/generic_actor.jpg";
                img.classList.add("error");
              };
              imageLoader.src = src;

              img.removeAttribute("data-src");
              this.#intersectionObserver.unobserve(img);
            }
          }
        }
        // entries.forEach((entry) => { });
      },
      {
        rootMargin: "50px",
        threshold: 0.1,
      },
    );
  }

  #attachEventListeners() {
    // Use event delegation for better performance
    this.addEventListener("click", this.#handleClick.bind(this));
    this.addEventListener("keydown", this.#handleKeyboard.bind(this));

    window.addEventListener("scroll", (event) => this.#handleScroll(event), {
      passive: true,
    });

    // Setup poster 3D hover effect
    const poster = this.querySelector(".poster-container");
    if (poster) {
      poster.addEventListener(
        "mousemove",
        this.#handlePosterMouseMove.bind(this),
      );
      poster.addEventListener(
        "mouseleave",
        this.#handlePosterMouseLeave.bind(this),
      );
    }
  }

  async #handleClick(event) {
    const target = event.target.closest('[id^="btn-"]');
    if (!target) return;

    // Add haptic feedback if supported
    if ("vibrate" in navigator) {
      navigator.vibrate(10);
    }

    const movieID = parseInt(this.#movieId, 10);

    // const movieId = target.dataset.movieId;
    switch (target.id) {
      case "btn-favorites":
        // await window.app.saveToCollection(movieID, "favorite");
        // window.app.Router.go("/account/favorites");
        this.#handleFavoriteClick(movieID);
        break;
      case "btn-watchlist":
        // await window.app.saveToCollection(movieID, "watchlist");
        // window.app.Router.go("/account/watchlist");
        this.#handleWatchlistClick(movieID);
        break;
      case "btn-trailer":
        this.#openTrailerModal(target.dataset.trailerUrl);

        break;
      case "modal-close":
      case "trailer-modal":
        if (event.target === target) {
          this.#closeTrailerModal();
        }
        break;
    }
  }

  async #handleFavoriteClick(movieId) {
    try {
      const button = this.shadowRoot.getElementById("btn-favorites");
      button.classList.add("loading");

      // const movieIdNumber = parseInt(movieId, 10);
      const response = await window.app?.saveToCollection?.(
        movieId,
        "favorite",
      );

      // Smooth success animation
      button.classList.remove("loading");
      button.classList.add("success");
      button.querySelector(".btn-text").textContent = "✅ Added!";

      window.app.publish("collection-updated", { type: "favorites" });

      if (response && response.data && response.data.success) {
        setTimeout(() => {
          button.classList.remove("success");
          window.app.Router.go("/account/favorites");
        }, 900);
      }
    } catch (error) {
      console.error("Failed to add to favorites:", error);
      this.#showToast("Failed to add to favorites", "error");
    }
  }

  async #handleWatchlistClick(movieId) {
    try {
      const button = this.shadowRoot.getElementById("btn-watchlist");
      button.classList.add("loading");
      // const movieIdNumber = parseInt(movieId, 10);

      const response = await window.app?.saveToCollection?.(
        movieId,
        "watchlist",
      );

      button.classList.remove("loading");
      button.classList.add("success");
      button.querySelector(".btn-text").textContent = "✅ Added!";

      window.app.publish("collection-updated", { type: "watchlist" });

      if (response && response.data && response.data.success) {
        setTimeout(() => {
          button.classList.remove("success");
          window.app.Router.go("/account/watchlist");
        }, 900);
      }
    } catch (error) {
      console.error("Failed to add to watchlist:", error);
      this.#showToast("Failed to add to watchlist", "error");
    }
  }

  #openTrailerModal(trailerUrl) {
    const modal = this.shadowRoot.getElementById("trailer-modal");
    const iframe = this.shadowRoot.getElementById("trailer-iframe");

    // Extract YouTube video ID and create embed URL
    const videoId = this.#extractYouTubeId(trailerUrl);
    if (!videoId) return;

    iframe.src = `https://www.youtube.com/embed/${videoId}?autoplay=1&rel=0&modestbranding=1`;

    // Show modal with animation
    modal.classList.add("active");
    document.body.style.overflow = "hidden";

    // Focus management for accessibility
    modal.querySelector(".modal-close").focus();
  }

  #closeTrailerModal() {
    const modal = this.shadowRoot.getElementById("trailer-modal");
    const iframe = this.shadowRoot.getElementById("trailer-iframe");

    modal.classList.remove("active");
    iframe.src = "";
    document.body.style.overflow = "";
  }

  /** Handle poster 3D mouse movement*/
  #handlePosterMouseMove(event) {
    if (window.matchMedia("(prefers-reduced-motion: reduce)").matches) return;

    const rect = event.currentTarget.getBoundingClientRect();
    const x = event.clientX - rect.left;
    const y = event.clientY - rect.top;
    const centerX = rect.width / 2;
    const centerY = rect.height / 2;

    const rotateX = (y - centerY) / 20;
    const rotateY = (centerX - x) / 20;

    event.currentTarget.style.transform = `perspective(1000px) rotateX(${rotateX}deg) rotateY(${rotateY}deg) translateZ(20px)`;
  }

  /** Reset poster transform on mouse leave*/
  #handlePosterMouseLeave(event) {
    event.currentTarget.style.transform = "";
  }

  #extractYouTubeId(url) {
    const regex =
      /(?:youtube\.com\/watch\?v=|youtu\.be\/|youtube\.com\/embed\/)([^&\n?#]+)/;
    const match = url?.match(regex);
    return match?.[1] || null;
  }

  async #animateEntrance() {
    // Use modern Web Animations API for smooth entrance
    const elements = [
      { selector: ".hero-section", delay: 0 },
      { selector: ".poster-container", delay: 200 },
      { selector: ".movie-info", delay: 400 },
      { selector: ".details-section", delay: 600 },
    ];

    const animations = elements
      .map(({ selector, delay }) => {
        const element = this.shadowRoot.querySelector(selector);
        if (!element) return null;

        return element.animate(
          [
            {
              opacity: 0,
              transform: "translateY(30px) scale(0.95)",
              filter: "blur(10px)",
            },
            {
              opacity: 1,
              transform: "translateY(0) scale(1)",
              filter: "blur(0)",
            },
          ],
          {
            duration: 800,
            delay,
            easing: "cubic-bezier(0.25, 0.46, 0.45, 0.94)",
            fill: "both",
          },
        );
      })
      .filter(Boolean);

    // Wait for all animations to complete
    await Promise.all(animations.map((anim) => anim.finished));
  }

  #initializeRippleEffects() {
    const buttons = this.shadowRoot.querySelectorAll(".action-btn");

    buttons.forEach((button) => {
      button.addEventListener("click", (e) => {
        const ripple = button.querySelector(".btn-ripple");
        const rect = button.getBoundingClientRect();
        const x = e.clientX - rect.left;
        const y = e.clientY - rect.top;

        ripple.style.left = `${x}px`;
        ripple.style.top = `${y}px`;
        ripple.classList.add("active");

        setTimeout(() => ripple.classList.remove("active"), 600);
      });
    });
  }

  #setupScrollAnimations() {
    // Modern scroll-driven animations with CSS @scroll-timeline
    if ("scrollTimeline" in window) {
      const castMembers = this.shadowRoot.querySelectorAll(".cast-member");
      castMembers.forEach((member, index) => {
        member.style.animationTimeline = "view()";
        member.style.animationName = "reveal";
        member.style.animationDelay = `${index * 0.1}s`;
      });
    }
  }

  #handleScroll() {
    // Throttled scroll handling with modern APIs
    const scrollY = window.scrollY;
    const backdrop = this.shadowRoot.querySelector(".backdrop-image");

    if (backdrop) {
      // Parallax effect with will-change optimization
      backdrop.style.transform = `translate3d(0, ${scrollY * 0.5}px, 0)`;
    }
  }

  #handleKeyboard(event) {
    // Enhanced keyboard navigation
    if (event.key === "Escape") {
      const modal = this.shadowRoot.querySelector(".modal.active");
      if (modal) {
        this.#closeTrailerModal();
        event.preventDefault();
      }
    }
  }

  #handleModalKeydown(event) {
    const modal = this.shadowRoot.getElementById("trailer-modal");
    if (!modal.classList.contains("active")) return;

    // Trap focus within modal
    if (event.key === "Tab") {
      const focusableElements = modal.querySelectorAll(
        'button, [href], input, select, textarea, [tabindex]:not([tabindex="-1"])',
      );
      const firstElement = focusableElements[0];
      const lastElement = focusableElements[focusableElements.length - 1];

      if (event.shiftKey && document.activeElement === firstElement) {
        lastElement.focus();
        event.preventDefault();
      } else if (!event.shiftKey && document.activeElement === lastElement) {
        firstElement.focus();
        event.preventDefault();
      }
    }
  }

  #showToast(message, type = "info") {
    // Modern toast notification system
    const toast = document.createElement("div");
    toast.className = `toast toast--${type}`;
    toast.textContent = message;

    document.body.appendChild(toast);

    // Animate in
    toast.animate(
      [
        { transform: "translateY(100%)", opacity: 0 },
        { transform: "translateY(0)", opacity: 1 },
      ],
      {
        duration: 300,
        easing: "cubic-bezier(0.25, 0.46, 0.45, 0.94)",
      },
    );

    // Auto remove
    setTimeout(() => {
      toast
        .animate(
          [
            { transform: "translateY(0)", opacity: 1 },
            { transform: "translateY(-100%)", opacity: 0 },
          ],
          {
            duration: 300,
            easing: "cubic-bezier(0.25, 0.46, 0.45, 0.94)",
          },
        )
        .finished.then(() => {
          toast.remove();
        });
    }, 3000);
  }

  #updateTheme(theme) {
    this.setAttribute("data-theme", theme);
  }

  #renderError(message) {
    this.shadowRoot.innerHTML = `
      <style>
        .error-container {
          display: flex;
          flex-direction: column;
          align-items: center;
          justify-content: center;
          min-height: 400px;
          padding: 2rem;
          text-align: center;
        }

        .error-icon {
          width: 64px;
          height: 64px;
          margin-bottom: 1rem;
          opacity: 0.5;
        }

        .error-message {
          font-size: 1.25rem;
          color: var(--color-text-secondary, #666);
          margin-bottom: 1rem;
        }

        .retry-button {
          padding: 0.75rem 1.5rem;
          background: var(--color-primary, #007bff);
          color: white;
          border: none;
          border-radius: 0.5rem;
          font-size: 1rem;
          cursor: pointer;
          transition: background-color 0.2s ease;
        }

        .retry-button:hover {
          background: var(--color-primary-dark, #0056b3);
        }
      </style>

      <div class="error-container">
        <svg class="error-icon" viewBox="0 0 24 24" fill="currentColor">
          <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm-2 15l-5-5 1.41-1.41L10 14.17l7.59-7.59L19 8l-9 9z"/>
        </svg>
        <p class="error-message">${message}</p>
        <button class="retry-button" onclick="this.getRootNode().host.#fetchAndRender()">
          Retry
        </button>
      </div>
    `;
  }

  /**
   * Update SEO meta tags
   */
  #updateMetaTags() {
    const { title, overview, poster_url } = this.#movieData;

    // Update or create meta tags
    this.#updateMetaTag("description", overview);
    this.#updateMetaTag("og:title", title);
    this.#updateMetaTag("og:description", overview);
    this.#updateMetaTag("og:image", poster_url);
    this.#updateMetaTag("og:url", window.location.href);
    this.#updateMetaTag("twitter:card", "summary_large_image");
    this.#updateMetaTag("twitter:title", title);
    this.#updateMetaTag("twitter:description", overview);
    this.#updateMetaTag("twitter:image", poster_url);
  }

  /**
   * Helper to update or create meta tags
   */
  #updateMetaTag(name, content) {
    if (!content) return;

    let meta = document.querySelector(
      `meta[name="${name}"], meta[property="${name}"]`,
    );

    if (!meta) {
      meta = document.createElement("meta");
      const attribute =
        name.startsWith("og:") || name.startsWith("twitter:")
          ? "property"
          : "name";
      meta.setAttribute(attribute, name);
      document.head.appendChild(meta);
    }

    meta.setAttribute("content", content);
  }

  /**
   * Refresh the component data
   */
  async refresh() {
    if (this.#movieId) {
      // Clear cache for this movie
      const cacheKey = `movie_${this.#movieId}`;
      MovieDetailsPage.#cache.delete(cacheKey);

      await this.#fetchAndRender();
    }
  }

  /**
   * Get current movie data
   */
  getMovieData() {
    return this.#movieData;
  }
}

// Register the custom element
customElements.define("movie-details-page", MovieDetailsPage);
