import { getElements, handleTransition, loadComponent } from "../utils/util.js";
import { routes } from "./Routes.js"; // Make sure this path is correct

export const Router = {
  /**
   * Initializes the router, setting up event listeners for browser history changes
   * and intercepting clicks on navigation links.
   */
  init() {
    // Listen for popstate events (browser back/forward buttons)
    window.addEventListener("popstate", () => {
      // For popstate, ensure we pass the full path including search, but don't add to history again
      this.go(location.pathname + location.search, false);
      console.log("📍 Router loaded. Current path:", location.pathname);
    });

    // Intercept clicks on 'a.navlink' elements for client-side routing
    interceptLinks(this);

    // Initial load: Go to the current URL.
    // Always add to history as it's the first page visited in this session (unless explicitly told not to).
    // The history API itself is smart enough not to add duplicate entries for the initial load.
    this.go(location.pathname + location.search, true);
  },

  /**
   * Navigates to a new route, updating browser history and loading the corresponding component.
   * Handles protected routes by redirecting to login if the user is not authenticated.
   * @param {string} route - The target route path (e.g., "/account", "/movies/123").
   * @param {boolean} [addToHistory=true] - Whether to add the route to the browser history.
   */
  async go(route, addToHistory = true) {
    console.log("Navigating to:", route);

    const matched = matchRoute(route, routes);

    // Case 1: No route matched at all. Display 404.
    if (!matched) {
      const notFoundPage = document.createElement("h1");
      notFoundPage.textContent = "Page not found";
      console.warn(`Router: No route matched for '${route}'. Displaying 404.`);
      // Update history for the 404 page, ensuring its URL is reflected
      updateHistory(route, addToHistory); // Keep the requested invalid URL in history for 404
      return handleTransition(notFoundPage);
    }

    const { route: matchedRoute, params } = matched;

    // Case 2: Route matched, but it's a protected route and the user is not logged in.
    // In this scenario, we must redirect to the login page immediately.
    if (matchedRoute.loggedIn && !window.app.Auth.isLoggedIn()) {
      console.warn(
        `Router: Attempted to access protected route '${route}'. User not logged in. Redirecting to /account/login.`,
      );

      // Find the login route component. It's essential that '/account/login' is defined in your Routes.js
      const loginRoute = routes.find((r) => r.path === "/account/login");
      if (!loginRoute) {
        console.error(
          "Router: /account/login route not found in Routes.js! Cannot redirect.",
        );
        const errorPage = document.createElement("h1");
        errorPage.textContent = "Login Page Not Found";
        // Fallback to 404 if login route itself is misconfigured
        updateHistory(route, addToHistory);
        return handleTransition(errorPage);
      }

      // Important: Push '/account/login' to history, replacing the original protected route if needed.
      // This prevents the back button from returning to the unauthorized page directly.
      // We explicitly set addToHistory to true here for the redirect.
      updateHistory("/account/login", true); // Ensure login page is the one in history

      // Load and transition to the LoginPage component directly
      const loginPageElement = await loadComponent(loginRoute.component);
      return handleTransition(loginPageElement);
    }

    // Case 3: Route matched and user is authenticated (if required), or it's a public route.
    // Proceed with normal navigation and render the matched component.
    updateHistory(route, addToHistory); // Update history only for the route we are successfully navigating to

    const pageElement = await loadComponent(matchedRoute.component);

    // Pass URL parameters to the component if any were extracted by RegExp routes
    if (params.length) {
      pageElement.params = params;
    }

    // Transition to the new page element
    handleTransition(pageElement);
  },
};

/**
 * Finds the matching route and any extracted parameters for the given route path.
 * This function is robust enough to handle trailing slashes by normalizing paths.
 * @param {string} route - The full route path (including query string).
 * @param {Array} routes - The list of route objects to match against.
 * @returns {Object|null} Returns an object with the matched route and params array, or null if no match.
 */
