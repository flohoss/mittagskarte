const { computePosition, offset, flip, shift, autoUpdate } =
  window.FloatingUIDOM;

let activeTooltipId = null;
let hideTimeout = null;
let activeTooltip = null;

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

// Helper function to immediately hide any active tooltip
function hideActiveTooltipImmediate() {
  if (activeTooltip && activeTooltipId) {
    // Clear any pending hide timeout
    if (hideTimeout) {
      clearTimeout(hideTimeout);
      hideTimeout = null;
    }
    
    // Remove event listeners
    activeTooltip.removeEventListener("mouseenter", handleTooltipMouseEnter);
    activeTooltip.removeEventListener("mouseleave", handleTooltipMouseLeave);

    // Restore to original position
    if (activeTooltip._originalParent) {
      if (activeTooltip._originalNextSibling)
        activeTooltip._originalParent.insertBefore(
          activeTooltip,
          activeTooltip._originalNextSibling
        );
      else activeTooltip._originalParent.appendChild(activeTooltip);
    }

    // Hide tooltip
    activeTooltip.classList.add("hidden");

    // Cleanup positioning
    if (activeTooltip._cleanup) {
      try {
        activeTooltip._cleanup();
      } catch (e) {}
      activeTooltip._cleanup = null;
    }

    // Reset state
    activeTooltipId = null;
    activeTooltip = null;
  }
}

// Show tooltip
function showTooltip(event, tooltipId) {
  const tooltip = document.getElementById(tooltipId);
  if (!tooltip || window.getComputedStyle(displayHelper).display === "none")
    return;

  // If this is already the active tooltip, do nothing
  if (activeTooltipId === tooltipId) {
    if (hideTimeout) {
      clearTimeout(hideTimeout);
      hideTimeout = null;
    }
    return;
  }

  // Hide any currently active tooltip immediately
  hideActiveTooltipImmediate();

  const trigger = event.target;
  activeTooltipId = tooltipId;
  activeTooltip = tooltip;

  // Store original position if not already stored
  if (!tooltip._originalParent) {
    tooltip._originalParent = tooltip.parentNode;
    tooltip._originalNextSibling = tooltip.nextSibling;
  }

  // Move to body and show
  document.body.appendChild(tooltip);
  tooltip.classList.remove("hidden");

  // Add event listeners
  tooltip.addEventListener("mouseenter", handleTooltipMouseEnter);
  tooltip.addEventListener("mouseleave", handleTooltipMouseLeave);

  // Setup positioning
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

// Hide tooltip with delay
function hideTooltip(tooltipId) {
  // Only hide if this is the currently active tooltip
  if (activeTooltipId !== tooltipId) return;
  
  // Clear any existing timeout
  if (hideTimeout) {
    clearTimeout(hideTimeout);
  }
  
  // Set new timeout
  hideTimeout = setTimeout(() => {
    hideActiveTooltipImmediate();
  }, 150);
}

function handleTooltipMouseEnter() {
  // Cancel hide timeout when mouse enters tooltip
  if (hideTimeout) {
    clearTimeout(hideTimeout);
    hideTimeout = null;
  }
}

function handleTooltipMouseLeave() {
  // Hide the currently active tooltip when mouse leaves
  if (activeTooltipId) {
    hideTooltip(activeTooltipId);
  }
}
