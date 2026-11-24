(function () {
  "use strict";

  const STORAGE_KEY = "f2md_android_prefs_v1";

  const state = {
    source: null, // { type: "folder" | "zip", name: string, sizeBytes: number }
    packing: false,
    format: "md",
    sortMode: "name",
    sortAsc: false,
    useFolderName: true,
    includeDate: false,
    includeTime: false,
    maxSizeKb: 500,
    skipLocks: true,
    result: null
  };

  const els = {
    versionPill: document.getElementById("versionPill"),
    sourceStatusBadge: document.getElementById("sourceStatusBadge"),
    btnPickFolder: document.getElementById("btnPickFolder"),
    btnPickZip: document.getElementById("btnPickZip"),
    selectedBanner: document.getElementById("selectedBanner"),
    selectedTypeTag: document.getElementById("selectedTypeTag"),
    selectedName: document.getElementById("selectedName"),
    selectedSub: document.getElementById("selectedSub"),
    filenamePreview: document.getElementById("filenamePreview"),
    btnSortDir: document.getElementById("btnSortDir"),
    sortDirLabel: document.getElementById("sortDirLabel"),
    inputMaxSize: document.getElementById("inputMaxSize"),
    chkFolderName: document.getElementById("chkFolderName"),
    chkDate: document.getElementById("chkDate"),
    chkTime: document.getElementById("chkTime"),
    chkSkipLocks: document.getElementById("chkSkipLocks"),
    btnPack: document.getElementById("btnPack"),
    packSpinner: document.getElementById("packSpinner"),
    packIcon: document.getElementById("packIcon"),
    packBtnText: document.getElementById("packBtnText"),
    progressLine: document.getElementById("progressLine"),
    resultPanel: document.getElementById("resultPanel"),
    resultFilename: document.getElementById("resultFilename"),
    resultFormatBadge: document.getElementById("resultFormatBadge"),
    metricFiles: document.getElementById("metricFiles"),
    metricSize: document.getElementById("metricSize"),
    metricTokens: document.getElementById("metricTokens"),
    metricTime: document.getElementById("metricTime"),
    skipSummary: document.getElementById("skipSummary"),
    btnShare: document.getElementById("btnShare"),
    btnSave: document.getElementById("btnSave"),
    btnCopy: document.getElementById("btnCopy"),
    btnCopyTree: document.getElementById("btnCopyTree"),
    tabFileCount: document.getElementById("tabFileCount"),
    treeOutput: document.getElementById("treeOutput"),
    fileFilterInput: document.getElementById("fileFilterInput"),
    fileList: document.getElementById("fileList"),
    previewHint: document.getElementById("previewHint"),
    previewOutput: document.getElementById("previewOutput"),
    toast: document.getElementById("toast")
  };

  let toastTimer = null;

  function showToast(message, level) {
    els.toast.textContent = message;
    els.toast.className = "toast visible " + (level || "info");
    if (toastTimer) clearTimeout(toastTimer);
    toastTimer = setTimeout(() => {
      els.toast.classList.remove("visible");
    }, 3200);
  }

  function formatBytes(bytes) {
    if (bytes < 0) return "";
    if (bytes < 1024) return bytes + " B";
    if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + " KB";
    return (bytes / (1024 * 1024)).toFixed(2) + " MB";
  }

  function formatTokens(n) {
    if (n >= 1000000) return "~" + (n / 1000000).toFixed(1) + "M";
    if (n >= 1000) return "~" + (n / 1000).toFixed(1) + "k";
    return "~" + n;
  }

  function savePrefs() {
    try {
      localStorage.setItem(
        STORAGE_KEY,
        JSON.stringify({
          format: state.format,
          sortMode: state.sortMode,
          sortAsc: state.sortAsc,
          useFolderName: state.useFolderName,
          includeDate: state.includeDate,
          includeTime: state.includeTime,
          maxSizeKb: state.maxSizeKb,
          skipLocks: state.skipLocks
        })
      );
    } catch (_) {}
  }

  function loadPrefs() {
    try {
      const raw = localStorage.getItem(STORAGE_KEY);
      if (!raw) return;
      const p = JSON.parse(raw);
      if (p.format === "md" || p.format === "txt") state.format = p.format;
      if (["name", "modified", "size", "created"].includes(p.sortMode)) state.sortMode = p.sortMode;
      if (typeof p.sortAsc === "boolean") state.sortAsc = p.sortAsc;
      if (typeof p.useFolderName === "boolean") state.useFolderName = p.useFolderName;
      if (typeof p.includeDate === "boolean") state.includeDate = p.includeDate;
      if (typeof p.includeTime === "boolean") state.includeTime = p.includeTime;
      if (typeof p.maxSizeKb === "number" && p.maxSizeKb >= 0) state.maxSizeKb = p.maxSizeKb;
      if (typeof p.skipLocks === "boolean") state.skipLocks = p.skipLocks;
    } catch (_) {}
  }

  function computePreviewFilename() {
    const now = new Date();
    const pad = (v) => String(v).padStart(2, "0");
    const parts = [];

    let root = "project";
    if (state.useFolderName && state.source && state.source.name) {
      root = state.source.name.replace(/\.zip$/i, "").replace(/[/\\:*?"<>|]+/g, "_").replace(/^_+|_+$/g, "") || "project";
    }
    parts.push(root);

    if (state.includeDate) {
      parts.push(`${now.getFullYear()}-${pad(now.getMonth() + 1)}-${pad(now.getDate())}`);
    }
    if (state.includeTime) {
      parts.push(`${pad(now.getHours())}-${pad(now.getMinutes())}-${pad(now.getSeconds())}`);
    }
    parts.push("context." + state.format);
    return parts.join("_");
  }

  function syncControlsUI() {
    document.querySelectorAll(".seg-btn").forEach((btn) => {
      const active = btn.dataset.format === state.format;
      btn.classList.toggle("active", active);
      btn.setAttribute("aria-pressed", String(active));
    });

    document.querySelectorAll("[data-sort]").forEach((btn) => {
      const active = btn.dataset.sort === state.sortMode;
      btn.classList.toggle("active", active);
      btn.setAttribute("aria-pressed", String(active));
    });

    if (state.sortMode === "name") {
      els.btnSortDir.classList.add("hidden");
    } else {
      els.btnSortDir.classList.remove("hidden");
      const label = state.sortMode === "size"
        ? (state.sortAsc ? "Asc (Smallest)" : "Desc (Largest)")
        : (state.sortAsc ? "Asc (Oldest)" : "Desc (Newest)");
      els.sortDirLabel.textContent = label;
    }

    els.inputMaxSize.value = String(state.maxSizeKb);
    document.querySelectorAll(".size-chip").forEach((btn) => {
      btn.classList.toggle("active", Number(btn.dataset.size) === state.maxSizeKb);
    });

    els.chkFolderName.checked = state.useFolderName;
    els.chkDate.checked = state.includeDate;
    els.chkTime.checked = state.includeTime;
    els.chkSkipLocks.checked = state.skipLocks;

    els.filenamePreview.textContent = computePreviewFilename();
    syncPackButton();
  }

  function syncPackButton() {
    if (state.packing) {
      els.btnPack.disabled = true;
      els.packSpinner.classList.remove("hidden");
      els.packIcon.classList.add("hidden");
      els.packBtnText.textContent = "Packing...";
      return;
    }
    els.packSpinner.classList.add("hidden");
    els.packIcon.classList.remove("hidden");
    if (!state.source) {
      els.btnPack.disabled = true;
      els.packBtnText.textContent = "Select a Folder or ZIP to Pack";
    } else {
      els.btnPack.disabled = false;
      const kind = state.source.type === "zip" ? "ZIP Archive" : "Folder";
      els.packBtnText.textContent = `Pack ${kind} to .${state.format}`;
    }
  }

  function applySource(src) {
    state.source = src;
    els.btnPickFolder.classList.toggle("selected", src.type === "folder");
    els.btnPickZip.classList.toggle("selected-zip", src.type === "zip");

    els.sourceStatusBadge.textContent = src.type === "zip" ? "ZIP Ready" : "Folder Ready";
    els.sourceStatusBadge.className = "badge " + (src.type === "zip" ? "badge-amber" : "badge-emerald");

    els.selectedBanner.classList.remove("hidden");
    els.selectedTypeTag.textContent = src.type === "zip" ? "ZIP" : "DIR";
    els.selectedTypeTag.className = "source-type-tag " + src.type;
    els.selectedName.textContent = src.name;
    els.selectedSub.textContent = src.type === "zip"
      ? (src.sizeBytes > 0 ? `Archive · ${formatBytes(src.sizeBytes)}` : "ZIP archive selected")
      : "Android SAF directory selected";

    els.filenamePreview.textContent = computePreviewFilename();
    syncPackButton();
  }

  function renderFileList(query) {
    const files = (state.result && state.result.files) || [];
    const q = (query || "").trim().toLowerCase();
    const filtered = q ? files.filter((f) => f.relPath.toLowerCase().includes(q) || f.lang.toLowerCase().includes(q)) : files;

    els.fileList.innerHTML = "";
    const frag = document.createDocumentFragment();
    for (const f of filtered) {
      const li = document.createElement("li");
      li.className = "file-item" + (f.skipped ? " skipped" : "");

      const pathSpan = document.createElement("span");
      pathSpan.className = "file-item-path";
      pathSpan.textContent = f.relPath;
      pathSpan.title = f.relPath;

      const meta = document.createElement("span");
      meta.className = "file-item-meta";

      const tag = document.createElement("span");
      if (f.skipped) {
        tag.className = "skip-tag";
        tag.textContent = f.reason || "skipped";
      } else {
        tag.className = "lang-tag";
        tag.textContent = f.lang || "text";
      }

      const sz = document.createElement("span");
      sz.className = "size-tag";
      sz.textContent = formatBytes(f.sizeBytes);

      meta.appendChild(tag);
      meta.appendChild(sz);
      li.appendChild(pathSpan);
      li.appendChild(meta);
      frag.appendChild(li);
    }

    if (filtered.length === 0) {
      const empty = document.createElement("li");
      empty.className = "file-item";
      empty.innerHTML = '<span class="file-item-path" style="color:var(--text-dim)">No matching files</span>';
      frag.appendChild(empty);
    }

    els.fileList.appendChild(frag);
  }

  // Native Callbacks
  window.onSafeAreaInsets = function (topPx, bottomPx) {
    document.documentElement.style.setProperty("--safe-top", (topPx || 0) + "px");
    document.documentElement.style.setProperty("--safe-bottom", (bottomPx || 0) + "px");
  };

  window.onSourceSelected = function (src) {
    applySource(src);
  };

  window.onPackProgress = function (msg) {
    els.progressLine.textContent = msg;
    els.progressLine.classList.remove("hidden");
  };

  window.onPackComplete = function (res) {
    state.packing = false;
    state.result = res;
    els.progressLine.classList.add("hidden");
    syncPackButton();

    els.resultFilename.textContent = res.suggestedFilename;
    els.resultFormatBadge.textContent = res.format.toUpperCase();
    els.metricFiles.textContent = String(res.packedFiles);
    els.metricSize.textContent = formatBytes(res.outputBytes);
    els.metricTokens.textContent = formatTokens(res.estimatedTokens);
    els.metricTime.textContent = res.durationMs >= 1000
      ? (res.durationMs / 1000).toFixed(2) + " s"
      : res.durationMs + " ms";

    const skipParts = [];
    if (res.skippedLargeFiles > 0) skipParts.push(`${res.skippedLargeFiles} oversized`);
    if (res.skippedBinaryOrIgnored > 0) skipParts.push(`${res.skippedBinaryOrIgnored} binary/ignored`);
    if (skipParts.length > 0) {
      els.skipSummary.textContent = "Filtered out: " + skipParts.join(" · ");
      els.skipSummary.classList.remove("hidden");
    } else {
      els.skipSummary.classList.add("hidden");
    }

    els.tabFileCount.textContent = String(res.files ? res.files.length : 0);
    els.treeOutput.textContent = res.treeText;
    els.previewOutput.textContent = res.previewText + (res.previewTruncated ? "\n\n... [Preview truncated — full document ready for Share / Save As] ..." : "");
    els.previewHint.textContent = res.previewTruncated
      ? `Showing first ${formatBytes(res.previewText.length)} of ${formatBytes(res.outputBytes)}`
      : `Full output (${formatBytes(res.outputBytes)})`;

    els.fileFilterInput.value = "";
    renderFileList("");

    els.resultPanel.classList.remove("hidden");
    els.resultPanel.scrollIntoView({ behavior: "smooth", block: "start" });
    showToast(`Packed ${res.packedFiles} files (${formatBytes(res.outputBytes)})`, "success");
  };

  window.onPackError = function (msg) {
    state.packing = false;
    els.progressLine.classList.add("hidden");
    syncPackButton();
    showToast(msg || "Pack failed", "error");
  };

  window.onNativeToast = function (msg, level) {
    showToast(msg, level);
  };

  // Event Bindings
  els.btnPickFolder.addEventListener("click", () => {
    if (window.F2mdNative) window.F2mdNative.pickFolder();
  });

  els.btnPickZip.addEventListener("click", () => {
    if (window.F2mdNative) window.F2mdNative.pickZip();
  });

  document.querySelectorAll(".seg-btn").forEach((btn) => {
    btn.addEventListener("click", () => {
      state.format = btn.dataset.format;
      savePrefs();
      syncControlsUI();
    });
  });

  document.querySelectorAll("[data-sort]").forEach((btn) => {
    btn.addEventListener("click", () => {
      const nextMode = btn.dataset.sort;
      if (state.sortMode === nextMode && nextMode !== "name") {
        state.sortAsc = !state.sortAsc;
      } else {
        state.sortMode = nextMode;
        state.sortAsc = false;
      }
      savePrefs();
      syncControlsUI();
    });
  });

  els.btnSortDir.addEventListener("click", () => {
    state.sortAsc = !state.sortAsc;
    savePrefs();
    syncControlsUI();
  });

  document.querySelectorAll(".size-chip").forEach((btn) => {
    btn.addEventListener("click", () => {
      state.maxSizeKb = Number(btn.dataset.size);
      savePrefs();
      syncControlsUI();
    });
  });

  els.inputMaxSize.addEventListener("input", () => {
    const val = parseInt(els.inputMaxSize.value, 10);
    state.maxSizeKb = Number.isFinite(val) && val >= 0 ? val : 0;
    document.querySelectorAll(".size-chip").forEach((btn) => {
      btn.classList.toggle("active", Number(btn.dataset.size) === state.maxSizeKb);
    });
    savePrefs();
  });

  els.chkFolderName.addEventListener("change", () => {
    state.useFolderName = els.chkFolderName.checked;
    savePrefs();
    syncControlsUI();
  });

  els.chkDate.addEventListener("change", () => {
    state.includeDate = els.chkDate.checked;
    savePrefs();
    syncControlsUI();
  });

  els.chkTime.addEventListener("change", () => {
    state.includeTime = els.chkTime.checked;
    savePrefs();
    syncControlsUI();
  });

  els.chkSkipLocks.addEventListener("change", () => {
    state.skipLocks = els.chkSkipLocks.checked;
    savePrefs();
  });

  els.btnPack.addEventListener("click", () => {
    if (!state.source || state.packing) return;
    state.packing = true;
    syncPackButton();
    els.progressLine.textContent = "Starting...";
    els.progressLine.classList.remove("hidden");

    const options = {
      format: state.format,
      sortMode: state.sortMode,
      sortAsc: state.sortAsc,
      useFolderName: state.useFolderName,
      pinReadme: state.useFolderName,
      includeDate: state.includeDate,
      includeTime: state.includeTime,
      maxSizeKb: state.maxSizeKb,
      skipLocks: state.skipLocks
    };

    if (window.F2mdNative) {
      window.F2mdNative.pack(JSON.stringify(options));
    }
  });

  els.btnShare.addEventListener("click", () => {
    if (window.F2mdNative) window.F2mdNative.shareOutput();
  });

  els.btnSave.addEventListener("click", () => {
    if (window.F2mdNative) window.F2mdNative.saveOutput();
  });

  els.btnCopy.addEventListener("click", () => {
    if (window.F2mdNative) window.F2mdNative.copyOutput();
  });

  els.btnCopyTree.addEventListener("click", () => {
    if (window.F2mdNative) window.F2mdNative.copyTree();
  });

  document.querySelectorAll(".tab-btn").forEach((btn) => {
    btn.addEventListener("click", () => {
      const target = btn.dataset.tab;
      document.querySelectorAll(".tab-btn").forEach((b) => {
        const active = b.dataset.tab === target;
        b.classList.toggle("active", active);
        b.setAttribute("aria-selected", String(active));
      });
      document.getElementById("paneTree").classList.toggle("active", target === "tree");
      document.getElementById("paneFiles").classList.toggle("active", target === "files");
      document.getElementById("panePreview").classList.toggle("active", target === "preview");
    });
  });

  els.fileFilterInput.addEventListener("input", () => {
    renderFileList(els.fileFilterInput.value);
  });

  // Bootstrap
  loadPrefs();
  syncControlsUI();

  if (window.F2mdNative) {
    try {
      const info = JSON.parse(window.F2mdNative.getAppInfo());
      if (info.version) els.versionPill.textContent = "v" + info.version;
      window.onSafeAreaInsets(info.safeTop || 0, info.safeBottom || 0);
      if (info.source) applySource(info.source);
    } catch (_) {}
  }
})();
