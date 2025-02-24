import "bun";
import { watch } from "chokidar";
import * as build from "./build";
import { Stats } from "fs";

const devmode = true;
const dir = process.env.DIR ?? "./dist";

const paths = ["./src", "./templates"];
const jsExts = ["ts", "js"];
const cssExts = ["css", "templ"];

async function buildCss() {
    const start = Date.now();
    const res = await build.buildCss(dir, devmode);
    const took = Date.now() - start;
    return res;
}

async function buildJs() {
    const start = Date.now();
    const res = await build.buildJs(dir, devmode);
    const took = Date.now() - start;
    return res;
}

async function handleChange(path: string, stats?: Stats) {
    if (!stats || stats.isDirectory()) {
        await Promise.all([buildCss(), buildJs()]);
        return;
    }

    if (stats.isFile()) {
        const s = path?.split(".");
        const ext = (s?.length ?? 0 >= 1) ? (s?.[s.length - 1] ?? "") : "";

        if (cssExts.includes(ext)) {
            await buildCss();
        }

        if (jsExts.includes(ext)) {
            await buildJs();
        }
    }
}

function handler(path: string, stats?: Stats) {
    if (stats?.isFile()) {
        console.log("Changed file", path);
    } else {
        console.log("Changed", path);
    }
    handleChange(path, stats).catch((_) => {});
}

if (import.meta.main) {
    await Promise.all([buildCss(), buildJs()]);

    const watcher = watch(paths, {
        usePolling: true,
        interval: 2000,
        alwaysStat: true,
    });

    watcher.on("change", handler);
}
