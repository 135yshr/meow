// Tab switching for code examples
document.addEventListener("DOMContentLoaded", function () {
  const tabs = document.querySelectorAll(".tab");
  const panels = document.querySelectorAll(".example-panel");

  tabs.forEach(function (tab) {
    tab.addEventListener("click", function () {
      var target = this.getAttribute("data-tab");

      tabs.forEach(function (t) { t.classList.remove("active"); });
      panels.forEach(function (p) { p.classList.remove("active"); });

      this.classList.add("active");
      var panel = document.getElementById("tab-" + target);
      if (panel) panel.classList.add("active");
    });
  });

  // Mobile nav toggle
  var toggle = document.querySelector(".nav-toggle");
  var navLinks = document.querySelector(".nav-links");
  if (toggle && navLinks) {
    toggle.addEventListener("click", function () {
      var isOpen = navLinks.classList.toggle("open");
      toggle.setAttribute("aria-expanded", isOpen ? "true" : "false");
    });
  }
});

// ── WASM lazy loading & example execution ──

var wasmReady = false;
var wasmLoading = false;

function getBaseURL() {
  var base = document.querySelector('base');
  if (base) return base.getAttribute('href');
  var link = document.querySelector('link[rel="canonical"]');
  if (link) {
    try {
      var url = new URL(link.getAttribute('href'));
      return url.pathname;
    } catch (e) {
      return '/meow/';
    }
  }
  return '/meow/';
}

function ensureWasm() {
  if (wasmReady) return Promise.resolve();
  if (wasmLoading) return wasmLoading;

  wasmLoading = new Promise(function (resolve, reject) {
    if (typeof Go === 'undefined') {
      reject(new Error('wasm_exec.js not loaded'));
      return;
    }
    var go = new Go();
    var wasmURL = getBaseURL() + 'playground/meow.wasm';
    WebAssembly.instantiateStreaming(fetch(wasmURL), go.importObject)
      .then(function (result) {
        go.run(result.instance);
        wasmReady = true;
        resolve();
      })
      .catch(function (err) {
        wasmLoading = false;
        reject(err);
      });
  });

  return wasmLoading;
}

function getCodeFromPanel(el) {
  var panel = el.closest('.example-panel');
  if (!panel) return '';
  var codeEl = panel.querySelector('.code-body code');
  return codeEl ? codeEl.textContent : '';
}

function runExample(btn) {
  var panel = btn.closest('.example-panel');
  if (!panel) return;
  var outputDiv = panel.querySelector('.code-run-output');
  if (!outputDiv) return;
  var outputPre = outputDiv.querySelector('pre');
  if (!outputPre) return;

  outputDiv.style.display = 'block';
  outputDiv.className = 'code-run-output loading';
  outputPre.textContent = 'Loading WASM...';

  var originalText = btn.textContent;
  btn.textContent = 'Loading...';
  btn.disabled = true;

  ensureWasm()
    .then(function () {
      btn.textContent = 'Running...';
      var code = getCodeFromPanel(btn);

      setTimeout(function () {
        try {
          var jsonResult = runMeow(code);
          var result = JSON.parse(jsonResult);
          outputDiv.className = 'code-run-output';

          if (result.error) {
            outputPre.textContent = result.output
              ? result.output + '\n--- Error ---\n' + result.error
              : result.error;
            if (!result.output) outputDiv.className = 'code-run-output error';
          } else {
            outputPre.textContent = result.output || '(no output)';
          }
        } catch (err) {
          outputDiv.className = 'code-run-output error';
          outputPre.textContent = 'Error: ' + err.message;
        }

        btn.textContent = originalText;
        btn.disabled = false;
      }, 10);
    })
    .catch(function (err) {
      outputDiv.className = 'code-run-output error';
      outputPre.textContent = 'Failed to load WASM: ' + err.message;
      btn.textContent = originalText;
      btn.disabled = false;
    });
}

function openInPlayground(link) {
  var code = getCodeFromPanel(link);
  var encoded = btoa(unescape(encodeURIComponent(code)));
  var url = getBaseURL() + 'playground/#code=' + encoded;
  window.open(url, '_blank');
}
