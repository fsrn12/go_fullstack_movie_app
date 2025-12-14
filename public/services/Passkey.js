const Passkeys = {
  register: async (username) => {
    const token = window.app?.Auth.getJwt();
    try {
      const response = await fetch("/api/passkey/register-begin", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          Authorization: token ? `Bearer $token}` : null,
        },
        body: JSON.stringify({ username: username }),
      });
      // if (!response.data.ok){

      // }
      const data = await response.json();
      app.showError(
        "Failed to get registration options from server.",
        data.err,
      );
      const options = await response.json();

      // A new public-private-key pair is created.
      // triggering browser to display the passkey modal
      const attestationResponse = await SimpleWebAuthnBrowser.startRegistration(
        {
          optionsJSON: options.publicKey,
        },
      );

      // Send attestationResponse back to server for verification and storage.
      const verificationResponse = await fetch(
        "/api/passkey/registration-end",
        {
          method: "POST",
          credentials: "same-origin",
          headers: {
            "Content-Type": "application/json",
            Authorization: token ? `Bearer $token}` : null,
          },
          body: JSON.stringify(attestationResponse),
        },
      );

      const msg = await verificationResponse.json();
      if (verificationResponse.data.ok) {
        app.showError(
          "Your passkey was saved. You can use it next time to login",
        );
      } else {
        app.showError(msg, false);
      }
    } catch (err) {
      app.showError("Error: " + err.message, false);
    }
  },

  // ===================================
  //  🔐 AUTHENTICATE=
  // ====================================
  authenticate: async (email) => {
    try {
      //STEP 1: GET LOGIN OPTIONS WITH CHALLENGE FROM BACKEND
      const response = await fetch("/api/passkey/authentication-begin", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ email }),
      });

      const data = await response.json();
      const { options } = data;

      //STEP 2: TRIGGER BROWSER TO DISPLAY PASSKEY/WEBAUTHN MODAL,
      // NOTE: CHALLENGE IS CONSIDERED SIGNED AFTER THIS STEP
      const assertionResponse = await SimpleWebAuthnBrowser.startAuthentication(
        { optionsJSON: options.publicKey },
      );

      // STEP 3: SEND ASSERTION RESPONSE TO BACKEND FOR VERIFICATION
      const verificationResponse = await fetch(
        "/api/passkey/authentication-end",
        {
          method: "POST",
          credentials: "same-origin",
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify(assertionResponse),
        },
      );

      const data2 = await verificationResponse.json();
      const { serverResponse } = data2;

      if (serverResponse.success) {
        app.Auth.setJwt(serverResponse.jwt);
        app.Router.go("/account/");
      } else {
        app.showError(msg, false);
      }
    } catch (err) {
      console.error(err);
      app.showError(
        "Oops! an unexpected error occurred while authenticating you using a Passkey",
        false,
      );
    }
  },
};
