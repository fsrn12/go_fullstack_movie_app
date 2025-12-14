export const validateName = (name) => {
  const trimmed = name.trim();
  if (!/^[A-Za-z\s'-]{2,}$/.test(trimmed)) {
    return "Please enter a valid full name";
  }
  return null;
};

export const validateEmail = (email) => {
  const trimmed = email.trim().toLowerCase();
  const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
  if (!emailRegex.test(trimmed)) {
    return "Please enter a valid email address";
  }
  return null;
};

export const validatePassword = (password) => {
  if (!/^(?=.*[A-Za-z])(?=.*\d)[A-Za-z\d@$!%*?&]{8,}$/.test(password)) {
    return "Password must be at least 8 characters and include both letters and numbers";
  }
  return null;
};

export const validatePasswordConfirmation = (password, confirmation) => {
  console.table(password, confirmation);
  if (password !== confirmation) {
    return "Passwords do not match";
  }
  return null;
};

export const validatePasswordForLogin = (password) => {
  if (!password || password.length < 1) {
    return "Please enter your password";
  }
  return null;
};
