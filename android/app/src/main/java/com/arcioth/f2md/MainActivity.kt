package com.arcioth.f2md

import android.annotation.SuppressLint
import android.content.ClipData
import android.content.ClipboardManager
import android.content.Context
import android.content.Intent
import android.graphics.Color
import android.net.Uri
import android.os.Build
import android.os.Bundle
import android.provider.OpenableColumns
import android.webkit.JavascriptInterface
import android.webkit.WebResourceRequest
import android.webkit.WebSettings
import android.webkit.WebView
import android.webkit.WebViewClient
import androidx.activity.result.contract.ActivityResultContracts
import androidx.appcompat.app.AppCompatActivity
import androidx.core.content.FileProvider
import androidx.core.view.ViewCompat
import androidx.core.view.WindowCompat
import androidx.core.view.WindowInsetsCompat
import com.arcioth.f2md.engine.F2mdPacker
import com.arcioth.f2md.engine.OutputFormat
import com.arcioth.f2md.engine.PackOptions
import com.arcioth.f2md.engine.PackResult
import com.arcioth.f2md.engine.SafFolderPacker
import com.arcioth.f2md.engine.SortMode
import com.arcioth.f2md.engine.SortSpec
import org.json.JSONArray
import org.json.JSONObject
import java.io.File
import java.io.FileOutputStream
import java.nio.charset.StandardCharsets
import java.util.concurrent.Executors

class MainActivity : AppCompatActivity() {

    private lateinit var webView: WebView
    private val worker = Executors.newSingleThreadExecutor()

    private sealed class SelectedSource {
        data class Folder(val treeUri: Uri, val displayName: String) : SelectedSource()
        data class ZipArchive(val uri: Uri?, val stagedFile: File?, val displayName: String, val sizeBytes: Long) : SelectedSource()
    }

    @Volatile
    private var selectedSource: SelectedSource? = null

    @Volatile
    private var lastResult: PackResult? = null

    @Volatile
    private var pageLoaded = false

    @Volatile
    private var pendingSafeTop = 0

    @Volatile
    private var pendingSafeBottom = 0

    private val pickFolderLauncher = registerForActivityResult(
        ActivityResultContracts.StartActivityForResult()
    ) { result ->
        val uri = result.data?.data ?: return@registerForActivityResult
        try {
            val flags = result.data?.flags?.and(
                Intent.FLAG_GRANT_READ_URI_PERMISSION
            ) ?: Intent.FLAG_GRANT_READ_URI_PERMISSION
            contentResolver.takePersistableUriPermission(uri, flags)
        } catch (_: SecurityException) {
            // Some providers do not support persistable permissions; transient read still works.
        }

        worker.execute {
            val name = SafFolderPacker.resolveTreeDisplayName(contentResolver, uri)
            selectedSource = SelectedSource.Folder(uri, name)
            notifySourceSelected("folder", name, -1L)
        }
    }

    private val pickZipLauncher = registerForActivityResult(
        ActivityResultContracts.StartActivityForResult()
    ) { result ->
        val uri = result.data?.data ?: return@registerForActivityResult
        handleZipUri(uri, stageImmediately = false)
    }

    private val saveDocumentLauncher = registerForActivityResult(
        ActivityResultContracts.StartActivityForResult()
    ) { result ->
        val destUri = result.data?.data ?: return@registerForActivityResult
        val current = lastResult ?: return@registerForActivityResult
        worker.execute {
            try {
                contentResolver.openOutputStream(destUri)?.use { out ->
                    current.outputFile.inputStream().use { input ->
                        input.copyTo(out, bufferSize = 32 * 1024)
                    }
                    out.flush()
                } ?: throw IllegalStateException("Could not open destination for writing")

                val savedName = queryDisplayName(destUri) ?: current.suggestedFilename
                notifyToast("Saved $savedName", "success")
            } catch (e: Exception) {
                notifyToast("Save failed: ${e.message}", "error")
            }
        }
    }

