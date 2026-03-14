"use strict";

(function () {
  var CONSENT_KEY = "meow-analytics-consent";

  window.dataLayer = window.dataLayer || [];
  function gtag() { dataLayer.push(arguments); }
  window.gtag = gtag;

  // Default: deny analytics until user consents
  gtag("consent", "default", {
    analytics_storage: "denied",
  });

  var saved = localStorage.getItem(CONSENT_KEY);
  if (saved === "accepted") {
    grantConsent();
  } else if (saved === "rejected") {
    // Stay denied, don't show banner again
  } else {
    showBanner();
  }

  function grantConsent() {
    gtag("consent", "update", { analytics_storage: "granted" });
    gtag("js", new Date());
    gtag("config", "G-081QKVSS83");
  }

  function showBanner() {
    var banner = document.getElementById("meow-consent-banner");
    if (banner) banner.hidden = false;
  }

  document.addEventListener("DOMContentLoaded", function () {
    var acceptBtn = document.getElementById("meow-consent-accept");
    var rejectBtn = document.getElementById("meow-consent-reject");
    var banner = document.getElementById("meow-consent-banner");

    if (acceptBtn) {
      acceptBtn.addEventListener("click", function () {
        localStorage.setItem(CONSENT_KEY, "accepted");
        grantConsent();
        if (banner) banner.hidden = true;
      });
    }

    if (rejectBtn) {
      rejectBtn.addEventListener("click", function () {
        localStorage.setItem(CONSENT_KEY, "rejected");
        if (banner) banner.hidden = true;
      });
    }
  });
})();
