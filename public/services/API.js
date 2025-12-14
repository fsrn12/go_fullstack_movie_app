const API = {
  baseURL: "/api/",

  /*
   Event listener for retried requests from the Auth module
   This listener must be set up globally, ideally in main.js or a central place
   that runs before any API calls.
  */
  _setupRetryRequestListener: false,

  init() {
    if (!this._setupRetryRequestListener) {
      document.addEventListener("auth:retry-request", async (event) => {
        const { request, resolve, reject } = event.detail; // Expecting a Request object
        try {
          // Re-execute the request object (which already has the updated JWT header)
          const res = await fetch(request);
          if (!res.ok) {
            const errorData = await res.json().catch(() => ({}));
            reject(new Error(errorData.message || "Retried request failed."));
            return;
          }
          const data = await res.json();
          resolve(data);
        } catch (error) {
          reject(error);
        }
      });
      this._setupRetryRequestListener = true; // Corrected: Using '_'
    }
  },

  /**
   * Generic fetch wrapper with Authorization header and error handling,
   * including JWT expiration and refresh logic.
   * @param {string} service - The API endpoint path.
   * @param {Object} [args] - Query parameters for GET requests.
   * @param {Object} [options] - Additional fetch options (e.g., method, body for send).
   * @returns {Promise<Object>} The JSON response from the API.
   */
  async _request(service, args, options = {}) {
    const token = window.app?.Auth.getJwt(); // Get JWT from Auth module
    const queryString = args ? new URLSearchParams(args).toString() : "";
    const url = API.baseURL + service + (queryString ? `?${queryString}` : "");

    const headers = {
      ...options.headers, // Allow overriding headers
    };

    if (!(options.body instanceof FormData)) {
      headers["Content-Type"] = "application/json";
    }
    if (token) {
      headers["Authorization"] = `Bearer ${token}`; // Add JWT to Authorization header
    }

    try {
      console.log(`📡 API Request: ${options.method || "GET"} ${url}`);
      // Create a Request object to ensure it's clonable for retry
      const requestToFetch = new Request(url, {
        method: options.method || "GET",
        headers: new Headers(headers), // Use Headers object for better handling
        body: options.body, // Ensure body is included for POST/PUT
        credentials: options.credentials, // default to 'omit' unless specified
        cache: options.cache,
        mode: options.mode,
        referrer: options.referrer,
        integrity: options.integrity,
        keepalive: options.keepalive,
        redirect: options.redirect,
      });

      const response = await fetch(requestToFetch);

      // Handle 401 Unauthorized errors
      if (response.status === 401) {
        console.warn("401 Unauthorized. Attempting to refresh token...");
        // Initiate the token refresh flow
        // We'll create a promise that resolves when the retry is successful or fails
        return new Promise((resolve, reject) => {
          // Dispatch a custom event to the Auth module to trigger refresh
          // and provide the original request details (the Request object) for retry
          document.dispatchEvent(
            new CustomEvent("auth:refresh-token", {
              detail: {
                originalRequest: requestToFetch.clone(), // Clone the request so it can be re-used
                resolve, // Pass resolve/reject to allow Auth module to control this promise
                reject,
              },
            }),
          );
        });
      }

      // Handle other non-OK responses
      if (!response.ok) {
        const errorData = await response.json().catch(() => ({})); // Try parsing JSON, fallback to empty object
        const errorMessage =
          errorData.message ||
          `API Error: ${response.status} ${response.statusText}`;
        console.error(
          `❌ API Error: ${response.status} - ${errorMessage}`,
          errorData,
        );
        window.app.showError(errorMessage);
        throw new Error(errorMessage);
      }

      const result = await response.json();
      console.log(`✅ API Response for ${service}:`, result);
      return result;
    } catch (err) {
      console.error("API._request error:", err);
      // If the error originated from the token refresh or if it's a network error
      if (
        err.message.includes("Failed to refresh token") ||
        err.message.includes("network")
      ) {
        // Auth module will typically handle redirection for failed refresh
      } else {
        window.app.showError(
          err.message ||
            "An unexpected network error occurred. Please try again.",
        );
      }
      throw err; // Re-throw the error to propagate it to the caller
    }
  },

  // ------------------------------------------------------------
  // Public API Methods using the internal _request helper
  // ------------------------------------------------------------

  getTopMovies: async () => {
    return await API._request("movies/top");
  },

  getRandomMovies: async () => {
    return await API._request("movies/random");
  },

  getMovieByID: async (id) => {
    return await API._request(`movies/${id}`);
  },

  searchMovies: async (q, order, genre) => {
    return await API._request("movies/search", { q, order, genre });
  },

  getGenres: async () => {
    return await API._request(`genres/`);
  },

  getFavorites: async () => {
    return await API._request("account/favorites");
  },

  getWatchlist: async () => {
    return await API._request("account/watchlist");
  },

  saveToCollection: async (movie_id, collection) => {
    return await API._request("account/save-to-collection", null, {
      method: "POST",
      body: JSON.stringify({ movie_id, collection }),
    });
  },

  removeFromCollection: async (movie_id, collection) => {
    return await API._request("account/remove-from-collection", null, {
      method: "POST",
      body: JSON.stringify({ movie_id, collection }),
    });
  },

  register: async (name, email, password) => {
    return await API._request("account/register", null, {
      method: "POST",
      body: JSON.stringify({ name, email, password }),
    });
  },

  login: async (email, password) => {
    return await API._request("account/login", null, {
      method: "POST",
      body: JSON.stringify({ email, password }),
    });
  },

  logout: async () => {
    // Backend needs to invalidate the refresh token in the HTTP-only cookie
    // Send an empty POST request or a specific logout payload
    return await API._request("account/logout", null, {
      method: "POST",
      // No explicit body needed usually, just the cookie for the backend to invalidate.
      credentials: "include", // must ensure refresh token is sent to backend for invalidation
    });
  },

  /**
   * Uploads a profile picture to the backend.
   * @param {File} file - The image file to upload.
   * @returns {Promise<object>} The response from the server, including the new picture URL.
   */
  async uploadProfilePicture(file) {
    if (!(file instanceof File)) {
      console.error("uploadProfilePicture: Invalid file provided.");
      throw new Error("Invalid file for upload.");
    }

    const formData = new FormData();
    formData.append("profilePicture", file);

    console.log(
      "📡 API Request: POST /api/account/profile-picture (File Upload)",
    );
    try {
      const response = await this._request("account/profile-picture", null, {
        method: "POST",
        body: formData, // FormData is automatically handled for content-type: multipart/form-data
      });
      console.log("✅ API Response for profile picture upload:", response);
      return response;
      0;
    } catch (error) {
      console.error("❌ API Error during profile picture upload:", error);
      throw error;
    }
  },

  /**
   * Fetches the user's full profile, which might include the profile picture URL.
   * This is an optional method. Can/should include profilePictureUrl in the JWT
   * or fetch it separately after login. This provides a clear way to get it.
   * @returns {Promise<object>} The user's profile data.
   */
  async fetchUserProfile() {
    console.log("📡 API Request: GET /api/account/profile");
    try {
      const response = await this._request("account/profile", null, {
        method: "GET",
      });
      console.log("✅ API Response for user profile:", response);
      return response;
    } catch (error) {
      console.error("❌ API Error during user profile fetch:", error);
      throw error;
    }
  },
};

