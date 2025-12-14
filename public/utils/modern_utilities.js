/**
 * Modern Utilities for Enhanced Web Components
 * Leveraging ES2025+ features and modern Web APIs
 */

// Enhanced createNode function with modern features
export function createNode(templateId, data = {}) {
  const template = document.getElementById(templateId);
  if (!template) {
    throw new Error(`Template with id "${templateId}" not found`);
  }

  const clone = template.content.cloneNode(true);

  // Modern template interpolation with safe HTML handling
  if (Object.keys(data).length > 0) {
    interpolateTemplate(clone, data);
  }

  return clone;
}

// Safe template interpolation using modern approaches
function interpolateTemplate(fragment, data) {
  const walker = document.createTreeWalker(
    fragment,
    NodeFilter.SHOW_TEXT,
    null,
    false,
  );

  const textNodes = [];
  let node;
  while ((node = walker.nextNode())) {
    textNodes.push(node);
  }

  textNodes.forEach((textNode) => {
    const text = textNode.textContent;
    const interpolatedText = text.replace(/\$\{([^}]+)\}/g, (match, key) => {
      const value = getNestedValue(data, key.trim());
      return value !== undefined ? String(value) : match;
    });

    if (interpolatedText !== text) {
      textNode.textContent = interpolatedText;
    }
  });
}

// Safe nested object value retrieval
function getNestedValue(obj, path) {
  return path.split(".").reduce((current, key) => current?.[key], obj);
}

// Modern performance-optimized debounce with AbortController
export function debounce(func, wait, options = {}) {
  let timeoutId;
  let abortController;

  const { leading = false, trailing = true, maxWait } = options;
  let lastCallTime;
  let lastInvokeTime = 0;

  function invokeFunc(time) {
    const args = lastArgs;
    const thisArg = lastThis;

    lastArgs = lastThis = undefined;
    lastInvokeTime = time;
    result = func.apply(thisArg, args);
    return result;
  }

  function leadingEdge(time) {
    lastInvokeTime = time;
    timeoutId = setTimeout(timerExpired, wait);
    return leading ? invokeFunc(time) : result;
  }

  function remainingWait(time) {
    const timeSinceLastCall = time - lastCallTime;
    const timeSinceLastInvoke = time - lastInvokeTime;
    const timeWaiting = wait - timeSinceLastCall;

    return maxWait !== undefined
      ? Math.min(timeWaiting, maxWait - timeSinceLastInvoke)
      : timeWaiting;
  }

  function shouldInvoke(time) {
    const timeSinceLastCall = time - lastCallTime;
    const timeSinceLastInvoke = time - lastInvokeTime;

    return (
      lastCallTime === undefined ||
      timeSinceLastCall >= wait ||
      timeSinceLastCall < 0 ||
      (maxWait !== undefined && timeSinceLastInvoke >= maxWait)
    );
  }

  function timerExpired() {
    const time = Date.now();
    if (shouldInvoke(time)) {
      return trailingEdge(time);
    }
    timeoutId = setTimeout(timerExpired, remainingWait(time));
  }

  function trailingEdge(time) {
    timeoutId = undefined;

    if (trailing && lastArgs) {
      return invokeFunc(time);
    }
    lastArgs = lastThis = undefined;
    return result;
  }

  function cancel() {
    if (timeoutId !== undefined) {
      clearTimeout(timeoutId);
    }
    abortController?.abort();
    lastInvokeTime = 0;
    lastArgs = lastCallTime = lastThis = timeoutId = undefined;
  }

  function flush() {
    return timeoutId === undefined ? result : trailingEdge(Date.now());
  }

  let lastArgs, lastThis, result;

  function debounced() {
    const time = Date.now();
    const isInvoking = shouldInvoke(time);

    lastArgs = arguments;
    lastThis = this;
    lastCallTime = time;

    if (isInvoking) {
      if (timeoutId === undefined) {
        return leadingEdge(lastCallTime);
      }
      if (maxWait) {
        timeoutId = setTimeout(timerExpired, wait);
        return invokeFunc(lastCallTime);
      }
    }
    if (timeoutId === undefined) {
      timeoutId = setTimeout(timerExpired, wait);
    }
    return result;
  }

  debounced.cancel = cancel;
  debounced.flush = flush;
  return debounced;
}

