export default class AnimatedLoading extends HTMLElement {
  #elements = 1;
  #type = "skeleton";
  #animationId = null;

  static get observedAttributes() {
    return ["elements", "type"];
  }

  connectedCallback() {
    this.#elements = parseInt(this.getAttribute("elements") || "1");
    this.#type = this.getAttribute("type") || "skeleton";

    this.#render();
    this.#startAnimation();
  }

  disconnectedCallback() {
    if (this.#animationId) {
      cancelAnimationFrame(this.#animationId);
    }
  }

  attributeChangedCallback(name, oldValue, newValue) {
    if (oldValue === newValue) return;

    if (name === "elements") {
      this.#elements = parseInt(newValue || "1");
    } else if (name === "type") {
      this.#type = newValue || "skeleton";
    }

    this.#render();
  }

  #render() {
    switch (this.#type) {
      case "dots":
        this.innerHTML = this.#renderDots();
        break;
      case "bars":
        this.innerHTML = this.#renderBars();
        break;
      case "pulse":
        this.innerHTML = this.#renderPulse();
        break;
      case "wave":
        this.innerHTML = this.#renderWave();
        break;
      default:
        this.innerHTML = this.#renderSkeleton();
    }
  }

  #renderSkeleton() {
    const skeletons = Array.from({ length: this.#elements }, (_, i) => {
      const height = 20 + i * 5;
      const width = Math.max(60, 100 - i * 15);
      const delay = i * 0.2;

      return `
                <div style="
                    height: ${height}px;
                    width: ${width}%;
                    background: linear-gradient(90deg,
                        var(--surface-1, #1a1a1a) 25%,
                        var(--surface-2, #2a2a2a) 50%,
                        var(--surface-1, #1a1a1a) 75%
                    );
                    background-size: 200% 100%;
                    animation: shimmer 1.5s ease-in-out infinite;
                    animation-delay: ${delay}s;
                    border-radius: 8px;
                    margin-bottom: 12px;
                " class="skeleton-item"></div>
            `;
    }).join("");

    return `
            <div class="skeleton-container">
                ${skeletons}
                <style>
                    @keyframes shimmer {
                        0% { background-position: -200% 0; }
                        100% { background-position: 200% 0; }
                    }
                </style>
            </div>
        `;
  }

  #renderDots() {
    const dots = Array.from(
      { length: this.#elements },
      (_, i) => `
            <div style="
                width: 12px;
                height: 12px;
                border-radius: 50%;
                background: linear-gradient(135deg, var(--primary, #ff6b35), var(--secondary, #4ecdc4));
                margin: 0 6px;
                animation: bounce 1.4s infinite ease-in-out;
                animation-delay: ${i * 0.16}s;
            "></div>
        `,
    ).join("");

    return `
            <div style="
                display: flex;
                justify-content: center;
                align-items: center;
                padding: 20px;
            ">
                ${dots}
                <style>
                    @keyframes bounce {
                        0%, 80%, 100% {
                            transform: scale(0.8);
                            opacity: 0.5;
                        }
                        40% {
                            transform: scale(1.2);
                            opacity: 1;
                        }
                    }
                </style>
            </div>
        `;
  }

  #renderBars() {
    const bars = Array.from(
      { length: this.#elements },
      (_, i) => `
            <div style="
                width: 4px;
                height: 40px;
                background: linear-gradient(135deg, var(--primary, #ff6b35), var(--secondary, #4ecdc4));
                margin: 0 3px;
                border-radius: 2px;
                animation: stretch 1.2s infinite ease-in-out;
                animation-delay: ${i * 0.1}s;
                transform-origin: bottom;
            "></div>
        `,
    ).join("");

    return `
            <div style="
                display: flex;
                justify-content: center;
                align-items: end;
                padding: 20px;
                height: 80px;
            ">
                ${bars}
                <style>
                    @keyframes stretch {
                        0%, 40%, 100% {
                            transform: scaleY(0.4);
                            opacity: 0.7;
                        }
                        20% {
                            transform: scaleY(1);
                            opacity: 1;
                        }
                    }
                </style>
            </div>
        `;
  }

  #renderPulse() {
    return `
            <div style="
                width: 60px;
                height: 60px;
                border-radius: 50%;
                background: linear-gradient(135deg, var(--primary, #ff6b35), var(--secondary, #4ecdc4));
                margin: 20px auto;
                animation: pulse 2s infinite;
            "></div>
            <style>
                @keyframes pulse {
                    0% {
                        transform: scale(1);
                        opacity: 1;
                    }
                    50% {
                        transform: scale(1.2);
                        opacity: 0.7;
                    }
                    100% {
                        transform: scale(1);
                        opacity: 1;
                    }
                }
            </style>
        `;
  }

  #renderWave() {
    const waves = Array.from(
      { length: this.#elements },
      (_, i) => `
            <div style="
                position: absolute;
                width: ${40 + i * 20}px;
                height: ${40 + i * 20}px;
                border: 2px solid var(--primary, #ff6b35);
                border-radius: 50%;
                animation: wave 2s infinite;
                animation-delay: ${i * 0.3}s;
                opacity: ${1 - i * 0.2};
            "></div>
        `,
    ).join("");

    return `
            <div style="
                position: relative;
                width: 100px;
                height: 100px;
                margin: 20px auto;
                display: flex;
                align-items: center;
                justify-content: center;
            ">
                ${waves}
                <style>
                    @keyframes wave {
                        0% {
                            transform: scale(0);
                            opacity: 1;
                        }
                        100% {
                            transform: scale(1);
                            opacity: 0;
                        }
                    }
                </style>
            </div>
        `;
  }

  #startAnimation() {
    // Performance monitoring for animations
    let lastTime = 0;
    console.log(this.#type);
    const animate = (currentTime) => {
      const deltaTime = currentTime - lastTime;

      // Only update if enough time has passed (60fps throttling)
      if (deltaTime >= 16.67) {
        // Custom animation logic can go here
        lastTime = currentTime;
      }

      this.#animationId = requestAnimationFrame(animate);
    };

    this.#animationId = requestAnimationFrame(animate);
  }

  // Public API
  setType(type) {
    this.setAttribute("type", type);
  }

  setElements(count) {
    this.setAttribute("elements", count.toString());
  }

  stop() {
    if (this.#animationId) {
      cancelAnimationFrame(this.#animationId);
      this.#animationId = null;
    }
  }

  start() {
    if (!this.#animationId) {
      this.#startAnimation();
    }
  }
}

// Auto-register the component when imported
// if (!customElements.get("animated-loading")) {
//   customElements.define("animated-loading", AnimatedLoading);
// }
