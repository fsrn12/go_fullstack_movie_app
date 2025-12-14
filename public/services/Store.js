const jwtSymbol = Symbol("jwt"); // Create a unique symbol for the private field

const Store = {
  // Use symbol to store the JWT value privately
  [jwtSymbol]: null,

  get loggedIn() {
    return Boolean(this[jwtSymbol]);
  },

  get jwt() {
    return this[jwtSymbol];
  },

  set jwt(value) {
    if (value === null || typeof value === "string") {
      this[jwtSymbol] = value;
      this._persistJwt(value);
    } else {
      throw new TypeError("JWT must be a string or null");
    }
  },

  _persistJwt(value) {
    if (value === null) {
      localStorage.removeItem("jwt");
    } else {
      localStorage.setItem("jwt", value);
    }
  },

  init() {
    const storedJwt = localStorage.getItem("jwt");
    if (storedJwt) {
      this[jwtSymbol] = storedJwt;
    }
  },

  clear() {
    this[jwtSymbol] = null;
    this._persistJwt(null);
  },
};

// Initialize Store with localStorage data
Store.init();

// Proxy for interacting with the Store
const proxiedStore = new Proxy(Store, {
  set(target, prop, value) {
    if (prop === "jwt") {
      target.jwt = value; // Use the setter of jwt for validation and persisting
    }
    return true;
  },
});

export default proxiedStore;

// const Store = {
//   jwt: null,
//   get loggedIn() {
//     return this.jwt !== null;
//   },
// };

// if (localStorage.getItem("jwt")) {
//   Store.jwt = localStorage.getItem("jwt");
// }

// const proxiedStore = new Proxy(Store, {
//   set: (target, prop, value) => {
//     if (prop == "jwt") {
//       target[prop] = value;
//       if (value == null) {
//         localStorage.removeItem("jwt");
//       } else {
//         localStorage.setItem("jwt", value);
//       }
//     }
//     return true;
//   },
// });

// export default proxiedStore;
