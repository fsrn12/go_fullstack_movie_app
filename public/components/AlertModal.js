import { createNode } from "../utils/util.js";

export default class AlertModal extends HTMLElement {
  connectedCallback() {
    this.appendChild(createNode("alert-modal"));
  }
}
customElements.define("alert-modal", AlertModal);
