import "bun";
import { build, Metafile } from "esbuild";
import * as path from "path";
import * as fs from "fs/promises";
import { exec as childExec } from "child_process";

type ExecResult = {
    stdout: string;
    stderr: string;
};

function exec(command: string): Promise<ExecResult> {
    return new Promise((resolve, reject) =>
        childExec(command, (error, stdout, stderr) => {
            if (error) {
                reject(error);
            } else {
                resolve({ stdout, stderr });
            }
        }),
    );
}

export async function buildJs(dir: string, dev: boolean) {
    console.time(".ts Build");

    const outdir = path.join(dir, "js");

    await fs.rm(outdir, { recursive: true, force: true });

    const entryPoints = (await fs.readdir("./src/entries"))
        .filter((s) => s.endsWith(".ts"))
        .map((s) => path.join("./src/entries", s));

    const res = await build({
        entryPoints,
        outdir,
        entryNames: dev ? "[dir]/[name]" : "[dir]/[name]-[hash]",
        platform: "browser",
        format: "esm",
        bundle: true,
        packages: "bundle",
        tsconfig: "tsconfig.json",
        legalComments: "eof",
        splitting: dev,
        minify: !dev,
        target: dev ? "es2022" : "es2017",
        metafile: !dev,
    });

    console.timeEnd(".ts Build");

    return res;
}

export async function buildCss(dir: string, dev: boolean) {
    console.time(".css Build");

    const outdir = path.join(dir, "css");

    await fs.rm(outdir, { recursive: true, force: true });

    await exec(`bun tailwindcss -i ./src/app.css -o ${outdir}/styles.css`);

    const entryPoints = (await fs.readdir(outdir))
        .filter((s) => s.endsWith(".css"))
        .map((s) => path.join(outdir, s));

    const res = await build({
        entryPoints,
        outdir,
        entryNames: dev ? "[dir]/[name]" : "[dir]/[name]-[hash]",
        allowOverwrite: true,
        bundle: true,
        legalComments: "eof",
        minify: !dev,
        metafile: !dev,
    });

    if (!dev) {
        await fs.rm(path.join(outdir, "styles.css")).catch(console.error);
    }

    console.timeEnd(".css Build");

    return res;
}

type SourceMap = {
    [key: string]: {
        url: string;
        integrity: string;
    };
};

async function hashFile(path: string | URL) {
    const file = await Bun.file(path).arrayBuffer();
    const integrityBuf = Bun.SHA256.hash(file) as Uint8Array;
    const integrity = Buffer.from(integrityBuf).toString("base64");

    return integrity;
}

async function sourceMap(
    meta: Metafile,
    entries: string,
    dir: string,
    ext: string,
): Promise<SourceMap> {
    console.time(`${ext} Source maps`);

    dir = dir.replace("./", "");
    entries = entries.replace("./", "");
    const sourcemap: SourceMap = {};

    let size = 0;

    for (const out in meta.outputs) {
        const entryPoint = meta.outputs[out].entryPoint;
        if (!entryPoint) {
            continue;
        }
        const pp = path.parse(entryPoint);

        if (pp.dir == entries && pp.ext == ext) {
            size += meta.outputs[out].bytes;

            const integrity = await hashFile(out);

            sourcemap[pp.name] = {
                url: out.replace(dir, ""),
                integrity: "sha256-" + integrity,
            };
        }
    }

    console.log(`${ext} total size: ${(size / 1024).toFixed(2)} KiB`);

    console.timeEnd(`${ext} Source maps`);

    return sourcemap;
}

if (import.meta.main) {
    console.time("Build complete");

    const dev =
        process.env.NODE_ENV == "dev" ||
        process.env.NODE_ENV == "development" ||
        process.env.DEBUG == "1" ||
        process.env.DEBUG == "true";

    const dir = process.env.DIR ?? "./dist";

    if (!dev) {
        const [js, css] = await Promise.all([
            buildJs(dir, dev).then((js) =>
                sourceMap(js.metafile!, "./src/entries", dir, ".ts"),
            ),
            buildCss(dir, dev).then((css) =>
                sourceMap(css.metafile!, path.join(dir, "css"), dir, ".css"),
            ),
        ]);

        const sourcemap = { js, css };
        await fs.writeFile(".source-map.json", JSON.stringify(sourcemap));
    } else {
        await Promise.all([buildJs(dir, dev), buildCss(dir, dev)]);
    }

    console.timeEnd("Build complete");
}
