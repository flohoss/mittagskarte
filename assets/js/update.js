function startUpdate(event) {
  const button = event.target;
  const spinner = event.target.firstChild;
  const icon = event.target.lastChild;
  button.disabled = true;
  spinner.classList.remove("hidden");
  icon.classList.add("hidden");
}

function stopUpdate(event) {
  handleResponse(event);
  const button = event.target;
  const spinner = event.target.firstChild;
  const icon = event.target.lastChild;
  button.disabled = false;
  spinner.classList.add("hidden");
  icon.classList.remove("hidden");
}

function showToast(isError, message) {
  const toast = document.createElement("div");
  toast.className =
    (isError ? "alert alert-error" : "alert alert-info") + " rounded-xl";
  toast.innerHTML = `<span>${message}</span>`;

  const container = document.getElementById("toast-container");
  container.appendChild(toast);

  setTimeout(() => toast.remove(), 5000);
}

function handleResponse(event) {
  const xhr = event.detail.xhr;
  const serverMessage = xhr.responseText || "";
  const msg = JSON.parse(serverMessage);

  if (xhr.status >= 300 && msg && msg.message) {
    showToast(true, msg.message);
  }
}
