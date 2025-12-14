import "./components/AccountPage.js";
import "./components/FavoritePage.js";
import "./components/HomePage.js";
import "./components/LoginPage.js";
// import "./components/MovieItem.js";
import "./components/MoviesPage.js";
import "./components/RegisterPage.js";
import "./components/WatchlistPage.js";
import API from "./services/API.js";
import { Auth } from "./services/Auth.js"; // ✨ Import Auth module
import { Router } from "./services/Router.js";
import "./serviceWorker.js";
import { showToast } from "./utils/toast.js";
import { getElement } from "./utils/util.js";
import {
  validateEmail,
  validateName,
  validatePassword,
  validatePasswordConfirmation,
  validatePasswordForLogin,
} from "./validation.js";

console.log("Main.js Connected ✅");

window.app = {
  Auth,
  API,
  Router,

  closeError: () => {
    getElement("#alert-modal").close();
  },

  showError: (message = "There was an error.", goToHome = false) => {
    getElement("#alert-modal").showModal();
    getElement("#alert-modal p").textContent = message;
    window.app.showAlert("Error!", message, "error");
    if (goToHome) app.Router.go("/");
  },

  // ------------------------
  // 🛂 REGISTER
  // ------------------------
  register: async (event) => {
    event.preventDefault();

    const formData = new FormData(event.target);
    const name = formData.get("name")?.trim() ?? "";
    const email = formData.get("email")?.trim().toLowerCase() ?? "";
    const password = formData.get("password") ?? "";
    const passwordConfirmation = formData.get("password-confirmation") ?? "";

    const errors = [
      validateName(name),
      validateEmail(email),
      validatePassword(password),
      validatePasswordConfirmation(password, passwordConfirmation),
    ].filter(Boolean); // remove nulls

    if (errors.length > 0) {
      return app.showError(errors.join(". "));
    }

    try {
      const response = await window.app?.API?.register(name, email, password);
      console.table(response); //TODO REMOVE
      if (response?.data?.success && response.data.jwt) {
        await window.app?.Auth.setJwt(response.data.jwt); // ✨ Store JWT after successful registration
        app.Router.go("/");
        console.log("Registered 😁"); //TODO REMOVE
      } else {
        app.showError(
          response?.data?.message ?? "Registration failed. Please try again.",
        );
      }
    } catch (error) {
      console.error("Registration error:", error);
      app.showError("A network or server error occurred. Please try again.");
    }
  },

  // ------------------------
  // 🚪 LOGIN
  // ------------------------
  login: async (event) => {
    event.preventDefault();

    const formData = new FormData(event.target);
    const email = formData.get("login-email")?.trim().toLowerCase() ?? "";
    const password = formData.get("login-password")?.trim() ?? "";

    const errors = [
      validateEmail(email),
      validatePasswordForLogin(password),
    ].filter(Boolean); // removes nulls

    if (errors.length > 0) {
      return app.showError(errors.join(". "));
    }

    try {
      const response = await app.API.login(email, password);
      console.log("response: ", response); //TODO REMOVE

      if (response?.data?.success && response.data.jwt) {
        await window.app?.Auth.setJwt(response.data.jwt); // ✨ Store JWT after successful login
        window.app.Auth.setUser(response.data.user);
        app.Router.go("/");
      } else {
        app.showError(response?.message ?? "Login failed");
      }
    } catch (error) {
      console.error("Login error:", error);
      app.showError("A network or server error occurred. Please try again.");
    }
  },

  // ------------------------
  // 🧹 LOGOUT
  // ------------------------
  logout: async () => {
    try {
      // Notify backend to invalidate refresh token
      await app.API.logout(); // ✨ Call API logout
      await app.Auth.clearJwt(); // ✨ Clear client-side JWT
      app.Router.go("/");
      console.log("Logged out 👋");
    } catch (error) {
      console.error("Logout error:", error);
      app.showError("Failed to log out cleanly. Please try again.");
      await app.Auth.clearJwt(); // Still clear client-side JWT even if backend fails
      app.Router.go("/");
    }
  },

  // ------------------------
  // 🔑 PASSKEY
  // ------------------------

  addNewPasskey: async () => {
    const username = "testUser";
    await Passkeys.register(username);
  },

  loginWithPasskey: async () => {
    const username = document.querySelector("#login-email").value;
    if (username.length < 4) {
      app.showError("To use a passkey, enter your email address first");
    } else {
      await Passkeys.authenticate(username);
    }
  },

  // ------------------------
  // 🔍 SEARCH
  // ------------------------
  search: (event) => {
    event.preventDefault();
    event.stopPropagation();
    const q = getElement("input[type=search]").value;
    app.Router.go("/movies?q=" + q);
  },

  searchOrderChange: (order) => {
    const urlParams = new URLSearchParams(window.location.search);
    const q = urlParams.get("q");
    const genre = urlParams.get("genre") ?? "";
    app.Router.go(`/movies?q=${q}&order=${order}&genre=${genre}`);
  },

  searchFilterChange: (genre) => {
    const urlParams = new URLSearchParams(window.location.search);
    const q = urlParams.get("q");
    const order = urlParams.get("order") ?? "";
    app.Router.go(`/movies?q=${q}&order=${order}&genre=${genre}`);
  },

  // ------------------------
  // 🧰 SAVE TO COLLECTION
  // ------------------------
  saveToCollection: async (movie_id, collection) => {
    if (app.Auth.isLoggedIn()) {
      // ✨ Check Auth.isLoggedIn()
      try {
        const response = await API.saveToCollection(movie_id, collection);
        return response;
      } catch (err) {
        console.error(err);
        app.showError(
          "An unexpected error occurred while saving to collection.",
        );
      }
    } else {
      app.showError(
        "You need to be logged in to save to your collection.",
        true,
      );
      // Redirect to home if not logged in
      app.Router.go("/account/login");
      throw new Error("Not logged in");
    }
  },

  // -----------------------------
  // 🗑️ REMOVE FROM COLLECTION
  // -----------------------------
  removeFromCollection: async (movie_id, collection) => {
    if (app.Auth.isLoggedIn()) {
      // ✨ Check Auth.isLoggedIn()
      try {
        const response = await API.removeFromCollection(movie_id, collection);
        if (!response.data.success) {
          app.showError("Could not remove movie from collection");
          return;
        }
        switch (collection) {
          case "favorite":
            app.Router.go("/account/favorites");
            break;
          case "watchlist":
            app.Router.go("/account/watchlist");
            break;
          default:
            app.Router.go("/");
        }
      } catch (err) {
        console.error(err);
        app.showError(
          "An unexpected error occurred while removing movie from collection.",
        );
      }
    } else {
      app.showError("You need to be logged in to edit your collection.", true);
      // Redirect to home if not logged in
      app.Router.go("/account/login");
      throw new Error("Not logged in");
    }
  },

  //---------------------------------------------
  // 💈 --- UI/Global Utility Functions ---
  //---------------------------------------------

  // Function to show the NEW global loading spinner with message
  showLoading: (message = "Loading...") => {
    const loadingOverlay = getElement("#global-loading-overlay");
    const loadingMessage = getElement("#global-loading-message");
    if (loadingOverlay && loadingMessage) {
      loadingMessage.textContent = message;
      loadingOverlay.style.display = "flex"; // Ensure it's 'flex' for vertical centering
    } else {
      console.warn(
        "Global loading overlay or message element not found for showLoading.",
      );
    }
  },

  // Function to hide the NEW global loading spinner
  hideLoading: () => {
    const loadingOverlay = getElement("#global-loading-overlay");
    if (loadingOverlay) {
      loadingOverlay.style.display = "none";
    }
  },

  showAlert: (title, message, type = "error") => {
    const alertModal = getElement("#alertModal");
    const alertModalTitle = getElement("#alertModalTitle");
    const alertModalMessage = getElement("#alertModalMessage");
    const alertModalCloseBtn = getElement("#alertModalCloseBtn");

    if (
      alertModal &&
      alertModalTitle &&
      alertModalMessage &&
      alertModalCloseBtn
    ) {
      alertModalTitle.textContent = title;
      alertModalMessage.textContent = message;

      if (type === "error") {
        alertModalTitle.style.color = "hsla(0, 65%, 51%, 1.00)";
      } else if (type === "success") {
        alertModalTitle.style.color = "#4caf50";
      } else {
        alertModalTitle.style.color = "#333";
      }

      alertModal.style.display = "flex";
      alertModalCloseBtn.onclick = () => {
        alertModal.style.display = "none";
      };
    } else {
      console.error(
        "Alert modal elements (ID: alertModal, alertModalTitle, etc.) not found in DOM. Falling back to native alert.",
      );
      alert(`${title}: ${message}`);
    }
  },

  showSuccess: (message) => {
    window.app.showAlert("Success!", message, "success");
  },
  showMessage: (message) => {
    window.app.showAlert("Information", message, "info");
  },
  showToast: showToast,

  // =================================
  // 📢 STATE MANAGEMENT
  // =================================
  eventBus: new Map(),
  // Publisher function to dispatch events
  publish: (eventName, data) => {
    if (window.app.eventBus.has(eventName)) {
      window.app.eventBus.get(eventName).forEach((callback) => callback(data));
    }
  },

  // Subscriber function to listen for events
  subscribe: (eventName, callback) => {
    if (!window.app.eventBus.has(eventName)) {
      window.app.eventBus.set(eventName, new Set());
    }
    window.app.eventBus.get(eventName).add(callback);
  },

  // Unsubscriber function for cleanup
  unsubscribe: (eventName, callback) => {
    if (window.app.eventBus.has(eventName)) {
      window.app.eventBus.get(eventName).delete(callback);
    }
  },
};

