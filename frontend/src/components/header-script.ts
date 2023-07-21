const TO_DARK_HTML = "<img src=\"/moon.svg\">";
const TO_LIGHT_HTML = "<img src=\"/sun.svg\">";

function updateThemeVariant(target: "dark" | "light") {
  document.documentElement.setAttribute("data-theme", target);
  localStorage.setItem("theme", target);
}

function getThemeVariant(): "dark" | "light" {
  let theme = localStorage.getItem("theme");

  if (!theme) {
    if (!window.matchMedia) {
      theme = "light";
    } else if (window.matchMedia("(prefers-color-scheme: dark)").matches) {
      theme = "dark";
    }
  }

  if (theme != "dark") {
    theme = "light";
  }

  return theme as "dark" | "light";
}

const button = document.getElementById("themeswap.button");

const initial = getThemeVariant();

if (initial == "dark") {
  button!.innerHTML = TO_LIGHT_HTML;
} else if (initial == "light") {
  button!.innerHTML = TO_DARK_HTML;
}

button?.addEventListener("click", () => {
  const current = getThemeVariant();

  if (current == "dark") {
    button.innerHTML = TO_DARK_HTML;
    updateThemeVariant("light");
  } else if (current == "light") {
    button.innerHTML = TO_LIGHT_HTML;
    updateThemeVariant("dark");
  }
});
