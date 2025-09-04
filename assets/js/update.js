function startUpdate(event) {
  const button = event.target;
  const spinner = event.target.firstChild;
  const icon = event.target.lastChild;
  button.disabled = true;
  spinner.classList.remove("hidden");
  icon.classList.add("hidden");
}

function stopUpdate(event) {
  const button = event.target;
  const spinner = event.target.firstChild;
  const icon = event.target.lastChild;
  button.disabled = false;
  spinner.classList.add("hidden");
  icon.classList.remove("hidden");
}
