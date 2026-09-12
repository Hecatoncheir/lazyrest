const copyButton = document.querySelector("[data-copy]");

if (copyButton) {
  copyButton.addEventListener("click", async () => {
    const label = copyButton.querySelector("span");
    try {
      await navigator.clipboard.writeText(copyButton.dataset.copy);
      label.textContent = "COPIED";
      setTimeout(() => { label.textContent = "COPY"; }, 1800);
    } catch {
      label.textContent = "SELECT";
    }
  });
}
