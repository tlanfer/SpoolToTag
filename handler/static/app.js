(function () {
    "use strict";

    const photoInput = document.getElementById("photo-input");
    const captureSection = document.getElementById("capture-section");
    const loadingSection = document.getElementById("loading-section");
    const resultSection = document.getElementById("result-section");
    const spoolForm = document.getElementById("spool-form");
    const errorBanner = document.getElementById("error-banner");
    const nfcStatus = document.getElementById("nfc-status");
    const nfcMessage = document.getElementById("nfc-message");

    const fieldType = document.getElementById("field-type");
    const fieldBrand = document.getElementById("field-brand");
    const fieldColorHex = document.getElementById("field-color-hex");
    const fieldColorPicker = document.getElementById("field-color-picker");
    const fieldMinTemp = document.getElementById("field-min-temp");
    const fieldMaxTemp = document.getElementById("field-max-temp");

    let selectedFile = null;
    let analyzeController = null;
    let nfcController = null;

    function show(el) { el.classList.remove("hidden"); }
    function hide(el) { el.classList.add("hidden"); }

    function showError(msg) {
        errorBanner.textContent = msg;
        show(errorBanner);
    }

    function hideError() {
        hide(errorBanner);
    }

    function resetToCapture() {
        if (analyzeController) { analyzeController.abort(); analyzeController = null; }
        if (nfcController) { nfcController.abort(); nfcController = null; }
        hide(loadingSection);
        hide(resultSection);
        hide(nfcStatus);
        hideError();
        photoInput.value = "";
        show(captureSection);
    }

    photoInput.addEventListener("change", async function () {
        if (!photoInput.files || !photoInput.files[0]) return;
        selectedFile = photoInput.files[0];
        hide(captureSection);
        hide(resultSection);
        hide(nfcStatus);
        hideError();
        show(loadingSection);

        analyzeController = new AbortController();

        try {
            const formData = new FormData();
            formData.append("image", selectedFile);

            const resp = await fetch("/api/analyze", {
                method: "POST",
                body: formData,
                signal: analyzeController.signal,
            });

            if (!resp.ok) {
                const text = await resp.text();
                throw new Error(text || "Analysis failed");
            }

            const data = await resp.json();
            fieldType.value = data.type || "";
            fieldBrand.value = data.brand || "";
            fieldColorHex.value = data.color_hex || "#000000";
            fieldColorPicker.value = data.color_hex || "#000000";
            fieldMinTemp.value = data.min_temp || "";
            fieldMaxTemp.value = data.max_temp || "";

            analyzeController = null;
            hide(loadingSection);
            show(resultSection);
        } catch (err) {
            analyzeController = null;
            hide(loadingSection);
            if (err.name === "AbortError") return;
            show(captureSection);
            showError("Analysis failed: " + err.message);
        }
    });

    document.getElementById("cancel-analyze-btn").addEventListener("click", resetToCapture);
    document.getElementById("retake-btn").addEventListener("click", resetToCapture);
    document.getElementById("nfc-done-btn").addEventListener("click", resetToCapture);

    // Sync color picker and hex input
    fieldColorHex.addEventListener("input", function () {
        if (/^#[0-9a-fA-F]{6}$/.test(fieldColorHex.value)) {
            fieldColorPicker.value = fieldColorHex.value;
        }
    });

    fieldColorPicker.addEventListener("input", function () {
        fieldColorHex.value = fieldColorPicker.value;
    });

    spoolForm.addEventListener("submit", async function (e) {
        e.preventDefault();
        hideError();
        hide(nfcStatus);

        const spoolData = {
            protocol: "openspool",
            version: "1.0",
            type: fieldType.value,
            color_hex: fieldColorHex.value,
            brand: fieldBrand.value,
            min_temp: parseInt(fieldMinTemp.value, 10),
            max_temp: parseInt(fieldMaxTemp.value, 10),
        };

        if (!("NDEFReader" in window)) {
            showError("Web NFC is not supported in this browser. Use Chrome on Android with NFC enabled.");
            return;
        }

        try {
            const json = JSON.stringify(spoolData);
            nfcMessage.textContent = "Hold your phone near the NFC tag...";
            show(nfcStatus);
            hide(resultSection);

            const ndef = new NDEFReader();
            nfcController = new AbortController();

            ndef.addEventListener("reading", async () => {
                try {
                    await ndef.write({ records: [{
                        recordType: "mime",
                        mediaType: "application/json",
                        data: new TextEncoder().encode(json),
                    }] });
                    nfcController.abort();
                    nfcController = null;
                    nfcMessage.textContent = "Tag written successfully!";
                } catch (writeErr) {
                    nfcController.abort();
                    nfcController = null;
                    hide(nfcStatus);
                    show(resultSection);
                    showError("NFC write failed: " + writeErr.message
                        + ". Make sure the tag is NDEF-formatted and not read-only.");
                }
            }, { once: true });

            await ndef.scan({ signal: nfcController.signal });
        } catch (err) {
            nfcController = null;
            if (err.name === "AbortError") return;
            hide(nfcStatus);
            show(resultSection);
            showError("NFC failed: " + err.message);
        }
    });
})();
