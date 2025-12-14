import { createNode, getElement } from "../utils/util.js";

/**
 * Custom Web Component representing the Account Page.
 * Manages user profile display and profile picture upload functionality.
 */
export default class AccountPage extends HTMLElement {
  #currentUser = null;
  #selectedFile = null;
  #originalPictureSrc = null;
  #elements = {};
  #abortController = new AbortController();

  /**
   * Set the current user data.
   * @param {Object} data - User data object.
   */
  set currentUser(data) {
    this.#currentUser = data;
  }

  /**
   * Get the current user data.
   * @returns {Object|null} Current user data or null if not set.
   */
  get currentUser() {
    return this.#currentUser;
  }

  /**
   * Lifecycle callback invoked when the element is added to the DOM.
   * Initializes the component UI and logic.
   */
  connectedCallback() {
    this.appendChild(createNode("template-account"));

    this.refreshPage = this.refreshPage.bind(this);
    window.app.subscribe("collection-updated", this.refreshPage);
    this.#render();
  }

  /**
   * Lifecycle callback invoked when the element is removed from the DOM.
   * Cleans up event listeners to prevent memory leaks.
   */
  disconnectedCallback() {
    window.app.unsubscribe("collection-updated", this.refreshPage);
    this.#abortController.abort();
  }

  /**
   * Private method to perform initial rendering, caching, event binding,
   * welcome message update, and user profile fetch.
   * @private
   */
  async #render() {
    this.#cacheElements();
    this.#bindEvents();

    this.#updateWelcomeMessage();
    await this.#fetchUserProfile();
  }

  /**
   * Cache and store references to important DOM elements for later use.
   * Also sets the original profile picture URL as a fallback.
   * @private
   */
  #cacheElements() {
    /**
     * Shortcut function to get element within this component.
     * @param {string} selector - CSS selector to query.
     * @returns {HTMLElement|null}
     */
    const $ = (selector) => getElement(selector, this);

    this.#elements = {
      welcomeHeading: $("#welcome-message"),
      profileDisplay: $("#profile-picture-display"),
      profileUpload: $("#profile-picture-upload"),
      uploadBtn: $("#upload-button"),
      saveBtn: $("#save-picture-button"),
      cancelBtn: $("#cancel-picture-button"),
    };

    this.#originalPictureSrc =
      this.#elements.profileDisplay?.src ?? "/images/generic_actor.jpg";
  }

  /**
   * Bind all relevant event listeners for user interactions.
   * Uses AbortController for cleanup on disconnect.
   * @private
   */
  #bindEvents() {
    const { profileUpload, uploadBtn, saveBtn, cancelBtn } = this.#elements;

    uploadBtn?.addEventListener("click", () => profileUpload?.click(), {
      signal: this.#abortController.signal,
    });

    profileUpload?.addEventListener("change", this.#onFileSelected.bind(this), {
      signal: this.#abortController.signal,
    });

    saveBtn?.addEventListener("click", this.#onSavePicture.bind(this), {
      signal: this.#abortController.signal,
    });

    cancelBtn?.addEventListener("click", this.#onCancelUpload.bind(this), {
      signal: this.#abortController.signal,
    });
  }

  /**
   * Update the welcome message based on the current authenticated user.
   * Displays first name if available, or email, or a fallback.
   * Uses try/catch to prevent failure if Auth API is unavailable.
   * @private
   */
  #updateWelcomeMessage() {
    try {
      const user = window.app.Auth.getUser();
      const name = user?.name ?? user?.email ?? "user";
      const firstName =
        typeof name === "string" && name.includes(" ")
          ? name.split(" ")[0]
          : name;

      this.#elements.welcomeHeading.textContent = `Welcome, ${firstName}!`;
    } catch (err) {
      console.error("Error updating welcome message:", err);
    }
  }

  /**
   * Fetch fresh user profile data asynchronously from the backend API.
   * Updates the internal user state and profile picture on success.
   * Logs error on failure without blocking UI.
   * @returns {Promise<void>}
   * @private
   */
  async #fetchUserProfile() {
    try {
      const { data } = await window.app.API.fetchUserProfile();
      if (data?.success) {
        this.#currentUser = data.User;
        this.#elements.profileDisplay.src =
          data.user?.profilePictureUrl ?? this.#originalPictureSrc;
      }
    } catch (err) {
      console.error("Failed to retrieve user profile", err);
    }
  }

  /**
   * Handle file selection event when user chooses a new profile picture.
   * Reads the file as a Data URL to show preview.
   * Updates UI to show Save and Cancel buttons.
   * @param {Event} event - Change event from file input.
   * @private
   */
  #onFileSelected(event) {
    const file = event.target.files?.[0];
    if (!file) return;

    this.#selectedFile = file;
    const reader = new FileReader();

    reader.onload = (e) => {
      this.#elements.profileDisplay.src = e.target.result;
      this.#toggleUploadButtons(true);
    };

    reader.readAsDataURL(file);
  }

  /**
   * Handle saving the selected profile picture by uploading to the backend.
   * Updates global user data on success.
   * Provides user feedback for success or failure.
   * Reverts preview on failure or error.
   * @returns {Promise<void>}
   * @private
   */
  async #onSavePicture() {
    if (!this.#selectedFile) return;

    window.app.showLoading("Uploading profile picture...");

    try {
      const { data } = await window.app.API.uploadProfilePicture(
        this.#selectedFile,
      );

      if (data?.success && data.profilePictureUrl) {
        const newUrl = data.profilePictureUrl;

        window.app.showToast(
          "Profile picture uploaded successfully!",
          "success",
        );
        window.app.Auth.setUser({ profilePictureUrl: newUrl });

        const user = window.app.Auth.getUser();
        if (user) user.profilePictureUrl = newUrl;

        this.#elements.profileDisplay.src = newUrl;
        this.#originalPictureSrc = newUrl;
        this.#selectedFile = null;
      } else {
        this.#handleUploadFailure(data?.message);
      }
    } catch (error) {
      console.error("Error uploading profile picture:", error);
      this.#handleUploadFailure();
    } finally {
      this.#toggleUploadButtons(false);
      window.app.hideLoading();
    }
  }

  /**
   * Handle cancelling the profile picture upload process.
   * Reverts preview to original image.
   * Resets file selection and toggles UI buttons.
   * @private
   */
  #onCancelUpload() {
    this.#elements.profileDisplay.src = this.#originalPictureSrc;
    this.#selectedFile = null;
    this.#toggleUploadButtons(false);
  }

  /**
   * Handles upload failure by showing error message and reverting preview image.
   * @param {string} [message="Failed to upload profile picture."] - Optional error message to display.
   * @private
   */
  #handleUploadFailure(message = "Failed to upload profile picture.") {
    window.app.showError(message);
    this.#elements.profileDisplay.src = this.#originalPictureSrc;
  }

  /**
   * Toggle visibility of upload, save, and cancel buttons based on editing state.
   * @param {boolean} isEditing - If true, show save/cancel and hide upload button.
   * @private
   */
  #toggleUploadButtons(isEditing) {
    this.#elements.uploadBtn.style.display = isEditing
      ? "none"
      : "inline-block";
    this.#elements.saveBtn.style.display = isEditing ? "inline-block" : "none";
    this.#elements.cancelBtn.style.display = isEditing
      ? "inline-block"
      : "none";
  }

  async refreshPage() {
    console.log("Collection Updated, refreshing page...");
    await this.#render();
  }
}

customElements.define("account-page", AccountPage);
