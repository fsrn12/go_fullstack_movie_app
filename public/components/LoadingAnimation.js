class LoadingAnimation extends HTMLElement {
  static get observedAttributes() {
    return ["bars", "width", "height", "color", "gap"];
  }

  constructor() {
    super();
    this.attachShadow({ mode: "open" });
    this._bars = 5;
    this._width = 20; // px
    this._height = 6; // px
    this._color = "#4a90e2";
    this._gap = 8; // px
  }

  connectedCallback() {
    this._render();
  }

  attributeChangedCallback(name, oldVal, newVal) {
    if (oldVal === newVal) return;

    switch (name) {
      case "bars":
        this._bars = parseInt(newVal) || 5;
        break;
      case "width":
        this._width = parseInt(newVal) || 20;
        break;
      case "height":
        this._height = parseInt(newVal) || 6;
        break;
      case "color":
        this._color = newVal || "#4a90e2";
        break;
      case "gap":
        this._gap = parseInt(newVal) || 8;
        break;
    }
    this._render();
  }

  _render() {
    const style = `
      <style>
        :host {
          display: inline-block;
          aria-role: status;
          aria-label: "Loading";
        }
        .wrapper {
          display: flex;
          align-items: center;
          gap: ${this._gap}px;
          height: ${this._height}px;

        }
        .bar {
          width: ${this._width}px;
          height: ${this._height}px;
          background-color: ${this._color};
               background: linear-gradient(
      90deg,
      lch(35% 5 10) 0%,
      lch(60% 10 15) 50%,
      lch(35% 5 10) 100%
    )
          border-radius: 3px;
          animation: wave 1.2s infinite ease-in-out;
        }
        .bar:nth-child(1) { animation-delay: -1.1s; }
        .bar:nth-child(2) { animation-delay: -1.0s; }
        .bar:nth-child(3) { animation-delay: -0.9s; }
        .bar:nth-child(4) { animation-delay: -0.8s; }
        .bar:nth-child(5) { animation-delay: -0.7s; }

        @keyframes wave {
          0%, 40%, 100% {
            transform: scaleY(0.4);
            opacity: 0.6;
          }
          20% {
            transform: scaleY(1);
            opacity: 1;
          }
        }
      </style>
    `;

    // Create bars based on this._bars
    let barsHTML = "";
    for (let i = 0; i < this._bars; i++) {
      barsHTML += `<div class="bar" style="animation-delay: ${
        -1.1 + i * 0.1
      }s"></div>`;
    }

    this.shadowRoot.innerHTML = `
      ${style}
      <div class="wrapper" role="status" aria-label="Loading">
        ${barsHTML}
      </div>
    `;
  }
}

customElements.define("loading-animation", LoadingAnimation);
