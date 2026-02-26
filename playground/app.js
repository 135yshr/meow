"use strict";

const editor = document.getElementById("editor");
const output = document.getElementById("output");
const runBtn = document.getElementById("run-btn");
const status = document.getElementById("status");
const examplesSelect = document.getElementById("examples");

let wasmReady = false;

async function loadWasm() {
    const go = new Go();
    try {
        const result = await WebAssembly.instantiateStreaming(
            fetch("meow.wasm"),
            go.importObject
        );
        go.run(result.instance);
        wasmReady = true;
        runBtn.disabled = false;
        status.textContent = "Ready";
        status.style.color = "#4caf50";
    } catch (err) {
        status.textContent = "Failed to load WASM: " + err.message;
        status.style.color = "#e94560";
    }
}

function run() {
    if (!wasmReady) return;
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

    status.textContent = "Running...";
    status.style.color = "#ff9800";

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

        status.textContent = "Ready";
        status.style.color = "#4caf50";
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
