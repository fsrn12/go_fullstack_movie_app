import { createNode } from "../utils/util.js";

export default class CollectionPage extends HTMLElement {
  constructor(endpoint, title, collection) {
    super();
    this.endpoint = endpoint;
    this.title = title;
    this.collection = collection;
  }

  connectedCallback() {
    console.log("COLLECTION PAGE 🔴");

    // Create the template content
    const templateContent = createNode("template-collection");

    // Check if the template exists
    if (templateContent) {
      this.appendChild(templateContent);
      this.render();
    } else {
      console.error("Template 'template-collection' not found.");
    }
  }

  async render() {
    try {
      const response = await this.endpoint();
      const { movies, count } = response.data;
      const movieList = this.querySelector("ul");
      movieList.addEventListener("click", this.#handleClick);

      movieList.innerHTML = "";
      if (movies && count > 0) {
        const items = [];
        for (let movie of movies) {
          const li = document.createElement("li");
          li.dataset.movieId = movie?.id;
          li.dataset.collection = this.collection;
          li.classList.add(`${this.collection}-item`);
          // Create the remove button
          const removeButton = document.createElement("button");

          const rippleDiv = document.createElement("div");
          rippleDiv.classList.add("btn-ripple");

          removeButton.className = "btn-remove action-btn primary btn-primary";
          removeButton.textContent = "Remove";
          removeButton.appendChild(rippleDiv);

          const movieItem = document.createElement("movie-item");
          movieItem.movie = movie;
          // li.appendChild(new MovieItemComponent(movie));
          // movieList.appendChild(li);

          li.appendChild(movieItem);
          li.appendChild(removeButton);
          items.push(li);
        }
        movieList.replaceChildren(...items);
      } else {
        movieList.innerHTML = "<h3>No movies to display</h3>";
      }
    } catch (err) {
      console.error("Error fetching collection data:", err);
      const movieList = this.querySelector("ul");
      if (movieList) {
        movieList.innerHTML = `<h3>Failed to load movies.</h3>`;
      }
    }
  }

  async #handleClick(event) {
    const removeBtn = event.target.closest(".btn-remove");

    if (removeBtn) {
      const listItem = removeBtn.closest("li");

      const movieId = parseInt(listItem.dataset.movieId, 10);
      const collection = listItem.dataset.collection;

      if (!movieId || !collection) {
        console.error("Missing movie ID or collection data.");
        return;
      }

      await window.app.removeFromCollection(movieId, collection);
    }
  }
}

// customElements.define("collection-page", CollectionPage);
