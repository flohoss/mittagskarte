// countdown.js

function updateCountdown() {
  const now = new Date();
  const lunchTime = new Date();

  // today's lunch at 12:30
  lunchTime.setHours(12, 30, 0, 0);

  let diff = lunchTime - now;

  // if it's past lunch → freeze at 00:00:00 until midnight
  if (now >= lunchTime) {
    diff = 0;
  }

  // calculate hours, minutes, seconds
  const hours = Math.floor(diff / (1000 * 60 * 60));
  const minutes = Math.floor((diff % (1000 * 60 * 60)) / (1000 * 60));
  const seconds = Math.floor((diff % (1000 * 60)) / 1000);

  const values = [hours, minutes, seconds];
  const spans = document.querySelectorAll(".countdown span");

  // update spans
  spans.forEach((span, i) => {
    const val = String(values[i]).padStart(2, "0");
    span.textContent = val;
    span.setAttribute("aria-label", values[i]);
    span.style.setProperty("--value", values[i]);
  });

  // color rules
  const countdownContainer = document.querySelector(".flex.gap-5");
  if (diff === 0) {
    countdownContainer.style.color = "oklch(70.4% 0.191 22.216)"; // Lunch time
  } else if (diff <= 1000 * 60 * 60) {
    countdownContainer.style.color = "oklch(75% 0.183 55.934)"; // < 1h left
  } else {
    countdownContainer.style.color = ""; // default
  }
}

// run immediately, then every second
updateCountdown();
setInterval(updateCountdown, 1000);
