/**
 * Enhanced YouTube Embed Web Component
 * Features: Lazy loading, better error handling, accessibility,
 * modern API support, and performance optimizations
 */
export default class YoutubeEmbed extends HTMLElement {
  // Private fields for better encapsulation
  #videoId = null;
  #isLoaded = false;
  #intersectionObserver = null;
  #resizeObserver = null;
  #playerReady = false;

  // Static properties for better performance
  static observedAttributes = [
    "data-url",
    "data-autoplay",
    "data-muted",
    "data-start",
    "data-end",
  ];

  // Modern static properties
  static get formAssociated() {
    return false;
  }

  constructor() {
    super();

    // Create shadow DOM for better encapsulation
    this.attachShadow({ mode: "open" });
  }

  connectedCallback() {
    // We set up observers and listeners here. They are only created once.
    if (!this.#intersectionObserver) {
      this.#setupIntersectionObserver();
    }
    if (!this.#resizeObserver) {
      this.#setupResizeObserver();
    }

    // Render initial state
    this.#renderPlaceholder();

    this.addEventListener("keydown", this.#handleKeydown);

    this.#setupAccessibility();

    this.#intersectionObserver?.observe(this);
    this.#resizeObserver?.observe(this);

    this.#processUrl();
  }

  disconnectedCallback() {
    this.#intersectionObserver?.disconnect();
    this.#resizeObserver?.disconnect();
    this.removeEventListener("keydown", this.#handleKeydown);
  }

  attributeChangedCallback(name, oldValue, newValue) {
    if (oldValue === newValue) return;

    switch (name) {
      case "data-url":
        this.#processUrl();
        break;
      case "data-autoplay":
      case "data-muted":
      case "data-start":
      case "data-end":
        if (this.#isLoaded) {
          this.#updatePlayerParams();
        }
        break;
    }
  }

  #setupIntersectionObserver() {
    if (!("IntersectionObserver" in window)) return;

    this.#intersectionObserver = new IntersectionObserver(
      (entries) => {
        entries.forEach((entry) => {
          if (entry.isIntersecting) {
            this.#handleIntersection();
          }
        });
      },
      {
        rootMargin: "100px", // Start loading before it comes into view
        threshold: 0.1,
      },
    );
  }

  #setupResizeObserver() {
    if (!("ResizeObserver" in window)) return;

