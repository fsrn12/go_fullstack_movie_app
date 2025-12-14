export const Auth = {
  _JWT_STORAGE_KEY: "jwt",
  _jwt: null,
  _isRefreshingToken: false,
  _refreshPromise: null,
  _user: null,

  /**
   * Initializes the Auth module by attempting to load a stored JWT from localStorage.
   * If a JWT is found, it decodes it and checks for validity/expiration.
   * Dispatches 'auth:loggedIn' or 'auth:loggedOut' events based on status.
   */
  async init() {
    // console.log("🔐 Auth Module Initializing...");
    try {
      this._jwt = await this._getStoredJwt();
      console.log(
        "Auth.init: JWT retrieved from storage:",
        !!this._jwt ? "Present" : "None",
      );

      if (this._jwt) {
        this._user = this._decodeJwt(this._jwt);
        console.log("Auth.init: Decoded User object:", this._user);

        if (this._user) {
          const isExpired = this._isTokenExpired(this._user.exp);
          const currentTimestamp = Math.floor(Date.now() / 1000);
          console.log(
            `Auth.init: Token expiration check (exp: ${this._user.exp}, current: ${currentTimestamp}): Expired? ${isExpired}`,
          );

          if (isExpired) {
            console.warn(
              "Auth.init: Stored JWT is expired. Clearing and forcing re-login/refresh on next protected request.",
            );
            this._jwt = null;
            this._user = null;
            // Optionally dispatch a logout event here too if it was an active session
            // document.dispatchEvent(new CustomEvent("auth:loggedOut"));
          } else {
            console.log(
              "✅ Auth.init: JWT loaded and valid. User:",
              this._user?.email || this._user?.name || "unknown",
            );
            document.dispatchEvent(
              new CustomEvent("auth:loggedIn", { detail: this._user }),
            );
          }
        } else {
          console.error(
            "Auth.init: JWT found but could not be decoded. Clearing it.",
          );
          this.clearJwt(); // Clear potentially malformed JWT
        }
      } else {
        console.log("ℹ️ Auth.init: No JWT found in storage on init.");
      }
      this._setupTokenRefreshListener(); // Always set up listener regardless of initial JWT presence
      console.log("Auth.init: Final isLoggedIn state:", this.isLoggedIn());
    } catch (error) {
      console.error("❌ Auth initialization failed:", error);
      this.clearJwt();
      // Using window.app.showError assuming main.js has already initialized window.app
      if (window.app && typeof window.app.showError === "function") {
        window.app.showError(
          "Authentication system failed to initialize. Please try logging in again.",
        );
      } else {
        console.error(
          "window.app.showError not available to display auth init error.",
        );
      }
    }
  },

  /**
   * Sets the JWT token after successful login/registration.
   * Stores it in localStorage and decodes user information.
   * @param {string} token - The JWT string.
   */
  async setJwt(token) {
    if (!token || typeof token !== "string") {
      console.error(
        "Attempted to set an invalid JWT (null, undefined, or not string).",
      );
      return;
    }
    this._jwt = token;
    try {
      await localStorage.setItem(this._JWT_STORAGE_KEY, token);
      this._user = this._decodeJwt(token);
      console.log(
        "✅ Auth.setJwt: JWT set and stored. Decoded User:",
        this._user?.email || this._user?.name || "unknown",
      );
      document.dispatchEvent(
        new CustomEvent("auth:loggedIn", { detail: this._user }),
      );
    } catch (error) {
      console.error("Failed to store JWT in localStorage:", error);
      if (window.app && typeof window.app.showError === "function") {
        window.app.showError(
          "Could not securely store your session. Please try again.",
        );
      }
    }
  },

  /**
   * Sets the User after successful login/registration.
   * Stores it in localStorage.
   * @param {object} user - The User object in response body.
   */
  setUser(data) {
    this._user = Object.assign(this._user, data);
  },

  /**
   * Retrieves the current JWT token.
   * @returns {string|null} The JWT token or null if not available.
   */
  getJwt() {
    return this._jwt;
  },

  /**
   * Checks if the user is currently logged in and their token is valid/non-expired.
   * @returns {boolean} True if logged in, false otherwise.
   */
  isLoggedIn() {
    const loggedIn = !!this._jwt && !this._isTokenExpired(this._user?.exp);
    // console.log(`Auth.isLoggedIn(): Current JWT present: ${!!this._jwt}, Token expired: ${this._isTokenExpired(this._user?.exp)}. Result: ${loggedIn}`);
    return loggedIn;
  },

  /**
   * Gets the decoded user object.
   * @returns {object|null} The user object or null if not available.
   */
  getUser() {
    return this._user;
  },

  /**
   * Clears the JWT token from memory and localStorage, effectively logging out the user.
   */
  async clearJwt() {
    console.log("Auth.clearJwt: Clearing JWT.");
    this._jwt = null;
    this._user = null;
    try {
      await localStorage.removeItem(this._JWT_STORAGE_KEY);
      console.log("🗑️ Auth.clearJwt: JWT cleared from storage.");
      document.dispatchEvent(new CustomEvent("auth:loggedOut"));
    } catch (error) {
      console.error("Failed to clear JWT from localStorage:", error);
    }
  },

  /**
   * Decodes a JWT token string into its payload object.
   * This function assumes a standard JWT structure (header.payload.signature).
   * It is crucial that the 'exp' (expiration), 'user_id', 'name', and 'email' claims match your backend's JWT structure.
   * @param {string} token - The JWT string to decode.
   * @returns {object|null} The decoded payload as an object, or null if decoding fails.
   */
  _decodeJwt(token) {
    if (!token) {
      console.warn("Auth._decodeJwt: Attempted to decode a null or empty JWT.");
      return null;
    }
    try {
      const base64Url = token.split(".")[1];
      if (!base64Url) {
        console.error(
          "Auth._decodeJwt: JWT token is not in expected format (missing payload part).",
        );
        return null;
      }
      const base64 = base64Url.replace(/-/g, "+").replace(/_/g, "/");
      const jsonPayload = decodeURIComponent(
        atob(base64)
          .split("")
          .map(function (c) {
            return "%" + ("00" + c.charCodeAt(0).toString(16)).slice(-2);
          })
          .join(""),
      );
      const payload = JSON.parse(jsonPayload);
      console.log("Auth._decodeJwt: Parsed JWT Payload (raw):", payload);

      // ✅ CRUCIAL MAPPING: Map backend claims to frontend's expected user object structure
      return {
        id: payload.user_id, // Maps backend's 'user_id' claim
        name: payload.name, // Maps backend's 'name' claim
        email: payload.email, // Maps backend's 'email' claim
        exp: payload.exp, // Maps backend's 'exp' (expiration) claim
      };
    } catch (error) {
      console.error(
        "Auth._decodeJwt: Error decoding or parsing JWT payload:",
        error,
      );
      return null;
    }
  },

  /**
   * Retrieves the JWT token from localStorage.
   * @returns {string|null} The JWT token string or null if not found.
   */
  async _getStoredJwt() {
    try {
      return await localStorage.getItem(this._JWT_STORAGE_KEY);
    } catch (error) {
      console.error("Error retrieving JWT from localStorage:", error);
      return null;
    }
  },

  /**
   * Checks if a token's expiration timestamp (exp) is in the past.
   * @param {number} exp - The expiration timestamp (Unix seconds) from the JWT payload.
   * @returns {boolean} True if the token is expired, false otherwise.
   */
  _isTokenExpired(exp) {
    if (!exp) {
      console.warn(
        "Auth._isTokenExpired: No 'exp' (expiration) claim found in token. Considering token expired/invalid.",
      );
      return true; // No expiration means it's invalid or considered expired for safety
    }
    const now = Date.now() / 1000; // Current time in seconds since epoch
    const expired = exp < now;
    // console.log(`_isTokenExpired check: exp=${exp}, now=${now}, expired=${expired}`); // Uncomment for very granular debugging
    return expired;
  },

  /**
   * Sets up an event listener to automatically refresh the access token
   * when an 'auth:retry-request' event is dispatched (e.g., by the API service).
   */
  _setupTokenRefreshListener() {
    document.addEventListener("auth:retry-request", async (event) => {
      console.log(
        "Auth: Caught 'auth:retry-request' event. Attempting token refresh...",
      );
      if (!this._isRefreshingToken) {
        this._isRefreshingToken = true;
        this._refreshPromise = this.refreshAccessToken();
      }
      try {
        await this._refreshPromise;
        event.detail.retry(); // Retry the original failed request
        console.log(
          "Auth: Token refreshed successfully and original request retried.",
        );
      } catch (error) {
        console.error("Auth: Token refresh failed.", error);
        this.clearJwt(); // Clear token on refresh failure
        // Redirect to login only if a specific login route is expected globally.
        // The router will typically handle protected route access for non-logged-in users.
        if (window.app && window.app.Router) {
          window.app.Router.go("/account/login");
        }
        window.app.showError("Your session has expired. Please log in again.");
      } finally {
        this._isRefreshingToken = false;
        this._refreshPromise = null;
      }
    });
  },

  /**
   * Attempts to refresh the access token using the refresh token (sent via HTTP-only cookie).
   * Dispatches 'auth:loggedIn' on success, 'auth:loggedOut' on failure.
   */
  async refreshAccessToken() {
    console.log("Auth: Sending refresh token request...");
    try {
      // Assuming your API service has a 'refreshToken' method that hits your backend's refresh endpoint
      const response = await window.app.API.refreshToken();
      if (response && response.success && response.jwt) {
        console.log("Auth: Token refresh successful. New JWT received.");
        await this.setJwt(response.jwt); // Store the new JWT
        return true;
      } else {
        console.warn("Auth: Token refresh failed or no new JWT received.");
        this.clearJwt();
        throw new Error(
          "Failed to refresh token: " + (response?.message || "Unknown reason"),
        );
      }
    } catch (error) {
      console.error("Auth: Error during refresh token API call:", error);
      this.clearJwt();
      throw error;
    }
  },

  /**
   * Starts a proactive token refresh interval before the access token expires.
   * This is optional and depends on whether your backend supports refresh token rotation/single use.
   * Currently, this is not actively called in your app, but kept for completeness.
   */
  startProactiveRefresh() {
    // This is more advanced and typically involves setting a timeout
    // based on the token's expiration time (e.g., refresh 5 minutes before expiry).
    // Not strictly necessary if you rely on the 'auth:retry-request' reactive refresh.
    console.warn(
      "Proactive token refresh not fully implemented/active in this version.",
    );
  },
};
