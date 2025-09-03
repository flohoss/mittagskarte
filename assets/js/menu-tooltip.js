const { computePosition, offset, flip, shift, autoUpdate } =
  window.FloatingUIDOM;

function showTooltip(event, tooltipId) {
  const trigger = event.target;
  const tooltip = document.getElementById(tooltipId);

  if (!tooltip) return;

  // Move tooltip to <body> to escape DaisyUI collapse overflow
  document.body.appendChild(tooltip);

  tooltip.classList.remove("hidden");

  trigger._tooltipCleanup = autoUpdate(trigger, tooltip, () => {
    computePosition(trigger, tooltip, {
      placement: "right",
      middleware: [offset(8), flip(), shift({ padding: 8 })],
    }).then(({ x, y }) => {
      Object.assign(tooltip.style, {
        position: "absolute",
        left: `${x}px`,
        top: `${y}px`,
      });
    });
  });
}

function hideTooltip(event, tooltipId) {
  const trigger = event.target;
  const tooltip = document.getElementById(tooltipId);

  if (!tooltip) return;

  tooltip.classList.add("hidden");

  if (trigger._tooltipCleanup) {
    trigger._tooltipCleanup();
    trigger._tooltipCleanup = null;
  }
}