    this.#resizeObserver = new ResizeObserver((entries) => {
      entries.forEach((entry) => {
        this.#handleResize(entry);
      });
    });
  }

  #setupAccessibility() {
    // Set ARIA attributes
    this.setAttribute("role", "region");
    this.setAttribute("aria-label", "YouTube video player");

    // Add tabindex for keyboard navigation
    if (!this.hasAttribute("tabindex")) {
      this.setAttribute("tabindex", "0");
    }
  }

  #processUrl() {
    const url = this.dataset.url;
    if (!url || url === "#") {
      this.#videoId = null;
      this.#renderError("No valid URL provided");
      return;
    }

    const videoId = this.#extractVideoId(url);
    if (!videoId) {
      this.#renderError("Invalid YouTube URL");
      return;
    }

    // Only update if video ID changed
    if (this.#videoId !== videoId) {
      this.#videoId = videoId;
      this.#isLoaded = false;

      if (this.#isInViewport()) {
        this.#loadVideo();
      } else {
        this.#renderPlaceholder();
      }
    }
  }

  #extractVideoId(url) {
    // Enhanced regex to handle various YouTube URL formats
    const patterns = [
      /(?:youtube\.com\/watch\?v=|youtu\.be\/|youtube\.com\/embed\/|youtube\.com\/v\/)([^&\n?#]+)/,
      /youtube\.com\/watch\?.*v=([^&\n?#]+)/,
      /youtube\.com\/shorts\/([^&\n?#]+)/,
    ];

    for (const pattern of patterns) {
      const match = url.match(pattern);
      if (match && match[1]) {
        return match[1];
      }
    }
    return null;
  }

  #isInViewport() {
    const rect = this.getBoundingClientRect();
    return (
      rect.top >= 0 &&
      rect.left >= 0 &&
      rect.bottom <=
        (window.innerHeight || document.documentElement.clientHeight) &&
      rect.right <= (window.innerWidth || document.documentElement.clientWidth)
    );
  }

  #handleIntersection() {
    if (!this.#isLoaded && this.#videoId) {
      this.#loadVideo();
      // Stop observing once loaded
      this.#intersectionObserver?.unobserve(this);
    }
  }

  #handleResize(entry) {
    // Maintain aspect ratio on resize
    const { width } = entry.contentRect;
    const height = (width * 9) / 16; // 16:9 aspect ratio
    this.style.height = `${height}px`;
  }

  #handleKeydown(event) {
    // Enhanced keyboard support
    switch (event.key) {
      case "Enter":
      case " ": // Space bar
        event.preventDefault();
        if (!this.#isLoaded) {
          this.#loadVideo();
        } else {
          // Focus the iframe if it exists
          const iframe = this.shadowRoot.querySelector("iframe");
          iframe?.focus();
        }
        break;
      case "Escape":
        // Allow users to exit fullscreen or blur the component
        this.blur();
        break;
    }
  }

  #renderPlaceholder() {
    if (!this.#videoId) {
      this.#renderError("No video to display");
      return;
    }

    const thumbnailUrl = this.#getThumbnailUrl(this.#videoId);

    this.shadowRoot.innerHTML = `
      <style>
        :host {
          display: block;
          position: relative;
          width: 100%;
          aspect-ratio: 16/9;
          background: #000;
          border-radius: 12px;
          overflow: hidden;
          cursor: pointer;
          transition: transform 0.2s ease, box-shadow 0.2s ease;
          outline: none;
        }

        :host(:hover) {
          transform: scale(1.02);
          box-shadow: 0 8px 25px rgba(0, 0, 0, 0.3);
        }

        :host(:focus-visible) {
          outline: 2px solid #ff0000;
          outline-offset: 2px;
        }

        .placeholder {
          position: relative;
          width: 100%;
          height: 100%;
          display: flex;
          align-items: center;
          justify-content: center;
          background: linear-gradient(135deg, #1a1a1a 0%, #2a2a2a 100%);
          border-radius: inherit;
          overflow: hidden;
        }

        .thumbnail {
          width: 100%;
          height: 100%;
          object-fit: cover;
          transition: transform 0.3s ease;
        }

        .placeholder:hover .thumbnail {
          transform: scale(1.05);
        }

        .play-overlay {
          position: absolute;
          inset: 0;
          display: flex;
          align-items: center;
          justify-content: center;
          background: rgba(0, 0, 0, 0.3);
          transition: background-color 0.2s ease;
        }

        .placeholder:hover .play-overlay {
          background: rgba(0, 0, 0, 0.5);
        }

        .play-button {
          width: 80px;
          height: 80px;
          background: rgba(255, 0, 0, 0.9);
          border: none;
          border-radius: 50%;
          display: flex;
          align-items: center;
          justify-content: center;
          cursor: pointer;
          transition: all 0.2s ease;
          backdrop-filter: blur(10px);
          -webkit-backdrop-filter: blur(10px);
        }

        .play-button:hover {
          background: #ff0000;
          transform: scale(1.1);
          box-shadow: 0 4px 20px rgba(255, 0, 0, 0.4);
        }

        .play-icon {
          width: 30px;
          height: 30px;
          fill: white;
          margin-left: 3px; /* Optical adjustment for play icon */
        }

        .loading-spinner {
          width: 40px;
          height: 40px;
          border: 3px solid rgba(255, 255, 255, 0.3);
          border-top: 3px solid #ff0000;
          border-radius: 50%;
          animation: spin 1s linear infinite;
        }

        @keyframes spin {
          0% { transform: rotate(0deg); }
          100% { transform: rotate(360deg); }
        }

        .error-message {
          color: #ff6b6b;
          text-align: center;
          padding: 20px;
          font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif;
        }

        /* Reduced motion support */
        @media (prefers-reduced-motion: reduce) {
          :host, .thumbnail, .play-button {
            transition: none !important;
          }
          .loading-spinner {
            animation: none !important;
          }
        }
      </style>

      <div class="placeholder" role="button" aria-label="Load YouTube video">
        <img
          class="thumbnail"
          src="${thumbnailUrl}"
          alt="YouTube video thumbnail"
          loading="lazy"
          decoding="async"
        />
        <div class="play-overlay">
          <button class="play-button" aria-label="Play video">
            <svg class="play-icon" viewBox="0 0 24 24">
              <path d="M8 5v14l11-7z"/>
            </svg>
          </button>
        </div>
      </div>
    `;

    // Add click handler to load video
    this.shadowRoot
      .querySelector(".placeholder")
      .addEventListener("click", () => {
        this.#loadVideo();
      });
  }

  #renderError(message) {
    this.shadowRoot.innerHTML = `
      <style>
        :host {
          display: block;
          width: 100%;
          aspect-ratio: 16/9;
          background: #1a1a1a;
          border-radius: 12px;
          border: 2px dashed #ff6b6b;
        }

        .error-container {
          display: flex;
          flex-direction: column;
          align-items: center;
          justify-content: center;
          height: 100%;
          padding: 20px;
          color: #ff6b6b;
          font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif;
          text-align: center;
        }

        .error-icon {
          width: 48px;
          height: 48px;
          margin-bottom: 16px;
          opacity: 0.7;
        }

        .error-title {
          font-size: 18px;
          font-weight: 600;
          margin-bottom: 8px;
        }

        .error-message {
          font-size: 14px;
          opacity: 0.8;
        }
      </style>

      <div class="error-container">
        <svg class="error-icon" viewBox="0 0 24 24" fill="currentColor">
          <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm-2 15l-5-5 1.41-1.41L10 14.17l7.59-7.59L19 8l-9 9z"/>
        </svg>
        <div class="error-title">Unable to Load Video</div>
        <div class="error-message">${message}</div>
      </div>
    `;
  }

  async #loadVideo() {
    if (this.#isLoaded || !this.#videoId) return;

    // Show loading state
    this.#renderLoading();

    try {
      // Build embed URL with parameters
      const embedUrl = this.#buildEmbedUrl();

      // Create iframe with modern attributes
      await this.#renderIframe(embedUrl);

      this.#isLoaded = true;

      // Dispatch custom event
      this.dispatchEvent(
        new CustomEvent("videoloaded", {
          detail: { videoId: this.#videoId },
          bubbles: true,
        }),
      );
    } catch (error) {
      console.error("Failed to load YouTube video:", error);
      this.#renderError("Failed to load video");
    }
  }

  #renderLoading() {
    this.shadowRoot.innerHTML = `
      <style>
        :host {
          display: block;
          width: 100%;
          aspect-ratio: 16/9;
          background: #000;
          border-radius: 12px;
          overflow: hidden;
        }

        .loading-container {
          display: flex;
          align-items: center;
          justify-content: center;
          height: 100%;
          background: linear-gradient(135deg, #1a1a1a 0%, #2a2a2a 100%);
        }

        .loading-spinner {
          width: 60px;
          height: 60px;
          border: 4px solid rgba(255, 255, 255, 0.3);
          border-top: 4px solid #ff0000;
          border-radius: 50%;
          animation: spin 1s linear infinite;
        }

        @keyframes spin {
          0% { transform: rotate(0deg); }
          100% { transform: rotate(360deg); }
        }

        @media (prefers-reduced-motion: reduce) {
          .loading-spinner {
            animation: none !important;
            border-top-color: #ff6b6b;
          }
        }
      </style>

      <div class="loading-container">
        <div class="loading-spinner" aria-label="Loading video..."></div>
      </div>
    `;
  }

  #buildEmbedUrl() {
    const baseUrl = `https://www.youtube.com/embed/${this.#videoId}`;
    const params = new URLSearchParams();

    // Default parameters for better UX
    params.set("rel", "0"); // Don't show related videos from other channels
    params.set("modestbranding", "1"); // Hide YouTube logo
    params.set("iv_load_policy", "3"); // Hide annotations
    params.set("color", "white"); // Use white progress bar
    params.set("playsinline", "1"); // Play inline on mobile
    params.set("enablejsapi", "1"); // Enable JavaScript API
    params.set("origin", window.location.origin); // Set origin for security

    // Add custom parameters
    if (this.dataset.autoplay === "true") params.set("autoplay", "1");
    if (this.dataset.muted === "true") params.set("mute", "1");
    if (this.dataset.start) params.set("start", this.dataset.start);
    if (this.dataset.end) params.set("end", this.dataset.end);

    return `${baseUrl}?${params.toString()}`;
  }

  async #renderIframe(src) {
    return new Promise((resolve, reject) => {
      const iframe = document.createElement("iframe");

      // Set modern iframe attributes
      iframe.src = src;
      iframe.title = "YouTube video player";
      iframe.width = "100%";
      iframe.height = "100%";
      iframe.style.border = "none";
      iframe.style.borderRadius = "inherit";
      iframe.loading = "lazy";
      iframe.allowFullscreen = true;
      iframe.allow =
        "accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture; web-share";
      iframe.referrerPolicy = "strict-origin-when-cross-origin";

      // Security attributes
      iframe.sandbox =
        "allow-scripts allow-same-origin allow-presentation allow-forms";

      // Handle load events
      iframe.onload = () => {
        this.#playerReady = true;
        resolve();
      };

      iframe.onerror = () => {
        reject(new Error("Failed to load iframe"));
      };

      // Create container with styles
      this.shadowRoot.innerHTML = `
        <style>
          :host {
            display: block;
            width: 100%;
            aspect-ratio: 16/9;
            border-radius: 12px;
            overflow: hidden;
            background: #000;
          }

          .video-container {
            position: relative;
            width: 100%;
            height: 100%;
            background: #000;
          }

          iframe {
            position: absolute;
            top: 0;
            left: 0;
            width: 100%;
            height: 100%;
            border: none;
            border-radius: inherit;
          }

          /* Focus styles for accessibility */
          iframe:focus-visible {
            outline: 2px solid #ff0000;
            outline-offset: 2px;
          }
        </style>

        <div class="video-container"></div>
      `;

      this.shadowRoot.querySelector(".video-container").appendChild(iframe);
    });
  }

  #getThumbnailUrl(videoId, quality = "maxresdefault") {
    // Try high quality first, fallback to medium quality
    return `https://img.youtube.com/vi/${videoId}/${quality}.jpg`;
  }

  #updatePlayerParams() {
    // If video is already loaded, we'd need to reload with new params
    if (this.#isLoaded) {
      this.#isLoaded = false;
      this.#loadVideo();
    }
  }

  // Public API methods
  play() {
    if (this.#isLoaded && this.#playerReady) {
      const iframe = this.shadowRoot.querySelector("iframe");
      iframe?.contentWindow?.postMessage(
        '{"event":"command","func":"playVideo","args":""}',
        "*",
      );
    } else if (!this.#isLoaded) {
      this.#loadVideo();
    }
  }

  pause() {
    if (this.#isLoaded && this.#playerReady) {
      const iframe = this.shadowRoot.querySelector("iframe");
      iframe?.contentWindow?.postMessage(
        '{"event":"command","func":"pauseVideo","args":""}',
        "*",
      );
    }
  }

  reload() {
    this.#isLoaded = false;
    this.#playerReady = false;
    if (this.#videoId) {
      this.#loadVideo();
    }
  }

  // Getter for video ID
  get videoId() {
    return this.#videoId;
  }

  // Getter for loaded state
  get isLoaded() {
    return this.#isLoaded;
  }
}

// Register the custom element
customElements.define("youtube-embed", YoutubeEmbed);
