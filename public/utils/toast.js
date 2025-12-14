// toast.js — or inject this in a script block on app load

/**
 * Creates and shows a toast notification.
 * @param {string} message - The message to display.
 * @param {"success"|"error"|"info"} [type="info"] - Type of toast.
 * @param {number} [duration=3000] - Duration before auto-dismiss (ms).
 */
export function showToast(message, type = "info", duration = 1000) {
  let container = document.querySelector("#toast-container");

  if (!container) {
    container = document.createElement("div");
    container.id = "toast-container";
    document.body.appendChild(container);

    Object.assign(container.style, {
      position: "fixed",
      top: "1rem",
      right: "1rem",
      display: "flex",
      flexDirection: "column",
      gap: "0.5rem",
      zIndex: 9999,
    });
  }

  const toast = document.createElement("div");
  toast.textContent = message;
  toast.className = `toast toast-${type}`;

  Object.assign(toast.style, {
    padding: "0.75rem 1.25rem",
    borderRadius: "6px",
    color: "white",
    fontSize: "0.875rem",
    minWidth: "200px",
    maxWidth: "300px",
    boxShadow: "0 2px 6px rgba(0,0,0,0.15)",
    opacity: "0",
    transform: "translateY(-10px)",
    transition: "opacity 0.3s ease, transform 0.3s ease",
    background:
      type === "success" ? "#4CAF50" : type === "error" ? "#F44336" : "#2196F3",
  });

  container.appendChild(toast);

  // Animate in
  requestAnimationFrame(() => {
    toast.style.opacity = "1";
    toast.style.transform = "translateY(0)";
  });

  // Remove after duration
  setTimeout(() => {
    toast.style.opacity = "0";
    toast.style.transform = "translateY(-10px)";
    toast.addEventListener("transitionend", () => toast.remove(), {
      once: true,
    });
  }, duration);
}
