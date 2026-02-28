(function () {
    "use strict";

    const photoInput = document.getElementById("photo-input");
    const captureSection = document.getElementById("capture-section");
    const previewSection = document.getElementById("preview-section");
    const previewImg = document.getElementById("preview-img");
    const analyzeBtn = document.getElementById("analyze-btn");
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

    function show(el) { el.classList.remove("hidden"); }
    function hide(el) { el.classList.add("hidden"); }

    function showError(msg) {
        errorBanner.textContent = msg;
        show(errorBanner);
    }

    function hideError() {
        hide(errorBanner);
    }

    photoInput.addEventListener("change", function () {
        if (!photoInput.files || !photoInput.files[0]) return;
        selectedFile = photoInput.files[0];
        previewImg.src = URL.createObjectURL(selectedFile);
        hide(captureSection);
        show(previewSection);
        hide(resultSection);
        hide(nfcStatus);
        hideError();
    });

    analyzeBtn.addEventListener("click", async function () {
        if (!selectedFile) return;
        hide(previewSection);
        show(loadingSection);
        hideError();

        try {
            const formData = new FormData();
            formData.append("image", selectedFile);

            const resp = await fetch("/api/analyze", {
                method: "POST",
                body: formData,
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

            hide(loadingSection);
            show(resultSection);
        } catch (err) {
            hide(loadingSection);
            show(previewSection);
            showError("Analysis failed: " + err.message);
        }
    });

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

            const ndef = new NDEFReader();
            const ctrl = new AbortController();

            // Wait for a tag to be in range, then write
            ndef.addEventListener("reading", async () => {
                try {
                    await ndef.write({ records: [{ recordType: "text", data: json }] });
                    ctrl.abort();
                    nfcMessage.textContent = "Tag written successfully!";
                } catch (writeErr) {
                    ctrl.abort();
                    hide(nfcStatus);
                    showError("NFC write failed: " + writeErr.message
                        + ". Make sure the tag is NDEF-formatted and not read-only.");
                }
            }, { once: true });

            await ndef.scan({ signal: ctrl.signal });
        } catch (err) {
            if (err.name === "AbortError") return;
            hide(nfcStatus);
            showError("NFC failed: " + err.message);
        }
    });
})();
