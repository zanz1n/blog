import { getTheme, setTheme, storeTheme, Theme } from "@lib/theming";

const currtheme = getTheme();
setTheme(currtheme);

document.addEventListener("DOMContentLoaded", () => {
    const lightswitch = document.getElementById(
        "lightswitch",
    ) as HTMLInputElement | null;

    if (lightswitch) {
        lightswitch.checked = currtheme == Theme.Light;
    }

    lightswitch.addEventListener("change", () => {
        const theme = lightswitch.checked ? Theme.Light : Theme.Dark;
        setTheme(theme);
        storeTheme(theme);
    });
});
