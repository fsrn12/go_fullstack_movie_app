import { createNode } from "../utils/util.js";

export default class LoginPage extends HTMLElement {
  connectedCallback() {
    this.appendChild(createNode("template-login"));
  }
}

customElements.define("login-page", LoginPage);
