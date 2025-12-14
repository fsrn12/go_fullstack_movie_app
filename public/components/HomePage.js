import API from "../../services/API.js";
import "../components/MovieItem.js";
import { createNode } from "../utils/util.js";
export default class HomePage extends HTMLElement {
  async render() {
    try {
      const topMovies = await API.getTopMovies();
      this.#renderMovies(
        topMovies.data.movies,
        this.querySelector("#top-10 ul"),
      );
      // console.log("[home] Top10:", topMovies.movies);

      const randomMovies = await API.getRandomMovies();
      this.#renderMovies(
        randomMovies.data.movies,
        this.querySelector("#random ul"),
      );
      // console.log("[home] Random:", randomMovies.movies);
    } catch (err) {
      console.error("Failed to load movies:", err);
      this.innerHTML = `<p class="error">Sorry, we couldn't load movies right now. Please try again later.</p>`;
    }
  }
  /**
   * Renders a list of movies into the specified <ul>
   * @param {Array} movies - List of movie objects
   * @param {HTMLUListElement} ul - Target list element
   */
  #renderMovies(movies = [], ul) {
    if (!ul) return;

    const items = [];

    for (const movie of movies) {
      const li = document.createElement("li");
      li.dataset.movieId = movie?.id;

      const movieItem = document.createElement("movie-item");
      movieItem.movie = movie;

      li.appendChild(movieItem);
      items.push(li);
    }

    ul.replaceChildren(...items);
  }

  connectedCallback() {
    this.appendChild(createNode("template-home"));
    // console.log("[home] connectedCallback");
    this.render();
  }
}

customElements.define("home-page", HomePage);
