import { createNode } from "../utils/util.js";

export default class RegisterPage extends HTMLElement {
  connectedCallback() {
    this.appendChild(createNode("template-register"));
  }
}

customElements.define("register-page", RegisterPage);
