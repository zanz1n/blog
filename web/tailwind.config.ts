import daisyui, { Config as DaisyUiConfig } from "daisyui";
import { Config } from "tailwindcss";

const daisyuicfg = {
    themes: ["light", "dark"],
} satisfies DaisyUiConfig;

export default {
    content: ["templates/**/*.templ"],
    plugins: [daisyui],
    daisyui: daisyuicfg,
} satisfies Config;
