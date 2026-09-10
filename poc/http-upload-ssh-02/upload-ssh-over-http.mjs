#!/usr/bin/env node
import { createHash } from "node:crypto";
import { readdir, readFile, stat } from "node:fs/promises";
import { request as httpRequest } from "node:http";
import { request as httpsRequest } from "node:https";
import { dirname, join, relative, resolve, sep } from "node:path";
import { fileURLToPath } from "node:url";

const args = process.argv.slice(2);
const host = (process.env.CC2_HOST || args.find((arg) => arg.startsWith("--host="))?.slice(7) || "").replace(/\/$/, "");
const token = process.env.CC2_TOKEN || args.find((arg) => arg.startsWith("--token="))?.slice(8) || "";
const filesRoot = resolve(dirname(fileURLToPath(import.meta.url)), "files");

if (!host || !token) {
    console.error("usage: CC2_HOST=http://192.168.x.x CC2_TOKEN=... node upload-ssh-over-http.mjs");
    console.error("or:    node upload-ssh-over-http.mjs --host=http://192.168.x.x --token=...");
    process.exit(1);
}

async function collectFiles(dir) {
    const entries = [];
    for (const item of await readdir(dir, { withFileTypes: true })) {
        const path = join(dir, item.name);
        if (item.isDirectory()) {
            entries.push(...(await collectFiles(path)));
        } else if (item.isFile()) {
            entries.push(path);
        }
    }
    return entries;
}

function uploadName(targetPath) {
    return encodeURIComponent(`../../${targetPath.replace(/^\/+/, "")}`);
}

async function upload(targetPath, body) {
    const url = new URL("/upload/udisk", host);
    const request = url.protocol === "https:" ? httpsRequest : httpRequest;
    const responseBody = await new Promise((resolvePromise, reject) => {
        const req = request(url, {
            method: "POST",
            headers: {
                "X-Token": token,
                "X-File-Name": uploadName(targetPath),
                "X-File-MD5": createHash("md5").update(body).digest("hex"),
                "Content-Range": `bytes 0-${body.length - 1}/${body.length}`,
                "Content-Type": "application/octet-stream",
                "Content-Length": body.length,
            },
        }, (res) => {
            const chunks = [];
            res.on("data", (chunk) => chunks.push(chunk));
            res.on("end", () => {
                const text = Buffer.concat(chunks).toString("utf8");
                if (res.statusCode !== 200) {
                    reject(new Error(`${targetPath} upload failed: HTTP ${res.statusCode} ${text}`));
                    return;
                }
                resolvePromise(text);
            });
        });
        req.on("error", reject);
        req.end(body);
    });

    try {
        const data = JSON.parse(responseBody);
        if (data.error_code !== 0) {
            throw new Error(`${targetPath} upload failed: ${responseBody}`);
        }
    } catch (error) {
        if (error instanceof SyntaxError) {
            throw new Error(`${targetPath} returned non-json response: ${responseBody}`);
        }
        throw error;
    }
}

const files = (await collectFiles(filesRoot)).sort((a, b) => {
    const ar = relative(filesRoot, a).split(sep).join("/");
    const br = relative(filesRoot, b).split(sep).join("/");
    if (ar === "etc/rc.local") return 1;
    if (br === "etc/rc.local") return -1;
    return ar.localeCompare(br);
});

for (const file of files) {
    const targetPath = `/${relative(filesRoot, file).split(sep).join("/")}`;
    const body = await readFile(file);
    if ((await stat(file)).size !== body.length) {
        throw new Error(`read size mismatch: ${file}`);
    }
    process.stdout.write(`upload ${targetPath} (${body.length} bytes) ... `);
    await upload(targetPath, body);
    process.stdout.write("ok\n");
}

console.log("done. reboot the printer, then ssh should be started by /etc/rc.local.");