export default API;
// Initialize the API module to set up event listeners
// It's good practice to call this from main.js DOMContentLoaded
// API.init(); // Remove this line from here if called from main.js

// import { Auth } from "../auth.js";

// const API = {
//   baseURL: "/api/",
//   getTopMovies: async () => {
//     return await API.fetch("movies/top");
//     // const res = await fetch("movies/top");
//     // if (!res.ok) throw new Error("Failed to fetch top movies");
//     // const data = await res.json();
//     // console.log("API getTopMovies →", data);
//     // return data;
//   },

//   getRandomMovies: async () => {
//     return await API.fetch("movies/random");
//     // const res = await fetch("movies/random");
//     // if (!res.ok) throw new Error("Failed to fetch Random movies");
//     // const data = await res.json();
//     // console.log("API getRandomMovies →", data);
//     // return data;
//   },

//   getMovieByID: async (id) => {
//     let res = await API.fetch(`movies/${id}`);
//     return res.data;
//   },

//   searchMovies: async (q, order, genre) => {
//     return await API.fetch("movies/search", { q, order, genre });
//   },

//   getGenres: async () => {
//     return await API.fetch(`genres/`);
//   },

//   getFavorites: async () => {
//     return await API.fetch("account/favorites");
//   },

//   getWatchlist: async () => {
//     return await API.fetch("account/watchlist");
//   },

//   saveToCollection: async (movie_id, collection) => {
//     return await API.send("account/save-to-collection", {
//       movie_id,
//       collection,
//     });
//   },

//   register: async (name, email, password) => {
//     return await API.send("account/register", { name, email, password });
//   },

//   login: async (email, password) => {
//     return await API.send("account/login", { email, password });
//   },

//   logout: () => {
//     Auth.clearJwt();
//     app.Router.go("/");
//   },

