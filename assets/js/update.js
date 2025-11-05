function getButtonElements(target) {
  // If target is a form, find the submit button
  const button = target.tagName === 'FORM' ? target.querySelector('[type="submit"]') : target;
  
  // Find spinner and icon by their classes
  const spinner = button.querySelector('.loading-spinner');
  const icon = button.querySelector('.icon-\\[heroicons--arrow-path\\], .icon-\\[heroicons--arrow-up-tray\\]');
  
  return {
    button,
    spinner,
    icon
  };
}

function setButtonLoadingState(elements, isLoading) {
  elements.button.disabled = isLoading;

  if (isLoading) {
    elements.spinner.classList.remove("hidden");
    elements.icon.classList.add("hidden");
  } else {
    elements.spinner.classList.add("hidden");
    elements.icon.classList.remove("hidden");
  }
}

function createToastElement(isError, message) {
  const toast = document.createElement("div");
  const alertType = isError ? "alert-error" : "alert-info";
  toast.className = `alert ${alertType} rounded-lg`;
  toast.innerHTML = `<span>${message}</span>`;
  return toast;
}

function addToastToContainer(toast) {
  const container = document.getElementById("toast-container");
  container.appendChild(toast);
  setTimeout(() => toast.remove(), 5000);
}

function parseServerResponse(responseText) {
  try {
    return JSON.parse(responseText || "");
  } catch {
    return null;
  }
}

function isErrorResponse(status) {
  return status >= 300;
}

function hasErrorMessage(parsedResponse) {
  return parsedResponse && parsedResponse.message;
}

function startUpdate(event) {
  const elements = getButtonElements(event.target);
  setButtonLoadingState(elements, true);
}

function stopUpdate(event) {
  handleResponse(event);
  const elements = getButtonElements(event.target);
  setButtonLoadingState(elements, false);
}

function showToast(isError, message) {
  const toast = createToastElement(isError, message);
  addToastToContainer(toast);
}

function handleResponse(event) {
  const xhr = event.detail.xhr;
  const parsedResponse = parseServerResponse(xhr.responseText);

  if (isErrorResponse(xhr.status) && hasErrorMessage(parsedResponse)) {
    showToast(true, parsedResponse.message);
  }
}
