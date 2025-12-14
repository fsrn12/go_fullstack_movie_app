export default class MovieItemComponent extends HTMLElement {
  #movie = null;

  set movie(data) {
    this.#movie = data;

    if (this.isConnected) {
      this.render();
    }
  }

  get movie() {
    return this.#movie;
  }

  connectedCallback() {
    if (this.#movie) {
      this.render();
    }
  }

  render() {
    if (!this.#movie) {
      return;
    }

    const { title, poster_url, release_year, id: movieId } = this.#movie;
    const url = `/movies/${movieId}`;

    this.innerHTML = `
      <a href="${url}" class="navlink"
      aria-"label=View details for ${title ?? "Untitled"}"
      onclick="event.preventDefault();app.Router.go(${url})"
      >
        <article>
          <img src="${poster_url ?? "/images/default_poster.jpg"}" alt="${
      title ?? "Untitled"
    } Poster" loading="lazy">
          <p>${title ?? "Untitled"} (${release_year ?? "N/A"})</p>
        </article>
      </a>
    `;

    // const link = this.querySelector("a");
    // link.setAttribute("aria-label", `View details for ${title ?? "Untitled"}`);
    // link.addEventListener("click", (e) => {
    //   e.preventDefault();
    //   app.Router.go(`/movies/${movieId}`);
    // });
  }
}

customElements.define("movie-item", MovieItemComponent);