// Modern throttle with requestAnimationFrame support
export function throttle(func, wait, options = {}) {
  const { leading = true, trailing = true, raf = false } = options;
  let timeout;
  let previous = 0;
  let result;

  const throttled = function (...args) {
    const now = Date.now();
    if (!previous && !leading) previous = now;

    const remaining = wait - (now - previous);

    if (remaining <= 0 || remaining > wait) {
      if (timeout) {
        clearTimeout(timeout);
        timeout = null;
      }
      previous = now;

      if (raf) {
        return new Promise((resolve) => {
          requestAnimationFrame(() => {
            result = func.apply(this, args);
            resolve(result);
          });
        });
      } else {
        result = func.apply(this, args);
      }
    } else if (!timeout && trailing) {
      timeout = setTimeout(() => {
        previous = leading === false ? 0 : Date.now();
        timeout = null;

        if (raf) {
          requestAnimationFrame(() => {
            result = func.apply(this, args);
          });
        } else {
          result = func.apply(this, args);
        }
      }, remaining);
    }

    return result;
  };

  throttled.cancel = function () {
    clearTimeout(timeout);
    previous = 0;
    timeout = result = undefined;
  };

  return throttled;
}

// Modern image loading with intersection observer and progressive enhancement
export class ModernImageLoader {
  #observer;
  #loadedImages = new WeakSet();

  constructor(options = {}) {
    const {
      rootMargin = "50px",
      threshold = 0.1,
      placeholder = "data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iMzIwIiBoZWlnaHQ9IjE4MCIgdmlld0JveD0iMCAwIDMyMCAxODAiIGZpbGw9Im5vbmUiIHhtbG5zPSJodHRwOi8vd3d3LnczLm9yZy8yMDAwL3N2ZyI+PHJlY3Qgd2lkdGg9IjMyMCIgaGVpZ2h0PSIxODAiIGZpbGw9IiNmMGYwZjAiLz48L3N2Zz4=",
    } = options;

    this.placeholder = placeholder;

    this.#observer = new IntersectionObserver(
      (entries) => {
        entries.forEach((entry) => {
          if (entry.isIntersecting) {
            this.#loadImage(entry.target);
            this.#observer.unobserve(entry.target);
          }
        });
      },
      { rootMargin, threshold },
    );
  }

  observe(img) {
    if (!img || this.#loadedImages.has(img)) return;

    // Set up lazy loading attributes
    if (img.dataset.src) {
      img.src = img.src || this.placeholder;
      img.loading = "lazy";
      img.decoding = "async";
      this.#observer.observe(img);
    }
  }

  async #loadImage(img) {
    const src = img.dataset.src;
    if (!src) return;

    try {
      // Preload the image
      const imageLoader = new Image();
      imageLoader.decoding = "async";

      await new Promise((resolve, reject) => {
        imageLoader.onload = resolve;
        imageLoader.onerror = reject;
        imageLoader.src = src;
      });

      // Apply smooth transition
      img.style.transition = "opacity 0.3s ease";
      img.style.opacity = "0";

      await new Promise((resolve) => {
        requestAnimationFrame(() => {
          img.src = src;
          img.style.opacity = "1";
          img.classList.add("loaded");
          resolve();
        });
      });

      this.#loadedImages.add(img);
    } catch (error) {
      console.warn("Failed to load image:", src, error);
      img.classList.add("error");
    } finally {
      img.removeAttribute("data-src");
    }
  }

  disconnect() {
    this.#observer?.disconnect();
  }
}

// Modern event delegation with better performance
export class EventManager {
  #listeners = new Map();
  #abortController = new AbortController();

  on(target, event, selector, handler, options = {}) {
    const { passive = true, once = false } = options;

    const delegatedHandler = (e) => {
      const matchedElement = e.target.closest(selector);
      if (matchedElement && target.contains(matchedElement)) {
        handler.call(matchedElement, e);
      }
    };

    const eventOptions = {
      passive,
      once,
      signal: this.#abortController.signal,
    };

    target.addEventListener(event, delegatedHandler, eventOptions);

    const key = `${event}-${selector}`;
    if (!this.#listeners.has(key)) {
      this.#listeners.set(key, []);
    }
    this.#listeners.get(key).push({ handler: delegatedHandler, target });

    return this;
  }

