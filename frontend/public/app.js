// Intersection Observer for reveal animations
const observer = new IntersectionObserver(
  (entries) => {
    entries.forEach((entry) => {
      if (entry.isIntersecting) {
        entry.target.classList.add("is-visible");
      }
    });
  },
  { threshold: 0.18 }
);

document.querySelectorAll(".reveal").forEach((node) => observer.observe(node));

// Mouse follow glow effect
let mouseX = 0;
let mouseY = 0;
let currentX = 0;
let currentY = 0;

document.addEventListener("mousemove", (e) => {
  mouseX = e.clientX;
  mouseY = e.clientY;
});

function updateCursor() {
  const dx = mouseX - currentX;
  const dy = mouseY - currentY;

  currentX += dx * 0.1;
  currentY += dy * 0.1;

  document.documentElement.style.setProperty("--cursor-x", `${currentX}px`);
  document.documentElement.style.setProperty("--cursor-y", `${currentY}px`);

  requestAnimationFrame(updateCursor);
}

updateCursor();

// Show cursor glow on mouse move
let glowTimeout;
document.addEventListener("mousemove", () => {
  document.body.style.setProperty("--glow-opacity", "0.15");

  clearTimeout(glowTimeout);
  glowTimeout = setTimeout(() => {
    document.body.style.setProperty("--glow-opacity", "0");
  }, 2000);
});

// Copy code functionality
document.querySelectorAll("code").forEach((codeBlock) => {
  const wrapper = document.createElement("div");
  wrapper.style.position = "relative";
  codeBlock.parentNode.insertBefore(wrapper, codeBlock);
  wrapper.appendChild(codeBlock);

  const copyBtn = document.createElement("button");
  copyBtn.textContent = "Copy";
  copyBtn.style.cssText = `
    position: absolute;
    top: 8px;
    right: 8px;
    padding: 6px 12px;
    background: rgba(128, 255, 211, 0.1);
    border: 1px solid rgba(128, 255, 211, 0.3);
    border-radius: 6px;
    color: var(--accent);
    font-size: 0.75rem;
    cursor: pointer;
    opacity: 0;
    transition: opacity 0.2s ease, background 0.2s ease;
    font-family: inherit;
  `;

  wrapper.addEventListener("mouseenter", () => {
    copyBtn.style.opacity = "1";
  });

  wrapper.addEventListener("mouseleave", () => {
    copyBtn.style.opacity = "0";
  });

  copyBtn.addEventListener("click", async () => {
    try {
      await navigator.clipboard.writeText(codeBlock.textContent);
      copyBtn.textContent = "Copied!";
      copyBtn.style.background = "rgba(128, 255, 211, 0.2)";

      setTimeout(() => {
        copyBtn.textContent = "Copy";
        copyBtn.style.background = "rgba(128, 255, 211, 0.1)";
      }, 2000);
    } catch (err) {
      console.error("Failed to copy:", err);
    }
  });

  wrapper.appendChild(copyBtn);
});

// Parallax effect for hero visual
const heroVisual = document.querySelector(".hero-visual");
if (heroVisual) {
  window.addEventListener("scroll", () => {
    const scrolled = window.pageYOffset;
    const rate = scrolled * 0.3;
    heroVisual.style.transform = `translateY(${rate}px)`;
  });
}
