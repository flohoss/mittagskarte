const { computePosition, offset, flip, shift, autoUpdate } =
  window.FloatingUIDOM;

// Store a reference to the active tooltip and its cleanup function
let activeTooltipId = null;
let activeTooltipCleanup = null;
let hideTimeout;

const links = document.querySelectorAll("[data-lg-blank]");
const displayHelper = document.querySelector("#display-helper");

function updateTargets() {
  const style = window.getComputedStyle(displayHelper);
  const isLg = style.display !== "none";

  links.forEach((link) => {
    if (isLg) {
      link.setAttribute("target", "_blank");
    } else {
      link.removeAttribute("target");
    }
  });
}

updateTargets();
window.addEventListener("resize", updateTargets);

function showTooltip(event, tooltipId) {
  const style = window.getComputedStyle(displayHelper);
  if (style.display === "none") {
    return;
  }

  // Clear any pending hide actions
  clearTimeout(hideTimeout);

  // If a different tooltip is already active, hide it first
  if (activeTooltipId && activeTooltipId !== tooltipId) {
    hideTooltip(null, activeTooltipId);
  }

  const trigger = event.target;
  const tooltip = document.getElementById(tooltipId);

  if (!tooltip) return;

  // Set the currently active tooltip
  activeTooltipId = tooltipId;

  if (!tooltip._originalParent) {
    tooltip._originalParent = tooltip.parentNode;
    tooltip._originalNextSibling = tooltip.nextSibling;
  }

  // Move tooltip to <body> to escape DaisyUI collapse overflow
  document.body.appendChild(tooltip);

  tooltip.classList.remove("hidden");

  // Add event listeners to the tooltip itself
  tooltip.addEventListener("mouseenter", handleTooltipMouseEnter);
  tooltip.addEventListener("mouseleave", handleTooltipMouseLeave);

  // Store the cleanup function for the trigger element
  activeTooltipCleanup = autoUpdate(trigger, tooltip, () => {
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
  // Use a timeout to delay hiding the tooltip
  hideTimeout = setTimeout(() => {
    const tooltip = document.getElementById(tooltipId);

    if (!tooltip) return;

    // Remove event listeners from the tooltip
    tooltip.removeEventListener("mouseenter", handleTooltipMouseEnter);
    tooltip.removeEventListener("mouseleave", handleTooltipMouseLeave);

    // Move tooltip back to original parent and position
    if (tooltip._originalParent) {
      if (tooltip._originalNextSibling) {
        tooltip._originalParent.insertBefore(
          tooltip,
          tooltip._originalNextSibling
        );
      } else {
        tooltip._originalParent.appendChild(tooltip);
      }
    }

    tooltip.classList.add("hidden");

    // Cleanup the floating-ui autoUpdate
    if (activeTooltipCleanup) {
      activeTooltipCleanup();
      activeTooltipCleanup = null;
    }
    activeTooltipId = null;
  }, 100); // 100ms delay
}

function handleTooltipMouseEnter() {
  clearTimeout(hideTimeout);
}

function handleTooltipMouseLeave(event) {
  // Re-hide the tooltip after a delay
  if (activeTooltipId) {
    hideTooltip(event, activeTooltipId);
  }
}
