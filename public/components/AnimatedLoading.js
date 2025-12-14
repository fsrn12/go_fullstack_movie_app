export default class AnimatedLoading extends HTMLElement {
  constructor() {
    super();
  }

  connectedCallback() {
    this.innerHTML = "";
    const qty = Number(this.dataset.elements ?? 1);
    const width = this.dataset.width ?? "6.25rem"; //100px
    const height = this.dataset.height ?? "0.625rem"; //10px

    this.setAttribute("role", "status");
    this.setAttribute("aria-label", "Loading...");

    for (let i = 0; i < qty; i++) {
      const wrapper = document.createElement("div");
      wrapper.classList.add("loading-wave");
      Object.assign(wrapper.style, {
        width,
        height,
        display: "inline-block",
        margin: "0.625rem",
      });
      this.append(wrapper);
    }

    // const elements = Array.from({ length: qty }, () => {
    //   const wrapper = document.createElement("div");
    //   wrapper.classList.add("loading-wave");
    //   Object.assign(wrapper.style, {
    //     width,
    //     height,
    //     display: "inline-block",
    //     margin: "0.625rem",
    //   });
    //   console.log(wrapper);
    //   return wrapper;
    // });
    // this.replaceChildren(...elements);
  }
}
customElements.define("animated-loading", AnimatedLoading);
