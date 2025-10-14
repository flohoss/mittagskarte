const { computePosition, offset, flip, shift, autoUpdate } =
  window.FloatingUIDOM;

let activeTooltipId = null;
let hideTimeout;

const links = document.querySelectorAll("[data-lg-blank]");
const displayHelper = document.querySelector("#display-helper");

// Update links target based on #display-helper visibility
function updateTargets() {
  const isLg = window.getComputedStyle(displayHelper).display !== "none";
  links.forEach((link) => {
    if (isLg) link.setAttribute("target", "_blank");
    else link.removeAttribute("target");
  });
}

updateTargets();
window.addEventListener("resize", updateTargets);

// Show tooltip
function showTooltip(event, tooltipId) {
  const tooltip = document.getElementById(tooltipId);
  if (!tooltip || window.getComputedStyle(displayHelper).display === "none")
    return;

  clearTimeout(hideTimeout);

  // Hide previous tooltip immediately if switching
  if (activeTooltipId && activeTooltipId !== tooltipId) {
    hideTooltip(activeTooltipId, true);
  }

  const trigger = event.target;
  activeTooltipId = tooltipId;

  if (!tooltip._originalParent) {
    tooltip._originalParent = tooltip.parentNode;
    tooltip._originalNextSibling = tooltip.nextSibling;
  }

  document.body.appendChild(tooltip);
  tooltip.classList.remove("hidden");

  tooltip.addEventListener("mouseenter", handleTooltipMouseEnter);
  tooltip.addEventListener("mouseleave", handleTooltipMouseLeave);

  // Cleanup previous tooltip autoUpdate if exists
  if (tooltip._cleanup) {
    try {
      tooltip._cleanup();
    } catch (e) {}
  }

  // Store autoUpdate cleanup on tooltip itself
  tooltip._cleanup = autoUpdate(trigger, tooltip, () => {
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

// Hide tooltip
function hideTooltip(tooltipId) {
  const doHide = () => {
    const tooltip = document.getElementById(tooltipId);
    if (!tooltip) return;

    tooltip.removeEventListener("mouseenter", handleTooltipMouseEnter);
    tooltip.removeEventListener("mouseleave", handleTooltipMouseLeave);

    if (tooltip._originalParent) {
      if (tooltip._originalNextSibling)
        tooltip._originalParent.insertBefore(
          tooltip,
          tooltip._originalNextSibling
        );
      else tooltip._originalParent.appendChild(tooltip);
    }

    tooltip.classList.add("hidden");

    if (tooltip._cleanup) {
      try {
        tooltip._cleanup();
      } catch (e) {}
      tooltip._cleanup = null;
    }

    if (activeTooltipId === tooltipId) activeTooltipId = null;
  };

  clearTimeout(hideTimeout);
  hideTimeout = setTimeout(doHide, 150);
}

function handleTooltipMouseEnter() {
  clearTimeout(hideTimeout);
}

function handleTooltipMouseLeave(event) {
  if (activeTooltipId) hideTooltip(activeTooltipId);
}
