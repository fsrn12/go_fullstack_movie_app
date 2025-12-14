import { lazy } from "../utils/util.js";

export const routes = [
  {
    // Home page
    name: "HomePage",
    path: "/",
    component: lazy("../components/HomePage.js"),
  },
  {
    // Dynamic movie details page, captures movie ID from URL
    name: "MovieDetailsPage",
    path: /\/movies\/(\d+)/,
    component: lazy("../components/MovieDetailsPage.js"),
  },
  {
    // Movies list page
    name: "MoviesPage",
    path: "/movies",
    component: lazy("../components/MoviesPage.js"),
  },
  {
    // Register page
    name: "RegisterPage",
    path: "/account/register",
    component: lazy("../components/RegisterPage.js"),
  },
  {
    // Login page
    name: "LoginPage",
    path: "/account/login",
    component: lazy("../components/LoginPage.js"),
  },
  {
    // User account page - requires login
    name: "AccountPage",
    path: "/account",
    component: lazy("../components/AccountPage.js"),
    loggedIn: true,
  },
  {
    // User favorites page - requires login
    name: "FavoritesPage",
    path: "/account/favorites",
    component: lazy("../components/FavoritePage.js"),
    loggedIn: true,
  },
  {
    // User watchlist page - requires login
    name: "WatchlistPage",
    path: "/account/watchlist",
    component: lazy("../components/WatchlistPage.js"),
    loggedIn: true,
  },
];
