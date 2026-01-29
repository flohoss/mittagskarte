const { computePosition, offset, flip, shift, autoUpdate } = window.FloatingUIDOM;

const tooltipState = {
  activeId: null,
  hideTimeout: null,
  activeElement: null,

  reset() {
    this.activeId = null;
    this.hideTimeout = null;
    this.activeElement = null;
  },

  clearHideTimeout() {
    if (this.hideTimeout) {
      clearTimeout(this.hideTimeout);
      this.hideTimeout = null;
    }
  },
};

const elements = {
  links: document.querySelectorAll('[data-lg-blank]'),
  displayHelper: document.querySelector('#display-helper'),
};

function isLargeScreen() {
  return window.getComputedStyle(elements.displayHelper).display !== 'none';
}

function setLinkTargets() {
  const shouldOpenInNewTab = isLargeScreen();
  elements.links.forEach((link) => {
    if (shouldOpenInNewTab) {
      link.setAttribute('target', '_blank');
    } else {
      link.removeAttribute('target');
    }
  });
}

function cleanupTooltipPositioning(tooltip) {
  if (tooltip._cleanup) {
    try {
      tooltip._cleanup();
    } catch (error) {
      console.warn('Error cleaning up tooltip positioning:', error);
    }
    tooltip._cleanup = null;
  }
}

function restoreTooltipPosition(tooltip) {
  if (tooltip._originalParent) {
    if (tooltip._originalNextSibling) {
      tooltip._originalParent.insertBefore(tooltip, tooltip._originalNextSibling);
    } else {
      tooltip._originalParent.appendChild(tooltip);
    }
  }
}

function removeTooltipEventListeners(tooltip) {
  tooltip.removeEventListener('mouseenter', handleTooltipMouseEnter);
  tooltip.removeEventListener('mouseleave', handleTooltipMouseLeave);
}

function hideActiveTooltipImmediate() {
  if (!tooltipState.activeElement || !tooltipState.activeId) return;

  const tooltip = tooltipState.activeElement;

  tooltipState.clearHideTimeout();
  removeTooltipEventListeners(tooltip);
  restoreTooltipPosition(tooltip);
  tooltip.classList.add('hidden');
  cleanupTooltipPositioning(tooltip);
  tooltipState.reset();
}

function isTooltipAlreadyActive(tooltipId) {
  return tooltipState.activeId === tooltipId;
}

function shouldShowTooltip(tooltip) {
  return tooltip && isLargeScreen();
}

function storeOriginalTooltipPosition(tooltip) {
  if (!tooltip._originalParent) {
    tooltip._originalParent = tooltip.parentNode;
    tooltip._originalNextSibling = tooltip.nextSibling;
  }
}

function addTooltipEventListeners(tooltip) {
  tooltip.addEventListener('mouseenter', handleTooltipMouseEnter);
  tooltip.addEventListener('mouseleave', handleTooltipMouseLeave);
}

function setupTooltipPositioning(trigger, tooltip) {
  tooltip._cleanup = autoUpdate(trigger, tooltip, () => {
    computePosition(trigger, tooltip, {
      placement: 'right',
      middleware: [offset(8), flip(), shift({ padding: 8 })],
    })
      .then(({ x, y }) => {
        Object.assign(tooltip.style, {
          position: 'absolute',
          left: `${x}px`,
          top: `${y}px`,
        });
      })
      .catch((error) => {
        console.warn('Error positioning tooltip:', error);
      });
  });
}

function showTooltip(event, tooltipId) {
  const tooltip = document.getElementById(tooltipId);

  if (!shouldShowTooltip(tooltip)) return;

  if (isTooltipAlreadyActive(tooltipId)) {
    tooltipState.clearHideTimeout();
    return;
  }

  hideActiveTooltipImmediate();

  const trigger = event.target;
  tooltipState.activeId = tooltipId;
  tooltipState.activeElement = tooltip;

  storeOriginalTooltipPosition(tooltip);
  document.body.appendChild(tooltip);
  tooltip.classList.remove('hidden');
  addTooltipEventListeners(tooltip);
  setupTooltipPositioning(trigger, tooltip);
}

function hideTooltip(tooltipId) {
  if (tooltipState.activeId !== tooltipId) return;

  tooltipState.clearHideTimeout();
  tooltipState.hideTimeout = setTimeout(() => {
    hideActiveTooltipImmediate();
  }, 150);
}

function handleTooltipMouseEnter() {
  tooltipState.clearHideTimeout();
}

function handleTooltipMouseLeave() {
  if (tooltipState.activeId) {
    hideTooltip(tooltipState.activeId);
  }
}

setLinkTargets();
window.addEventListener('resize', setLinkTargets);