document.addEventListener("DOMContentLoaded", async () => {
  try {
    await window.app.Auth.init();
    window.app.Router.init();

    if ("serviceWorker" in navigator) {
      try {
        const registration = await navigator.serviceWorker.register(
          "./serviceWorker.js",
          {
            scope: "./",
          },
        );
        console.log("🛠️ Service worker registered", registration);
      } catch (error) {
        console.error("Service worker registration failed", error);
      }
    }
  } catch (error) {
    console.error("Error during application initialization:", error);
    window.app.showError("Failed to initialize application.");
  } finally {
    window.app.hideLoading(); // ***THIS IS CRITICAL*** Hide loader after ALL init operations
  }
});

// saveToCollection: async (movie_id, collection) => {
//   console.log("Inside SaveToCollection: ", collection);
//   if (app.Auth.isLoggedIn()) {
//     // ✨ Check Auth.isLoggedIn()
//     try {
//       const response = await API.saveToCollection(movie_id, collection);
//       if (!response.data.success) {
//         app.showError("Could not save movie to collection");
//         return;
//       }
//       switch (collection) {
//         case "favorite":
//           console.log("Going to Favorites Page 🚗");
//           app.Router.go("/account/favorites");
//           break;
//         case "watchlist":
//           app.Router.go("/account/watchlist");
//           break;
//         default:
//           app.Router.go("/");
//       }
//     } catch (err) {
//       console.error(err);
//       app.showError(
//         "An unexpected error occurred while saving to collection.",
//       );
//     }
//   } else {
//     app.showError(
//       "You need to be logged in to save to your collection.",
//       true,
//     );
//     // Redirect to home if not logged in
//     app.Router.go("/account/login");
//   }
// },

// window.addEventListener("DOMContentLoaded", async () => {
//   console.log("⏳ DOMContentLoaded");
//   try {
//     await app.Auth.init(); // ✨ Initialize Auth module
//     console.log("🔐 JWT Loaded:", app.Auth.getJwt() ? "Yes" : "No"); // More descriptive logging
//     console.log("✅ loggedIn:", app.Auth.isLoggedIn());

//     app.Router.init();
//     console.log("🚦 Router initialized");

//     navigator.serviceWorker.register("/serviceWorker.js");
//     console.log("🛠️ Service worker registered");

//     // Optional: Start proactive refresh if desired (otherwise rely on 401 interceptor)
//     // app.Auth.startProactiveRefresh();
//   } catch (err) {
//     console.error("❌ Startup failed:", err);
//     app.showError("App failed to start. Please reload.");
//   }
// });
