// Long-string breaker
document.querySelectorAll('.wsr td:nth-child(2)').forEach(el => {
  el.innerHTML = el.innerHTML.replace(/[-_]/g, '$&<wbr>');
});

// Windows key toggler
document.addEventListener("DOMContentLoaded", () => {
  const stars = "***********";
  const lockOpen = "🔓";
  const lockClosed = "🔒";

  const rows = document.querySelectorAll(".wsr-box table tr");

  rows.forEach(row => {
    const keyCell = row.cells[0]; // first td of this row

    if (keyCell && keyCell.textContent.trim() === "OriginalProductKey") {
      const nextCell = row.cells[1];

      if (nextCell) {
        const winKeyVal = nextCell.textContent.trim();

        nextCell.innerHTML = ""; // clear existing content first

        const winKeyBox = document.createElement("span");
        winKeyBox.classList.add(".winkey", ".winkey-box");
        winKeyBox.textContent = stars;

        const winKeyBtn = document.createElement("span");
        winKeyBtn.classList.add("winkey", "winkey-btn");
        winKeyBtn.textContent = lockClosed;

        nextCell.appendChild(winKeyBox); // then append the span
        nextCell.appendChild(winKeyBtn);

        if (winKeyBox && winKeyBtn) {
          winKeyBtn.addEventListener("click", () => {
            if (winKeyBtn.textContent.trim() === lockClosed) {
              winKeyBox.textContent = winKeyVal;
              winKeyBtn.textContent = lockOpen;
            } else {
              winKeyBox.textContent = stars;
              winKeyBtn.textContent = lockClosed;
            }
          });
        }
      }
    }
  });
});