  off(event, selector) {
    const key = `${event}-${selector}`;
    const listeners = this.#listeners.get(key);

    if (listeners) {
      listeners.forEach(({ handler, target }) => {
        target.removeEventListener(event, handler);
      });
      this.#listeners.delete(key);
    }

    return this;
  }

  destroy() {
    this.#abortController.abort();
    this.#listeners.clear();
  }
}

// Modern animation utilities with Web Animations API
export class AnimationManager {
  static fadeIn(element, options = {}) {
    const { duration = 300, easing = "ease-out", delay = 0 } = options;

    return element.animate(
      [
        { opacity: 0, transform: "translateY(20px)" },
        { opacity: 1, transform: "translateY(0)" },
      ],
      {
        duration,
        easing,
        delay,
        fill: "both",
      },
    );
  }

  static fadeOut(element, options = {}) {
    const { duration = 300, easing = "ease-in", delay = 0 } = options;

    return element.animate(
      [
        { opacity: 1, transform: "translateY(0)" },
        { opacity: 0, transform: "translateY(-20px)" },
      ],
      {
        duration,
        easing,
        delay,
        fill: "both",
      },
    );
  }

  static slideIn(element, direction = "left", options = {}) {
    const {
      duration = 400,
      easing = "cubic-bezier(0.25, 0.46, 0.45, 0.94)",
      delay = 0,
    } = options;

    const transforms = {
      left: ["translateX(-100%)", "translateX(0)"],
      right: ["translateX(100%)", "translateX(0)"],
      up: ["translateY(-100%)", "translateY(0)"],
      down: ["translateY(100%)", "translateY(0)"],
    };

    return element.animate(
      [
        { transform: transforms[direction][0], opacity: 0 },
        { transform: transforms[direction][1], opacity: 1 },
      ],
      {
        duration,
        easing,
        delay,
        fill: "both",
      },
    );
  }

  static scale(element, from = 0.8, to = 1, options = {}) {
    const {
      duration = 300,
      easing = "cubic-bezier(0.68, -0.55, 0.265, 1.55)",
      delay = 0,
    } = options;

    return element.animate(
      [
        { transform: `scale(${from})`, opacity: 0 },
        { transform: `scale(${to})`, opacity: 1 },
      ],
      {
        duration,
        easing,
        delay,
        fill: "both",
      },
    );
  }

  static staggered(elements, animation, options = {}) {
    const { stagger = 100 } = options;

    return Promise.all(
      Array.from(elements).map((element, index) => {
        const animationOptions = {
          ...options,
          delay: (options.delay || 0) + index * stagger,
        };

        return animation(element, animationOptions).finished;
      }),
    );
  }
}

// Modern local storage with compression and encryption support
export class ModernStorage {
  #prefix;
  #compress;
  #encrypt;
  #key;

  constructor(options = {}) {
    const {
      prefix = "app_",
      compress = false,
      encrypt = false,
      encryptionKey = null,
    } = options;

    this.#prefix = prefix;
    this.#compress = compress;
    this.#encrypt = encrypt;
    this.#key = encryptionKey;
  }