function matchRoute(route, routes) {
  // 1. Strip query parameters
  let routePath = route.includes("?") ? route.split("?")[0] : route;

  // 2. Normalize trailing slashes: Remove trailing slash if it's not the root path "/"
  if (routePath.length > 1 && routePath.endsWith("/")) {
    routePath = routePath.slice(0, -1);
  }
  console.log(`matchRoute: Normalized incoming routePath to "${routePath}"`);

  for (const r of routes) {
    if (typeof r.path === "string") {
      // Normalize the route's defined path as well for comparison
      let definedPath = r.path;
      if (definedPath.length > 1 && definedPath.endsWith("/")) {
        definedPath = definedPath.slice(0, -1);
      }
      console.log(`matchRoute: Comparing with definedPath "${definedPath}"`);

      if (definedPath === routePath) {
        return { route: r, params: [] };
      }
    } else if (r.path instanceof RegExp) {
      // For RegExp, apply normalization to the routePath before testing the regex.
      // The regex itself should ideally be written to account for trailing slashes
      // if they are part of the dynamic path. For now, the overall path normalization is key.
      const match = r.path.exec(routePath);
      if (match) return { route: r, params: match.slice(1) };
    }
  }
  return null;
}

/**
 * Updates the browser history with the given route if required.
 * @param {string} route - The route to push to the browser history.
 * @param {boolean} addToHistory - Whether to add the route to history (pushState) or replace (replaceState).
 */
function updateHistory(route, addToHistory) {
  if (addToHistory) {
    history.pushState(null, "", route);
  } else {
    // Consider replaceState here for initial load or certain transitions
    // For popstate, the URL is already changed, so no need to push/replace
    // For initial load 'go' calls, `pushState` is usually correct.
    // This `else` branch might be unused based on `Router.init` and `popstate`
    history.replaceState(null, "", route);
  }
}

/**
 * Adds click event listeners to navigation links to intercept and handle
 * client-side routing instead of full page reloads.
 * Only targets links with the 'navlink' class.
 * @param {Object} routerInstance - The router instance to invoke navigation (e.g., `Router`).
 */
function interceptLinks(routerInstance) {
  // Ensure 'a.navlink' exists to avoid errors from getElements if none are found.
  try {
    const navLinks = getElements("a.navlink");
    for (const link of navLinks) {
      link.addEventListener("click", (event) => {
        event.preventDefault(); // Prevent default full page reload
        const href = link.getAttribute("href");
        if (href) {
          routerInstance.go(href); // Use the router to navigate
        }
      });
    }
  } catch (error) {
    console.info(
      "No 'a.navlink' elements found to intercept. This might be expected for certain pages.",
      error.message,
    );
    // Do not re-throw, as it's common for some pages to not have navlinks.
  }
}

// function matchRoute(route, routes) {
//   // Strip query parameters for route matching
//   const routePath = route.includes("?") ? route.split("?")[0] : route;

//   for (const r of routes) {
//     if (typeof r.path === "string" && r.path === routePath) {
//       return { route: r, params: [] };
//     } else if (r.path instanceof RegExp) {
//       // Use the routePath (without query params) for RegExp matching
//       const match = r.path.exec(routePath);
//       if (match) return { route: r, params: match.slice(1) };
//     }
//   }
//   return null;
// }

// import { getElements, handleTransition, loadComponent } from "../util.js";
// import { routes } from "./Routes.js";

// export const Router = {
//   init() {
//     window.addEventListener("popstate", () => {
//       // this.go(location.pathname, false);
//       //! For popstate, ensure we pass the full path including search, but don't add to history again
//       this.go(location.pathname + location.search, false);
//       console.log("📍 Router loaded. Current path:", location.pathname);
//     });
//     interceptLinks(this);
//     // Initial load: Go to the current URL, always adding to history as it's the first page.
//     // The history API itself is smart enough not to add duplicate entries for the initial load.
//     this.go(location.pathname + location.search, true);
//     // this.go(location.pathname + location.search);
//   },

//   async go(route, addToHistory = true) {
//     console.log("Navigating to:", route);

//     //! This should ideally happen AFTER a successful match and before rendering
//     // to avoid history issues if a redirect occurs.
//     updateHistory(route, addToHistory);

//     const matched = matchRoute(route, routes);
//     if (!matched) {
//       // If no route matches AND it's not a protected route redirecting away, show 404
//       const notFoundPage = document.createElement("h1");
//       notFoundPage.textContent = "Page not found";
//       console.warn(`Router: No route matched for '${route}'. Displaying 404.`);
//       return handleTransition(notFoundPage);
//     }

//     const { route: matchedRoute, params } = matched;

