"use strict";

const editor = document.getElementById("editor");
const output = document.getElementById("output");
const runBtn = document.getElementById("run-btn");
const status = document.getElementById("status");
const examplesSelect = document.getElementById("examples");
const loading = document.getElementById("loading");
const loadingElapsed = document.getElementById("loading-elapsed");

const NARROW_VIEWPORT = "(max-width: 768px)";

let wasmReady = false;

function setStatus(text, state) {
    status.textContent = text;
    status.className = state ? "is-" + state : "";
}

// On a narrow screen the panels are stacked, so the Output panel can sit
// below the fold. Bring it into view once there is something to read.
function revealOutput() {
    if (!window.matchMedia || !window.matchMedia(NARROW_VIEWPORT).matches) return;
    const panel = output.parentElement;
    if (!panel || typeof panel.scrollIntoView !== "function") return;
    try {
        panel.scrollIntoView({ behavior: "smooth", block: "nearest" });
    } catch (e) {
        panel.scrollIntoView();
    }
}

// The wasm is served gzipped, so response.body yields decompressed bytes while
// content-length reports the compressed size: a byte counter built from those
// two numbers would be wrong. Reading the stream would also give up streaming
// compilation. So the indicator is honest and indeterminate, and only the
// elapsed seconds are counted.
function startElapsedTicker() {
    const started = Date.now();
    return window.setInterval(() => {
        if (!loadingElapsed) return;
        const seconds = Math.round((Date.now() - started) / 1000);
        if (seconds < 3) return;
        loadingElapsed.textContent = "Still downloading — " + seconds + "s so far.";
    }, 1000);
}

async function instantiateWasm(go) {
    const response = await fetch("meow.wasm");
    if (!response.ok) {
        throw new Error("HTTP " + response.status + " while fetching meow.wasm");
    }
    const contentType = response.headers.get("content-type") || "";
    if (typeof WebAssembly.instantiateStreaming === "function" &&
        contentType.indexOf("application/wasm") !== -1) {
        try {
            return await WebAssembly.instantiateStreaming(response, go.importObject);
        } catch (err) {
            // The body is spent; fall through to a fresh, non-streaming fetch.
            console.warn("Streaming instantiation failed, retrying without it:", err);
        }
    } else {
        const bytes = await response.arrayBuffer();
        return WebAssembly.instantiate(bytes, go.importObject);
    }
    const retry = await fetch("meow.wasm");
    if (!retry.ok) {
        throw new Error("HTTP " + retry.status + " while fetching meow.wasm");
    }
    return WebAssembly.instantiate(await retry.arrayBuffer(), go.importObject);
}

async function loadWasm() {
    const go = new Go();
    const ticker = startElapsedTicker();
    try {
        const result = await instantiateWasm(go);
        go.run(result.instance);
        wasmReady = true;
        runBtn.disabled = false;
        setStatus("Ready", "ready");
    } catch (err) {
        setStatus("Compiler failed to load", "error");
        output.textContent =
            "Failed to load the Meow compiler (" + err.message + ").\n\n" +
            "The playground needs to download a ~1.5 MB WebAssembly build. " +
            "Check your connection and reload the page.";
        output.className = "error";
    } finally {
        window.clearInterval(ticker);
        if (loading) loading.hidden = true;
    }
}

function run() {
    if (!wasmReady) {
        if (loading && !loading.hidden) {
            setStatus("Still downloading the compiler...", "busy");
        }
        return;
    }
    if (typeof runMeow !== "function") {
        output.textContent = "WASM not properly initialized";
        output.className = "error";
        return;
    }

    const source = editor.value;
    if (!source.trim()) {
        output.textContent = "";
        output.className = "";
        return;
    }

    setStatus("Running...", "busy");

    setTimeout(() => {
        try {
            const jsonResult = runMeow(source);
            const result = JSON.parse(jsonResult);

            if (result.error) {
                output.textContent = result.output
                    ? result.output + "\n--- Error ---\n" + result.error
                    : result.error;
                output.className = result.output ? "" : "error";
            } else {
                output.textContent = result.output || "(no output)";
                output.className = "";
            }
        } catch (err) {
            output.textContent = "Internal error: " + err.message;
            output.className = "error";
        }

        setStatus("Ready", "ready");
        revealOutput();
    }, 10);
}

runBtn.addEventListener("click", run);

document.addEventListener("keydown", (e) => {
    if ((e.ctrlKey || e.metaKey) && e.key === "Enter") {
        e.preventDefault();
        run();
    }
});

// Handle Tab key in editor
editor.addEventListener("keydown", (e) => {
    if (e.key === "Tab") {
        e.preventDefault();
        document.execCommand("insertText", false, "    ");
    }
});

// Populate examples dropdown
if (typeof MEOW_EXAMPLES !== "undefined") {
    MEOW_EXAMPLES.forEach((ex) => {
        const opt = document.createElement("option");
        opt.value = ex.name;
        opt.textContent = ex.name;
        examplesSelect.appendChild(opt);
    });
}

examplesSelect.addEventListener("change", () => {
    const name = examplesSelect.value;
    if (!name || typeof MEOW_EXAMPLES === "undefined") return;
    const ex = MEOW_EXAMPLES.find((e) => e.name === name);
    if (ex) {
        editor.value = ex.code;
        output.textContent = "";
        output.className = "";
    }
});

// Load code from URL hash or set default example
function loadFromHash() {
    var hash = location.hash;
    if (hash && hash.indexOf('#code=') === 0) {
        try {
            var encoded = hash.substring(6);
            var code = decodeURIComponent(escape(atob(encoded)));
            editor.value = code;
            return true;
        } catch (e) {
            // Fall through to default
        }
    }
    return false;
}

if (!loadFromHash()) {
    if (typeof MEOW_EXAMPLES !== "undefined" && MEOW_EXAMPLES.length > 0) {
        editor.value = MEOW_EXAMPLES[0].code;
    }
}

window.addEventListener("hashchange", loadFromHash);

loadWasm();