  async set(key, value, options = {}) {
    try {
      let data = JSON.stringify(value);

      if (this.#compress) {
        data = await this.#compressData(data);
      }

      if (this.#encrypt && this.#key) {
        data = await this.#encryptData(data);
      }

      const storageData = {
        value: data,
        timestamp: Date.now(),
        compressed: this.#compress,
        encrypted: this.#encrypt,
        expires: options.expires,
      };

      localStorage.setItem(this.#prefix + key, JSON.stringify(storageData));
      return true;
    } catch (error) {
      console.error("Storage set error:", error);
      return false;
    }
  }

  async get(key, defaultValue = null) {
    try {
      const stored = localStorage.getItem(this.#prefix + key);
      if (!stored) return defaultValue;

      const storageData = JSON.parse(stored);

      // Check expiration
      if (storageData.expires && Date.now() > storageData.expires) {
        this.remove(key);
        return defaultValue;
      }

      let data = storageData.value;

      if (storageData.encrypted && this.#key) {
        data = await this.#decryptData(data);
      }

      if (storageData.compressed) {
        data = await this.#decompressData(data);
      }

      return JSON.parse(data);
    } catch (error) {
      console.error("Storage get error:", error);
      return defaultValue;
    }
  }

  remove(key) {
    localStorage.removeItem(this.#prefix + key);
  }

  clear() {
    const keys = Object.keys(localStorage);
    keys.forEach((key) => {
      if (key.startsWith(this.#prefix)) {
        localStorage.removeItem(key);
      }
    });
  }

  async #compressData(data) {
    if ("CompressionStream" in window) {
      const stream = new CompressionStream("gzip");
      const writer = stream.writable.getWriter();
      const reader = stream.readable.getReader();

      writer.write(new TextEncoder().encode(data));
      writer.close();

      const chunks = [];
      let done = false;

      while (!done) {
        const { value, done: readerDone } = await reader.read();
        done = readerDone;
        if (value) chunks.push(value);
      }

      const compressed = new Uint8Array(
        chunks.reduce((acc, chunk) => acc + chunk.length, 0),
      );
      let offset = 0;
      for (const chunk of chunks) {
        compressed.set(chunk, offset);
        offset += chunk.length;
      }

      return btoa(String.fromCharCode(...compressed));
    }
    return data; // Fallback if compression not supported
  }

  async #decompressData(compressedData) {
    if ("DecompressionStream" in window) {
      try {
        const compressed = Uint8Array.from(atob(compressedData), (c) =>
          c.charCodeAt(0),
        );
        const stream = new DecompressionStream("gzip");
        const writer = stream.writable.getWriter();
        const reader = stream.readable.getReader();

        writer.write(compressed);
        writer.close();

        const chunks = [];
        let done = false;

        while (!done) {
          const { value, done: readerDone } = await reader.read();
          done = readerDone;
          if (value) chunks.push(value);
        }

        const decompressed = new Uint8Array(
          chunks.reduce((acc, chunk) => acc + chunk.length, 0),
        );
        let offset = 0;
        for (const chunk of chunks) {
          decompressed.set(chunk, offset);
          offset += chunk.length;
        }

        return new TextDecoder().decode(decompressed);
      } catch (error) {
        console.error("Decompression error:", error);
      }
    }
    return compressedData; // Fallback if decompression fails or not supported
  }

  async #encryptData(data) {
    if ("crypto" in window && "subtle" in crypto) {
      try {
        const encoder = new TextEncoder();
        const keyMaterial = await crypto.subtle.importKey(
          "raw",
          encoder.encode(this.#key),
          "PBKDF2",
          false,
          ["deriveBits", "deriveKey"],
        );

        const salt = crypto.getRandomValues(new Uint8Array(16));
        const key = await crypto.subtle.deriveKey(
          {
            name: "PBKDF2",
            salt: salt,
            iterations: 100000,
            hash: "SHA-256",
          },
          keyMaterial,
          { name: "AES-GCM", length: 256 },
          false,
          ["encrypt", "decrypt"],
        );

        const iv = crypto.getRandomValues(new Uint8Array(12));
        const encrypted = await crypto.subtle.encrypt(
          { name: "AES-GCM", iv: iv },
          key,
          encoder.encode(data),
        );

        const encryptedData = new Uint8Array(
          salt.length + iv.length + encrypted.byteLength,
        );
        encryptedData.set(salt, 0);
        encryptedData.set(iv, salt.length);
        encryptedData.set(new Uint8Array(encrypted), salt.length + iv.length);

        return btoa(String.fromCharCode(...encryptedData));
      } catch (error) {
        console.error("Encryption error:", error);
      }
    }
    return data; // Fallback if encryption fails or not supported
  }

  async #decryptData(encryptedData) {
    if ("crypto" in window && "subtle" in crypto) {
      try {
        const decoder = new TextDecoder();
        const encoder = new TextEncoder();
        const data = Uint8Array.from(atob(encryptedData), (c) =>
          c.charCodeAt(0),
        );

        const salt = data.slice(0, 16);
        const iv = data.slice(16, 28);
        const encrypted = data.slice(28);

        const keyMaterial = await crypto.subtle.importKey(
          "raw",
          encoder.encode(this.#key),
          "PBKDF2",
          false,
          ["deriveBits", "deriveKey"],
        );

        const key = await crypto.subtle.deriveKey(
          {
            name: "PBKDF2",
            salt: salt,
            iterations: 100000,
            hash: "SHA-256",
          },
          keyMaterial,
          { name: "AES-GCM", length: 256 },
          false,
          ["encrypt", "decrypt"],
        );

        const decrypted = await crypto.subtle.decrypt(
          { name: "AES-GCM", iv: iv },
          key,
          encrypted,
        );

        return decoder.decode(decrypted);
      } catch (error) {
        console.error("Decryption error:", error);
      }
    }
    return encryptedData; // Fallback if decryption fails or not supported
  }
}

// Modern form validation with real-time feedback
export class FormValidator {
  #form;
  #rules = new Map();
  #errors = new Map();
  #touched = new Set();

  constructor(form, options = {}) {
    this.#form = form;
    const {
      validateOnInput = true,
      validateOnBlur = true,
      showErrorsOnSubmit = true,
    } = options;

    if (validateOnInput) {
      this.#form.addEventListener("input", this.#handleInput.bind(this));
    }

    if (validateOnBlur) {
      this.#form.addEventListener("blur", this.#handleBlur.bind(this), true);
    }

    if (showErrorsOnSubmit) {
      this.#form.addEventListener("submit", this.#handleSubmit.bind(this));
    }
  }

  addRule(fieldName, validator, message) {
    if (!this.#rules.has(fieldName)) {
      this.#rules.set(fieldName, []);
    }
    this.#rules.get(fieldName).push({ validator, message });
    return this;
  }

  required(fieldName, message = "This field is required") {
    return this.addRule(
      fieldName,
      (value) => {
        return value && value.toString().trim().length > 0;
      },
      message,
    );
  }

  email(fieldName, message = "Please enter a valid email address") {
    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    return this.addRule(
      fieldName,
      (value) => {
        return !value || emailRegex.test(value);
      },
      message,
    );
  }

  minLength(
    fieldName,
    length,
    message = `Minimum length is ${length} characters`,
  ) {
    return this.addRule(
      fieldName,
      (value) => {
        return !value || value.toString().length >= length;
      },
      message,
    );
  }

  maxLength(
    fieldName,
    length,
    message = `Maximum length is ${length} characters`,
  ) {
    return this.addRule(
      fieldName,
      (value) => {
        return !value || value.toString().length <= length;
      },
      message,
    );
  }

  pattern(fieldName, regex, message = "Invalid format") {
    return this.addRule(
      fieldName,
      (value) => {
        return !value || regex.test(value);
      },
      message,
    );
  }

  custom(fieldName, validator, message) {
    return this.addRule(fieldName, validator, message);
  }

  validateField(fieldName) {
    const field = this.#form.querySelector(`[name="${fieldName}"]`);
    if (!field) return true;

    const value = field.value;
    const rules = this.#rules.get(fieldName) || [];
    const errors = [];

    for (const { validator, message } of rules) {
      if (!validator(value, field)) {
        errors.push(message);
      }
    }

    if (errors.length > 0) {
      this.#errors.set(fieldName, errors);
      this.#showFieldError(field, errors[0]);
      return false;
    } else {
      this.#errors.delete(fieldName);
      this.#clearFieldError(field);
      return true;
    }
  }

  validateAll() {
    let isValid = true;

    for (const fieldName of this.#rules.keys()) {
      if (!this.validateField(fieldName)) {
        isValid = false;
      }
    }

    return isValid;
  }

  getErrors() {
    return Object.fromEntries(this.#errors);
  }

  clearErrors() {
    this.#errors.clear();
    this.#form.querySelectorAll(".error-message").forEach((el) => el.remove());
    this.#form
      .querySelectorAll(".error")
      .forEach((el) => el.classList.remove("error"));
  }

  #handleInput(event) {
    const field = event.target;
    const fieldName = field.name;

    if (this.#touched.has(fieldName)) {
      this.validateField(fieldName);
    }
  }

  #handleBlur(event) {
    const field = event.target;
    const fieldName = field.name;

    this.#touched.add(fieldName);
    this.validateField(fieldName);
  }

  #handleSubmit(event) {
    if (!this.validateAll()) {
      event.preventDefault();
      // Focus first error field
      const firstErrorField = this.#form.querySelector(".error");
      firstErrorField?.focus();
    }
  }

