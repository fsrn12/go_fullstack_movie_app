import API from "../../services/API.js";
import "../components/MovieItem.js";
import { createNode } from "../utils/util.js";

export default class MoviesPage extends HTMLElement {
  async render(query) {
    const urlParams = new URLSearchParams(location.search);
    const order = urlParams.get("order") ?? "";
    const genre = urlParams.get("genre") ?? "";

    const movieList = this.querySelector("ul");
    if (!movieList) return;

    const res = await API.searchMovies(query, order, genre);

    let { movies } = res.data;
    // console.log("[MoviesPage] Movies:", movies);
    if (movies && movies?.length > 0) {
      // movieList.innerHTML = "";
      const items = [];
      for (const movie of movies) {
        const li = document.createElement("li");
        li.dataset.movieId = movie?.id;

        const movieItem = document.createElement("movie-item");
        movieItem.movie = movie;

        li.appendChild(movieItem);
        items.push(li);
      }
      movieList.replaceChildren(...items);
    } else {
      movieList.innerHTML =
        "<h3>There are no movies available with your search</h3>";
    }

    if (order) {
      this.querySelector("#order").value = order;
    }

    if (genre) {
      this.querySelector("#filter").value = genre;
    }
  }

  async loadGenres() {
    // const genres = await API.getGenres();
    const { default: genres } = await import("../genres.json", {
      with: {
        type: "json",
      },
    });

    const select = this.querySelector("select#filter");

    select.innerHTML = `<option value="">Filter by Genre</option>`;

    for (const { id, name } of genres) {
      const option = document.createElement("option");
      option.value = id;
      option.textContent = name;
      select.appendChild(option);
    }
  }

  connectedCallback() {
    this.appendChild(createNode("template-movies"));

    const urlParams = new URLSearchParams(window.location.search);
    const query = urlParams.get("q");

    if (query) {
      this.querySelector("h2").textContent = `'${query}' movies`;
      this.render(query);
    } else {
      app.showError?.("Missing search query.");
    }

    this.loadGenres();
  }
}

customElements.define("movie-page", MoviesPage);