//     // --- Authentication Check (Crucial for direct access) ---
//     //! Before loading the component,
//     // check authentication status for protected routes
//     // If the matched route requires login AND the user is not currently logged in,
//     // we must redirect to the login page immediately.
//     if (matchedRoute.loggedIn && !window.app.Auth.isLoggedIn()) {
//       // ✨ Check app.Auth.isLoggedIn()
//       console.warn(
//         `Attempted to access protected route '${route}'. Redirecting to login.`,
//       );

//       // Don't add the protected route to history if we're immediately redirecting away from it.
//       // This prevents the user from being able to hit "back" and land on a blank protected page.
//       if (location.pathname !== "/account/login") {
//         // Prevent infinite loop if /account/login itself is considered protected somehow
//         updateHistory("/account/login", true); // Push /account/login to history
//       }
//       // Load and transition to the LoginPage component directly
//       const loginPageElement = await loadComponent(
//         routes.find((r) => r.path === "/account/login").component,
//       );
//       return handleTransition(loginPageElement);

//       // return this.go("/account/login"); // Redirect to login
//     }

//     // --- Normal Route Handling (if authenticated or public route) ---
//     // Only update history if we're actually going to render this matched route.
//     // This avoids pushing a protected route to history right before redirecting away.
//     updateHistory(route, addToHistory);

//     // Load and transition to the component if authenticated or if it's a public route
//     const pageElement = await loadComponent(matchedRoute.component);

//     if (params.length) {
//       pageElement.params = params;
//     }

//     handleTransition(pageElement);
//   },
// };

// /**
//  * Finds the matching route and any extracted parameters for the given route path.
//  * @param {string} route - The full route path (including query string).
//  * @param {Array} routes - The list of route objects to match against.
//  * @returns {Object|null} Returns an object with the matched route and params array, or null if no match.
//  */
// function matchRoute(route, routes) {
//   const routePath = route.includes("?") ? route.split("?")[0] : route;

//   for (const r of routes) {
//     if (typeof r.path === "string" && r.path === routePath) {
//       return { route: r, params: [] };
//     } else if (r.path instanceof RegExp) {
//       const match = r.path.exec(route);
//       if (match) return { route: r, params: match.slice(1) };
//     }
//   }
//   return null;
// }

// /**
//  * Updates the browser history with the given route if required.
//  * @param {string} route - The route to push to the browser history.
//  * @param {boolean} addToHistory - Whether to add the route to history.
//  */
// function updateHistory(route, addToHistory) {
//   if (addToHistory) {
//     history.pushState(null, "", route);
//   }
// }

// /**
//  * Adds click event listeners to navigation links to intercept and handle
//  * client-side routing instead of full page reloads.
//  * @param {Object} routerInstance - The router instance to invoke navigation.
//  */
// function interceptLinks(routerInstance) {
//   for (const link of getElements("a.navlink")) {
//     link.addEventListener("click", (event) => {
//       event.preventDefault();
//       const href = link.getAttribute("href");
//       routerInstance.go(href);
//     });
//   }
// }
// export const Router = {
//   // Initialize the router with popstate listener and link enhancements
//   init() {
//     window.addEventListener("popstate", () => {
//       this.go(location.pathname, false);
//     });

//     // Enhance current links with `click` event listener
//     for (const link of getElements("a.navlink")) {
//       link.addEventListener("click", (event) => {
//         event.preventDefault();
//         const href = link.getAttribute("href");
//         this.go(href);
//       });
//     }

//     // Go to the current path when initializing
//     this.go(location.pathname + location.search);
//   },

//   // Main routing function
//   async go(route, addToHistory = true) {
//     console.log("Navigating to:", route);
//     // if (route === "/account" && !app.Store.loggedIn) {
//     //   this.go("/account/login");
//     //   return;
//     // }

//     // Manage browser history if needed
//     if (addToHistory) {
//       history.pushState(null, "", route);
//     }

//     let pageElement = null;
//     const routePath = route.includes("?") ? route.split("?")[0] : route;
//     let needsLogin = false;

//     // Look through routes to find matching component
//     for (const r of routes) {
//       if (typeof r.path === "string" && r.path === routePath) {
//         pageElement = new r.component();
//         needsLogin = r.loggedIn ?? false; // Use nullish coalescing to set default
//         break;
//       } else if (r.path instanceof RegExp) {
//         const match = r.path.exec(route);
//         if (match) {
//           pageElement = new r.component();
//           pageElement.params = match.slice(1); // Use captured params
//           needsLogin = r.loggedIn ?? false;
//           break;
//         }
//       }
//     }