  #showFieldError(field, message) {
    this.#clearFieldError(field);

    field.classList.add("error");
    field.setAttribute("aria-invalid", "true");

    const errorElement = document.createElement("div");
    errorElement.className = "error-message";
    errorElement.textContent = message;
    errorElement.setAttribute("role", "alert");
    errorElement.setAttribute("aria-live", "polite");

    field.parentNode.insertBefore(errorElement, field.nextSibling);
  }

  #clearFieldError(field) {
    field.classList.remove("error");
    field.removeAttribute("aria-invalid");

    const errorElement = field.parentNode.querySelector(".error-message");
    errorElement?.remove();
  }
}

// Modern API client with retry logic and caching
export class APIClient {
  #baseURL;
  #defaultHeaders;
  #cache = new Map();
  #interceptors = { request: [], response: [] };

  constructor(baseURL = "", options = {}) {
    this.#baseURL = baseURL.replace(/\/$/, "");
    this.#defaultHeaders = {
      "Content-Type": "application/json",
      ...options.headers,
    };
  }

  addRequestInterceptor(interceptor) {
    this.#interceptors.request.push(interceptor);
    return this;
  }

  addResponseInterceptor(interceptor) {
    this.#interceptors.response.push(interceptor);
    return this;
  }

  async get(url, options = {}) {
    return this.#request("GET", url, null, options);
  }

  async post(url, data, options = {}) {
    return this.#request("POST", url, data, options);
  }

  async put(url, data, options = {}) {
    return this.#request("PUT", url, data, options);
  }

  async delete(url, options = {}) {
    return this.#request("DELETE", url, null, options);
  }

  async #request(method, url, data, options = {}) {
    const {
      headers = {},
      cache = false,
      retry = 0,
      timeout = 10000,
      signal,
    } = options;

    const fullURL = url.startsWith("http") ? url : `${this.#baseURL}${url}`;
    const cacheKey = `${method}:${fullURL}`;

    // Check cache for GET requests
    if (method === "GET" && cache && this.#cache.has(cacheKey)) {
      const cached = this.#cache.get(cacheKey);
      if (Date.now() - cached.timestamp < (cache === true ? 300000 : cache)) {
        return cached.data;
      }
    }

    let config = {
      method,
      headers: { ...this.#defaultHeaders, ...headers },
      signal: signal || AbortSignal.timeout(timeout),
    };

    if (data) {
      config.body = typeof data === "string" ? data : JSON.stringify(data);
    }

    // Apply request interceptors
    for (const interceptor of this.#interceptors.request) {
      config = (await interceptor(config)) || config;
    }

    let attempt = 0;
    let lastError;

    while (attempt <= retry) {
      try {
        const response = await fetch(fullURL, config);

        let result = response;

        // Apply response interceptors
        for (const interceptor of this.#interceptors.response) {
          result = (await interceptor(result)) || result;
        }

        if (!result.ok) {
          throw new Error(`HTTP ${result.status}: ${result.statusText}`);
        }

        const responseData = await result.json().catch(() => result.text());

        // Cache successful GET requests
        if (method === "GET" && cache && result.ok) {
          this.#cache.set(cacheKey, {
            data: responseData,
            timestamp: Date.now(),
          });
        }

        return responseData;
      } catch (error) {
        lastError = error;
        attempt++;

        if (attempt <= retry && error.name !== "AbortError") {
          // Exponential backoff
          await new Promise((resolve) =>
            setTimeout(resolve, Math.pow(2, attempt) * 1000),
          );
        }
      }
    }

    throw lastError;
  }

  clearCache() {
    this.#cache.clear();
  }
}

// Export all utilities
// export {
//   ModernImageLoader as ImageLoader,
//   EventManager,
//   AnimationManager,
//   ModernStorage as Storage,
//   FormValidator,
//   APIClient
// };