    @SuppressLint("SetJavaScriptEnabled")
    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)

        WindowCompat.setDecorFitsSystemWindows(window, false)
        window.statusBarColor = Color.TRANSPARENT
        window.navigationBarColor = Color.TRANSPARENT
        WindowCompat.getInsetsController(window, window.decorView).apply {
            isAppearanceLightStatusBars = false
            isAppearanceLightNavigationBars = false
        }

        webView = WebView(this).apply {
            setBackgroundColor(Color.parseColor("#0B0E11"))
            overScrollMode = WebView.OVER_SCROLL_NEVER
            settings.apply {
                javaScriptEnabled = true
                domStorageEnabled = true
                allowFileAccess = true
                allowContentAccess = false
                cacheMode = WebSettings.LOAD_NO_CACHE
                builtInZoomControls = false
                displayZoomControls = false
            }
            addJavascriptInterface(F2mdBridge(), "F2mdNative")
            webViewClient = object : WebViewClient() {
                override fun shouldOverrideUrlLoading(view: WebView?, request: WebResourceRequest?): Boolean {
                    val url = request?.url ?: return false
                    if (url.scheme == "https" || url.scheme == "http") {
                        startActivity(Intent(Intent.ACTION_VIEW, url))
                        return true
                    }
                    return false
                }

                override fun onPageFinished(view: WebView?, url: String?) {
                    pageLoaded = true
                    pushSafeInsetsToJs()
                    selectedSource?.let { src ->
                        when (src) {
                            is SelectedSource.Folder -> notifySourceSelected("folder", src.displayName, -1L)
                            is SelectedSource.ZipArchive -> notifySourceSelected("zip", src.displayName, src.sizeBytes)
                        }
                    }
                }
            }
        }

        setContentView(webView)

        ViewCompat.setOnApplyWindowInsetsListener(webView) { _, insets ->
            val bars = insets.getInsets(
                WindowInsetsCompat.Type.systemBars() or WindowInsetsCompat.Type.displayCutout()
            )
            val density = resources.displayMetrics.density
            pendingSafeTop = (bars.top / density).toInt()
            pendingSafeBottom = (bars.bottom / density).toInt()
            pushSafeInsetsToJs()
            insets
        }

        webView.loadUrl("file:///android_asset/index.html")
        handleIncomingIntent(intent)
    }

    override fun onNewIntent(intent: Intent) {
        super.onNewIntent(intent)
        setIntent(intent)
        handleIncomingIntent(intent)
    }

    override fun onDestroy() {
        super.onDestroy()
        worker.shutdownNow()
    }

    private fun handleIncomingIntent(intent: Intent?) {
        if (intent == null) return
        val uri: Uri? = when (intent.action) {
            Intent.ACTION_VIEW -> intent.data
            Intent.ACTION_SEND -> {
                if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.TIRAMISU) {
                    intent.getParcelableExtra(Intent.EXTRA_STREAM, Uri::class.java)
                } else {
                    @Suppress("DEPRECATION")
                    intent.getParcelableExtra(Intent.EXTRA_STREAM)
                }
            }
            else -> null
        }
        if (uri != null) {
            handleZipUri(uri, stageImmediately = true)
        }
    }

    private fun handleZipUri(uri: Uri, stageImmediately: Boolean) {
        worker.execute {
            try {
                val (displayName, sizeBytes) = queryFileMeta(uri)
                val finalName = displayName ?: "archive.zip"

                if (stageImmediately) {
                    val staged = stageUriToCache(uri, "incoming_archive.zip")
                    if (!isValidZipHeader(staged)) {
                        staged.delete()
                        notifyToast("Selected file is not a valid ZIP archive", "error")
                        return@execute
                    }
                    selectedSource = SelectedSource.ZipArchive(null, staged, finalName, staged.length())
                    notifySourceSelected("zip", finalName, staged.length())
                    notifyToast("Loaded $finalName", "success")
                } else {
                    // Quick PK header check before accepting
                    val validHeader = contentResolver.openInputStream(uri)?.use { stream ->
                        val magic = ByteArray(4)
                        val n = stream.read(magic)
                        n >= 2 && magic[0] == 0x50.toByte() && magic[1] == 0x4B.toByte()
                    } ?: false

                    if (!validHeader && !finalName.endsWith(".zip", ignoreCase = true)) {
                        notifyToast("Selected file is not a ZIP archive", "error")
                        return@execute
                    }

                    selectedSource = SelectedSource.ZipArchive(uri, null, finalName, sizeBytes)
                    notifySourceSelected("zip", finalName, sizeBytes)
                }
            } catch (e: Exception) {
                notifyToast("Could not open archive: ${e.message}", "error")
            }
        }
    }

    private fun stageUriToCache(uri: Uri, filename: String): File {
        val dir = File(cacheDir, "staging").apply { mkdirs() }
        val target = File(dir, filename)
        contentResolver.openInputStream(uri)?.use { input ->
            FileOutputStream(target).use { output ->
                input.copyTo(output, bufferSize = 64 * 1024)
            }
        } ?: throw IllegalStateException("Cannot read selected ZIP file")
        return target
    }

    private fun isValidZipHeader(file: File): Boolean {
        if (!file.exists() || file.length() < 4) return false
        return file.inputStream().use { s ->
            val b = ByteArray(2)
            s.read(b) == 2 && b[0] == 0x50.toByte() && b[1] == 0x4B.toByte()
        }
    }

    private fun queryDisplayName(uri: Uri): String? = queryFileMeta(uri).first

    private fun queryFileMeta(uri: Uri): Pair<String?, Long> {
        var name: String? = null
        var size = -1L
        try {
            contentResolver.query(
                uri,
                arrayOf(OpenableColumns.DISPLAY_NAME, OpenableColumns.SIZE),
                null,
                null,
                null
            )?.use { cursor ->
                if (cursor.moveToFirst()) {
                    val nameIdx = cursor.getColumnIndex(OpenableColumns.DISPLAY_NAME)
                    val sizeIdx = cursor.getColumnIndex(OpenableColumns.SIZE)
                    if (nameIdx >= 0 && !cursor.isNull(nameIdx)) {
                        name = cursor.getString(nameIdx)
                    }
                    if (sizeIdx >= 0 && !cursor.isNull(sizeIdx)) {
                        size = cursor.getLong(sizeIdx)
                    }
                }
            }
        } catch (_: Exception) {
            // Fallback to URI path segment
        }
        if (name.isNullOrBlank()) {
            name = uri.lastPathSegment?.substringAfterLast('/')
        }
        return name to size
    }

    private fun pushSafeInsetsToJs() {
        if (!pageLoaded) return
        evalJs("window.onSafeAreaInsets && window.onSafeAreaInsets($pendingSafeTop, $pendingSafeBottom);")
    }

    private fun notifySourceSelected(type: String, name: String, sizeBytes: Long) {
        val json = JSONObject().apply {
            put("type", type)
            put("name", name)
            put("sizeBytes", sizeBytes)
        }
        evalJs("window.onSourceSelected && window.onSourceSelected($json);")
    }

    private fun notifyProgress(message: String) {
        val quoted = JSONObject.quote(message)
        evalJs("window.onPackProgress && window.onPackProgress($quoted);")
    }

    private fun notifyToast(message: String, level: String = "info") {
        val qMsg = JSONObject.quote(message)
        val qLevel = JSONObject.quote(level)
        evalJs("window.onNativeToast && window.onNativeToast($qMsg, $qLevel);")
    }

    private fun evalJs(script: String) {
        runOnUiThread {
            if (!isFinishing && !isDestroyed) {
                webView.evaluateJavascript(script, null)
            }
        }
    }

    inner class F2mdBridge {

        @JavascriptInterface
        fun getAppInfo(): String {
            return JSONObject().apply {
                put("version", "0.4.0")
                put("safeTop", pendingSafeTop)
                put("safeBottom", pendingSafeBottom)
                selectedSource?.let { src ->
                    put("source", JSONObject().apply {
                        when (src) {
                            is SelectedSource.Folder -> {
                                put("type", "folder")
                                put("name", src.displayName)
                                put("sizeBytes", -1L)
                            }
                            is SelectedSource.ZipArchive -> {
                                put("type", "zip")
                                put("name", src.displayName)
                                put("sizeBytes", src.sizeBytes)
                            }
                        }
                    })
                }
            }.toString()
        }

        @JavascriptInterface
        fun pickFolder() {
            runOnUiThread {
                val intent = Intent(Intent.ACTION_OPEN_DOCUMENT_TREE).apply {
                    addFlags(
                        Intent.FLAG_GRANT_READ_URI_PERMISSION or
                            Intent.FLAG_GRANT_PERSISTABLE_URI_PERMISSION
                    )
                }
                pickFolderLauncher.launch(intent)
            }
        }

        @JavascriptInterface
        fun pickZip() {
            runOnUiThread {
                val intent = Intent(Intent.ACTION_OPEN_DOCUMENT).apply {
                    addCategory(Intent.CATEGORY_OPENABLE)
                    type = "*/*"
                    putExtra(
                        Intent.EXTRA_MIME_TYPES,
                        arrayOf(
                            "application/zip",
                            "application/x-zip-compressed",
                            "application/x-zip",
                            "application/octet-stream"
                        )
                    )
                    addFlags(Intent.FLAG_GRANT_READ_URI_PERMISSION)
                }
                pickZipLauncher.launch(intent)
            }
        }

        @JavascriptInterface
        fun pack(optionsJson: String) {
            val source = selectedSource
            if (source == null) {
                evalJs("window.onPackError && window.onPackError(${JSONObject.quote("Pick a folder or ZIP file first.")});")
                return
            }

            worker.execute {
                try {
                    val obj = JSONObject(optionsJson)
                    val format = OutputFormat.parse(obj.optString("format", "md"))
                    val sortMode = SortMode.parse(obj.optString("sortMode", "name"))
                    val sortAsc = obj.optBoolean("sortAsc", false)
                    val useFolderName = obj.optBoolean("useFolderName", true)
                    val pinReadme = obj.optBoolean("pinReadme", useFolderName)
                    val includeDate = obj.optBoolean("includeDate", false)
                    val includeTime = obj.optBoolean("includeTime", false)
                    val maxSizeKb = obj.optLong("maxSizeKb", 500L).coerceAtLeast(0L)
                    val skipLocks = obj.optBoolean("skipLocks", true)

                    val options = PackOptions(
                        format = format,
                        sort = SortSpec(sortMode, sortAsc),
                        useFolderName = useFolderName,
                        pinReadme = pinReadme,
                        includeDate = includeDate,
                        includeTime = includeTime,
                        maxSizeKb = maxSizeKb,
                        skipLocks = skipLocks
                    )

                    val exportDir = File(cacheDir, "exports").apply { mkdirs() }
                    val tempOut = File(exportDir, "latest_pack.${format.ext}")

                    val packResult = when (source) {
                        is SelectedSource.Folder -> {
                            SafFolderPacker.packTreeUri(
                                contentResolver = contentResolver,
                                treeUri = source.treeUri,
                                options = options,
                                outputFile = tempOut,
                                onProgress = ::notifyProgress
                            )
                        }
                        is SelectedSource.ZipArchive -> {
                            val zipFile = source.stagedFile ?: run {
                                notifyProgress("Reading ZIP archive...")
                                val uri = source.uri ?: throw IllegalStateException("ZIP source unavailable")
                                stageUriToCache(uri, "working_archive.zip")
                            }
                            if (!isValidZipHeader(zipFile)) {
                                throw IllegalArgumentException("File does not appear to be a valid ZIP archive.")
                            }
                            F2mdPacker.packZipFile(
                                zipFile = zipFile,
                                archiveDisplayName = source.displayName,
                                options = options,
                                outputFile = tempOut,
                                onProgress = ::notifyProgress
                            )
                        }
                    }

                    // Rename output file to the clean suggested filename inside exports/ for Share intent
                    val namedOut = File(exportDir, packResult.suggestedFilename)
                    if (namedOut.exists()) namedOut.delete()
                    val finalFile = if (tempOut.renameTo(namedOut)) namedOut else tempOut
                    val finalResult = packResult.copy(outputFile = finalFile)
                    lastResult = finalResult

                    val filesArr = JSONArray()
                    for (f in finalResult.files) {
                        filesArr.put(
                            JSONObject().apply {
                                put("relPath", f.relPath)
                                put("lang", f.lang)
                                put("sizeBytes", f.sizeBytes)
                                put("skipped", f.skipped)
                                if (f.reason != null) put("reason", f.reason)
                            }
                        )
                    }

                    val payload = JSONObject().apply {
                        put("rootName", finalResult.rootName)
                        put("suggestedFilename", finalResult.suggestedFilename)
                        put("format", finalResult.format.ext)
                        put("sourceType", finalResult.sourceType)
                        put("treeText", finalResult.treeText)
                        put("previewText", finalResult.previewText)
                        put("previewTruncated", finalResult.previewTruncated)
                        put("packedFiles", finalResult.packedFiles)
                        put("skippedLargeFiles", finalResult.skippedLargeFiles)
                        put("skippedBinaryOrIgnored", finalResult.skippedBinaryOrIgnored)
                        put("outputBytes", finalResult.outputBytes)
                        put("estimatedTokens", finalResult.estimatedTokens)
                        put("durationMs", finalResult.durationMs)
                        put("files", filesArr)
                    }

                    evalJs("window.onPackComplete && window.onPackComplete($payload);")
                } catch (e: Exception) {
                    val msg = e.message ?: e.javaClass.simpleName
                    evalJs("window.onPackError && window.onPackError(${JSONObject.quote(msg)});")
                }
            }
        }

        @JavascriptInterface
        fun saveOutput() {
            val current = lastResult
            if (current == null || !current.outputFile.exists()) {
                notifyToast("Nothing packed yet", "error")
                return
            }
            runOnUiThread {
                val intent = Intent(Intent.ACTION_CREATE_DOCUMENT).apply {
                    addCategory(Intent.CATEGORY_OPENABLE)
                    type = current.format.mimeType
                    putExtra(Intent.EXTRA_TITLE, current.suggestedFilename)
                }
                saveDocumentLauncher.launch(intent)
            }
        }

        @JavascriptInterface
        fun shareOutput() {
            val current = lastResult
            if (current == null || !current.outputFile.exists()) {
                notifyToast("Nothing packed yet", "error")
                return
            }
            runOnUiThread {
                try {
                    val contentUri = FileProvider.getUriForFile(
                        this@MainActivity,
                        "${applicationContext.packageName}.fileprovider",
                        current.outputFile
                    )
                    val sendIntent = Intent(Intent.ACTION_SEND).apply {
                        type = current.format.mimeType
                        putExtra(Intent.EXTRA_STREAM, contentUri)
                        putExtra(Intent.EXTRA_SUBJECT, current.suggestedFilename)
                        addFlags(Intent.FLAG_GRANT_READ_URI_PERMISSION)
                    }
                    startActivity(Intent.createChooser(sendIntent, "Share ${current.suggestedFilename}"))
                } catch (e: Exception) {
                    notifyToast("Share failed: ${e.message}", "error")
                }
            }
        }

        @JavascriptInterface
        fun copyOutput() {
            val current = lastResult
            if (current == null || !current.outputFile.exists()) {
                notifyToast("Nothing packed yet", "error")
                return
            }
            // Android Binder IPC limit is ~1 MB; cap safe clipboard copy around 750 KB
            val maxClipboardBytes = 750 * 1024L
            if (current.outputBytes > maxClipboardBytes) {
                notifyToast(
                    "Output is ${current.outputBytes / 1024} KB (too large for Android clipboard). Use Share or Save.",
                    "warn"
                )
                return
            }
            worker.execute {
                try {
                    val text = current.outputFile.readText(StandardCharsets.UTF_8)
                    runOnUiThread {
                        val cm = getSystemService(Context.CLIPBOARD_SERVICE) as ClipboardManager
                        cm.setPrimaryClip(ClipData.newPlainText(current.suggestedFilename, text))
                    }
                    notifyToast("Copied ${current.packedFiles} files to clipboard", "success")
                } catch (e: Exception) {
                    notifyToast("Copy failed: ${e.message}", "error")
                }
            }
        }

        @JavascriptInterface
        fun copyTree() {
            val current = lastResult
            if (current == null) {
                notifyToast("Nothing packed yet", "error")
                return
            }
            runOnUiThread {
                val cm = getSystemService(Context.CLIPBOARD_SERVICE) as ClipboardManager
                cm.setPrimaryClip(ClipData.newPlainText("${current.rootName} tree", current.treeText))
            }
            notifyToast("Directory tree copied", "success")
        }
    }
}