//   send: async (service, data) => {
//     try {
//       // const token = window.app?.Auth.getJwt();

//       const res = await fetch(API.baseURL + service, {
//         method: "POST",
//         headers: {
//           "Content-Type": "application/json",
//           // ...(token && { Authorization: `Bearer ${token}` }),
//         },
//         body: JSON.stringify(data),
//       });
//       const result = await res.json();
//       console.log("Send Result: ", result); //TODO REMOVE
//       return result;
//     } catch (err) {
//       console.error(err);
//     }
//   },

//   fetch: async (service, args) => {
//     try {
//       // const token = window.app?.Auth.getJwt();
//       const queryString = args ? new URLSearchParams(args).toString() : "";
//       const url = API.baseURL + service + "?" + queryString;
//       console.table("URL", url);
//       const response = await fetch(url, {
//         headers: {
//           "Content-Type": "application/json",
//           // ...(token && { Authorization: `Bearer ${token}` }),
//         },
//       });

//       const result = await response.json();
//       console.log("📨 Fetch Result: ", result);
//       return result;
//     } catch (err) {
//       console.error("API.fetch error:", err);
//       app.showError(err, false);
//       throw err;
//     }
//   },
// };

// export default API;

// const API = {
//   baseURL: "/api/",
//   getTopMovies: async () => {
//     return await API.fetch("movies/top");
//   },

//   getRandomMovies: async () => {
//     return await API.fetch("movies/random");
//   },

//   getMovieByID: async (id) => {
//     let res = await API.fetch(`movies/${id}`);
//     return res.data;
//   },

//   searchMovies: async (q, order, genre) => {
//     return await API.fetch("movies/search", { q, order, genre });
//   },

//   getGenres: async () => {
//     return await API.fetch(`genres/`);
//   },

//   getFavorites: async () => {
//     return await API.fetch("account/favorites");
//   },

//   getWatchlist: async () => {
//     return await API.fetch("account/watchlist");
//   },

//   saveToCollection: async (movie_id, collection) => {
//     return await API.send("account/save-to-collection", {
//       movie_id,
//       collection,
//     });
//   },

//   register: async (name, email, password) => {
//     return await API.send("account/register", { name, email, password });
//   },

//   login: async (email, password) => {
//     return await API.send("account/login", { email, password });
//   },

//   logout: () => {
//     Auth.clearJwt();
//     app.Router.go("/");
//   },

//   send: async (service, data) => {
//     try {
//       // const token = Auth.getJwt();

//       const res = await fetch(API.baseURL + service, {
//         method: "POST",
//         headers: {
//           "Content-Type": "application/json",
//           // ...(token && { Authorization: `Bearer ${token}` }),
//           // Authorization: app.Store.jwt ? `Bearer ${app.Store.jwt}` : null,
//         },
//         body: JSON.stringify(data),
//       });
//       const result = await res.json();
//       console.log("Send Result: ", result); //TODO REMOVE
//       return result;
//     } catch (err) {
//       console.error(err);
//     }
//   },

//   fetch: async (service, args) => {
//     try {
//       // const token = Auth.getJwt();
//       const queryString = args ? new URLSearchParams(args).toString() : "";
//       const url = API.baseURL + service + "?" + queryString;
//       console.table("URL", url);
//       const response = await fetch(url, {
//         headers: {
//           "Content-Type": "application/json",
//           // ...(token && { Authorization: `Bearer ${token}` }),
//           // Authorization: app.Store.jwt ? `Bearer ${app.Store.jwt}` : null,
//         },
//       });

//       const result = await response.json();
//       return result;
//       // console.log("Response Status:", response.status);
//       // console.log(
//       //   "Response Content-Type:",
//       //   response.headers.get("Content-Type"),
//       // );
//       // console.log("Response: ", response); //TODO REMOVE
//       // Check if response is okay (status 2xx)
//       // if (!response.ok) {
//       //   console.error(
//       //     "Server responded with an error: ",
//       //     response.status,
//       //     response.statusText,
//       //   );
//       //   throw new Error("Server Error");
//       // }

//       // Check if the response is JSON
//       // const contentType = response.headers.get("Content-Type");
//       // if (contentType && contentType.includes("application/json")) {
//       //   const result = await response.json();
//       //   console.log("data: ", result); //TODO REMOVE
//       //   return result;
//       // } else {
//       //   const text = await response.text(); // Get raw response text
//       //   console.error("Unexpected response format:", text);
//       //   throw new Error("Unexpected response format");
//       // }
//       // return result;
//     } catch (err) {
//       console.error("API.fetch error:", err);
//       app.showError(err, false);
//       throw err;
//     }
//   },
// };

// export default API;
