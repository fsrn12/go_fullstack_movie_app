const JWT_WORKER_URL = "/workers/jwtWorker.js";

class AuthState extends HTMLElement {
  constructor() {
    super();
    this.worker = new Worker(JWT_WORKER_URL, { type: "module" });
    this.jwt = null;

    this.worker.onmessage = (e) => {
      const { type, token } = e.data;
      if (type === "jwt") {
        this.jwt = token;
        window.app.Store.jwt = token;
      } else if (type === "expired") {
        window.app.logout();
      }
    };
  }

  connectedCallback() {
    this.worker.postMessage({ type: "init" });
  }

  static refreshJWTManually() {
    document
      .querySelector("auth-state")
      ?.worker.postMessage({ type: "refresh" });
  }
}

customElements.define("auth-state", AuthState);
