export default class GenreList extends HTMLElement {
  constructor() {
    super();
    this.attachShadow({ mode: "open" });
    this.shadowRoot.innerHTML = `
      <style>
        ul {
          display: flex;
          flex-wrap: wrap;
          gap: 0.75rem;
        }
        li {
          background: oklch(100% 3.5594404092893915e-8 106.37411396324427 / 10%);
          padding: 0.4rem 1rem;
          border-radius: 20px;
          font-size: 0.9rem;
        }
      </style>
      <ul></ul>
    `;
  }

  set genres(genres) {
    const ul = this.shadowRoot.querySelector("ul");
    ul.innerHTML = "";
    genres.forEach(({ id, name }) => {
      const li = document.createElement("li");
      li.textContent = name ?? "unknown";
      li.dataset.genreId = id ?? "unknown";
      ul.appendChild(li);
    });
  }
}

customElements.define("genre-list", GenreList);