//     // If login is required but not logged in, redirect to login
//     if (pageElement) {
//       if (needsLogin && app.Store.loggedIn === false) {
//         return this.go("/account/login");
//       }
//     }

//     // Default 404 if no route matches
//     if (!pageElement) {
//       pageElement = document.createElement("h1");
//       pageElement.textContent = "Page not found";
//     }

//     // Insert new page with transition effect
//     const oldPage = getElement("main").firstElementChild;
//     if (oldPage) oldPage.style.viewTransitionName = "old";

//     pageElement.style.viewTransitionName = "new";

//     // Start transition if supported
//     if (!document.startViewTransition) {
//       updatePage(pageElement);
//     } else {
//       document.startViewTransition(() => {
//         updatePage(pageElement);
//       });
//     }
//   },
// };

// export const Router = {
//   init: () => {
//     window.addEventListener("popstate", () => {
//       Router.go(location.pathname, false);
//     });

//     // Enhance current links in the document
//     for (let link of Array.from(getElements("a.navlink"))) {
//       link.addEventListener("click", (event) => {
//         event.preventDefault();
//         const href = link.getAttribute("href");
//         Router.go(href);
//       });
//     }
//     Router.go(location.pathname + location.search);
//   },
//   go: (route, addToHistory = true) => {
//     if (addToHistory) {
//       history.pushState(null, "", route);
//     }

//     let pageElement = null;

//     const routePath = route.includes("?") ? route.split("?")[0] : route;

//     let needsLogin = false;

//     for (const r of routes) {
//       if (typeof r.path === "string" && r.path === routePath) {
//         pageElement = new r.component();
//         needsLogin = r.loggedIn === true;
//         break;
//       } else if (r.path instanceof RegExp) {
//         const match = r.path.exec(route);
//         if (match) {
//           pageElement = new r.component();
//           const params = match.slice(1);
//           pageElement.params = params;
//           needsLogin = r.loggedIn === true;
//           break;
//         }
//       }
//     }

//     if (pageElement) {
//       // If page is from routes
//       if (needsLogin && app.Store.loggedIn == false) {
//         app.Router.go("/account/login");
//         return;
//       }
//     }

//     if (pageElement == null) {
//       pageElement = document.createElement("h1");
//       pageElement.textContent = "Page not found";
//     }

//     // insert new page in UI
//     const oldPage = getElement("main").firstElementChild;
//     if (oldPage) {
//       oldPage.style.viewTransitionName = "old";
//     }
//     pageElement.style.viewTransitionName = "new";

//     if (!document.startViewTransition) {
//       updatePage(pageElement);
//     } else {
//       document.startViewTransition(() => {
//         updatePage(pageElement);
//       });
//     }
//   },
// };

// function redirect(route, addToHistory = true) {
//   if (addToHistory) {
//     history.pushState(null, "", route);
//   }

//   let pageElement = null;

//   const routePath = route.includes("?") ? route.split("?")[0] : route;

//   let needsLogin = false;

//   for (const r of routes) {
//     if (typeof r.path === "string" && r.path === routePath) {
//       pageElement = new r.component();
//       needsLogin = r.loggedIn === true;
//       break;
//     } else if (r.path instanceof RegExp) {
//       const match = r.path.exec(route);
//       if (match) {
//         pageElement = new r.component();
//         const params = match.slice(1);
//         pageElement.params = params;
//         needsLogin = r.loggedIn === true;
//         break;
//       }
//     }
//   }

//   if (pageElement) {
//     // If page is from routes
//     if (needsLogin && app.Store.loggedIn == false) {
//       app.Router.go("/account/login");
//       return;
//     }
//   }

//   if (pageElement == null) {
//     pageElement = document.createElement("h1");
//     pageElement.textContent = "Page not found";
//   }

//   // insert new page in UI
//   const oldPage = getElement("main").firstElementChild;
//   if (oldPage) {
//     oldPage.style.viewTransitionName = "old";
//   }
//   pageElement.style.viewTransitionName = "new";

//   if (!document.startViewTransition) {
//     updatePage(pageElement);
//   } else {
//     document.startViewTransition(() => {
//       updatePage(pageElement);
//     });
//   }
// }
