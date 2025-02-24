export enum Theme {
    Light = "light",
    Dark = "dark",
}

export function getTheme(): Theme {
    const theme = localStorage.getItem("theme");

    if (theme == Theme.Light || theme == Theme.Dark) {
        return theme;
    }

    const isDark = window.matchMedia("(prefers-color-scheme: dark)").matches;

    if (isDark) {
        return Theme.Dark;
    } else {
        return Theme.Light;
    }
}

export function storeTheme(variant: Theme) {
    localStorage.setItem("theme", variant);
}

export function setTheme(variant: Theme) {
    document.querySelector("html").setAttribute("data-theme", variant);
}
