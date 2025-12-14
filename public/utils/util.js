export const getElement = function (el, parent = document) {
  const element = parent.querySelector(el);
  if (!element) {
    console.warn("No element found for selector: " + el);
    return null;
  }
  return element;
};

export const getElements = function (el, parent = document) {
  const elements = parent.querySelectorAll(el);
  if (elements.length === 0) {
    console.warn("No elements found for selector: " + el);
    return [];
  }
  return elements;
};

export const createNode = function (id) {
  const template = document.getElementById(id);
  if (!template) {
    console.warn(`Template with id '${id}' not found.`);
    return null;
  }
  return template.content.cloneNode(true);
};

export const lazy = (path) => async () => (await import(path)).default;

/**
 * Loads a component, supporting both lazy-loaded (async) and static imports,
 * then creates a new instance of the component.
 * @param {Function|Object} component - The component constructor or async loader function.
 * @returns {Promise<Object>} An instance of the component.
 */
export async function loadComponent(component) {
  // If component is a function, call it (lazy loaded), else use it directly (static import)
  const Comp = typeof component === "function" ? await component() : component;
  return new Comp();
}

export const updatePage = function (parent) {
  const mainEl = getElement("main");
  mainEl.innerHTML = "";
  mainEl.appendChild(parent);
};

/**
 * Handles page transition animations when switching pages.
 * Uses View Transitions API if available, otherwise falls back to direct update.
 * @param {HTMLElement} newPage - The new page element to display.
 */
export function handleTransition(newPage) {
  const mainEl = document.querySelector("main");
  if (!mainEl) {
    console.error("Main content element not found.");
    return;
  }

  const oldPage = mainEl.firstElementChild;

  // Check if the browser supports the View Transitions API
  if (!document.startViewTransition) {
    // Fallback for browsers that don't support View Transitions
    if (oldPage) {
      mainEl.removeChild(oldPage);
    }
    mainEl.appendChild(newPage);
    return;
  }

  // Use View Transitions API if available
  document.startViewTransition(() => {
    if (oldPage) {
      oldPage.style.viewTransitionName = "old"; // Apply a name for the transition
      oldPage.remove(); // Safely remove the old page from the DOM
    }
    mainEl.appendChild(newPage); // Append the new page
    newPage.style.viewTransitionName = "new"; // Apply a name for the new page
  });
}

// export function handleTransition(newPage) {
//   const oldPage = getElement("main").firstElementChild;
//   if (oldPage) oldPage.style.viewTransitionName = "old";

//   newPage.style.viewTransitionName = "new";

//   if (!document.startViewTransition) {
//     updatePage(newPage);
//   } else {
//     document.startViewTransition(() => {
//       updatePage(newPage);
//     });
//   }
// }

// export interface IUser {
//   UserID: string;
//   Name: string;
//   Email: string;
// }
