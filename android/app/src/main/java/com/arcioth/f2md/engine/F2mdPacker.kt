package com.arcioth.f2md.engine

import java.io.BufferedOutputStream
import java.io.ByteArrayOutputStream
import java.io.File
import java.io.FileOutputStream
import java.io.InputStream
import java.nio.ByteBuffer
import java.nio.charset.CharacterCodingException
import java.nio.charset.CodingErrorAction
import java.nio.charset.StandardCharsets
import java.text.SimpleDateFormat
import java.util.Date
import java.util.Locale
import java.util.zip.ZipEntry
import java.util.zip.ZipFile

object F2mdPacker {

    const val PREVIEW_MAX_BYTES = 96 * 1024

    val defaultIgnoreDirs: Set<String> = setOf(
        ".git",
        "node_modules",
        "target",
        "dist",
        "build",
        ".next",
        "__pycache__",
        ".turbo",
        ".nuxt",
        "out",
        "bin",
        "obj",
        ".venv",
        "venv",
        "env",
        "vendor",
        ".idea",
        ".vscode",
        ".cache",
        ".gradle",
        "coverage",
        ".nyc_output",
        "tmp",
        "__MACOSX"
    )

    val commonTextFilenames: Map<String, String> = mapOf(
        "dockerfile" to "dockerfile",
        "containerfile" to "dockerfile",
        "makefile" to "makefile",
        "justfile" to "makefile",
        "gemfile" to "ruby",
        "procfile" to "yaml",
        "vagrantfile" to "ruby",
        "rakefile" to "ruby",
        "cmakelists.txt" to "cmake",
        "license" to "text",
        "licence" to "text",
        "copying" to "text",
        "readme" to "markdown",
        "caddyfile" to "caddyfile",
        "earthfile" to "earthfile",
        "authors" to "text",
        "contributors" to "text"
    )

    val lockFiles: Set<String> = setOf(
        "package-lock.json",
        "yarn.lock",
        "pnpm-lock.yaml",
        "cargo.lock",
        "composer.lock",
        "poetry.lock",
        "gemfile.lock",
        "mix.lock",
        "flake.lock",
        "bun.lockb"
    )

    val binaryExtensions: Set<String> = setOf(
        // Images
        ".png", ".jpg", ".jpeg", ".gif", ".webp", ".ico",
        ".bmp", ".tiff", ".tif", ".avif", ".heic", ".psd",
        // Audio / Video
        ".mp3", ".wav", ".ogg", ".flac", ".aac", ".m4a",
        ".mp4", ".mkv", ".webm", ".mov", ".avi", ".flv",
        // Archives & Packages
        ".zip", ".tar", ".gz", ".tgz", ".bz2", ".xz",
        ".7z", ".rar", ".iso", ".zst", ".deb", ".rpm",
        // Executables, Libraries & Bytecode
        ".exe", ".bin", ".dll", ".so", ".dylib", ".a",
        ".lib", ".o", ".obj", ".class", ".pyc", ".pyo",
        ".wasm", ".apk", ".aab", ".dex", ".jar",
        // Documents & Fonts
        ".pdf", ".docx", ".xlsx", ".pptx", ".odt", ".ods",
        ".woff", ".woff2", ".ttf", ".otf", ".eot",
        // Databases
        ".db", ".sqlite", ".sqlite3", ".parquet"
    )

    val textExtensions: Map<String, String> = mapOf(
        // Custom rule formats
        ".acv" to "ini",
        ".ctr" to "ini",
        ".rule" to "ini",

        // Systems & Low-level
        ".c" to "c",
        ".h" to "c",
        ".cpp" to "cpp",
        ".cc" to "cpp",
        ".cxx" to "cpp",
        ".hpp" to "cpp",
        ".hxx" to "cpp",
        ".hh" to "cpp",
        ".inl" to "cpp",
        ".rs" to "rust",
        ".go" to "go",
        ".zig" to "zig",
        ".d" to "d",
        ".nim" to "nim",
        ".v" to "verilog",
        ".sv" to "verilog",
        ".svh" to "verilog",
        ".vhd" to "vhdl",
        ".vhdl" to "vhdl",
        ".s" to "assembly",
        ".asm" to "assembly",
        ".nasm" to "assembly",

        // Web & Scripting
        ".js" to "javascript",
        ".mjs" to "javascript",
        ".cjs" to "javascript",
        ".jsx" to "jsx",
        ".ts" to "typescript",
        ".mts" to "typescript",
        ".cts" to "typescript",
        ".tsx" to "tsx",
        ".vue" to "vue",
        ".svelte" to "svelte",
        ".astro" to "astro",
        ".html" to "html",
        ".htm" to "html",
        ".xhtml" to "html",
        ".css" to "css",
        ".scss" to "scss",
        ".sass" to "sass",
        ".less" to "less",
        ".py" to "python",
        ".pyi" to "python",
        ".pyw" to "python",
        ".rb" to "ruby",
        ".rake" to "ruby",
        ".gemspec" to "ruby",
        ".php" to "php",
        ".phtml" to "php",
        ".lua" to "lua",
        ".pl" to "perl",
        ".pm" to "perl",
        ".tcl" to "tcl",
        ".awk" to "awk",
        ".sed" to "sed",
        ".sh" to "bash",
        ".bash" to "bash",
        ".zsh" to "zsh",
        ".fish" to "fish",
        ".ps1" to "powershell",
        ".psm1" to "powershell",
        ".bat" to "batch",
        ".cmd" to "batch",
        ".r" to "r",
        ".jl" to "julia",
        ".dart" to "dart",

        // Functional & BEAM
        ".ex" to "elixir",
        ".exs" to "elixir",
        ".erl" to "erlang",
        ".hrl" to "erlang",
        ".clj" to "clojure",
        ".cljs" to "clojure",
        ".cljc" to "clojure",
        ".edn" to "clojure",
        ".hs" to "haskell",
        ".lhs" to "haskell",
        ".ml" to "ocaml",
        ".mli" to "ocaml",
        ".lisp" to "lisp",
        ".lsp" to "lisp",
        ".scm" to "lisp",
        ".rkt" to "racket",

        // JVM & .NET
        ".java" to "java",
        ".kt" to "kotlin",
        ".kts" to "kotlin",
        ".scala" to "scala",
        ".sc" to "scala",
        ".groovy" to "groovy",
        ".gradle" to "groovy",
        ".cs" to "csharp",
        ".fs" to "fsharp",
        ".fsx" to "fsharp",
        ".vb" to "vb",
        ".swift" to "swift",

        // Config, Data & Schemas
        ".json" to "json",
        ".json5" to "json",
        ".jsonc" to "json",
        ".yaml" to "yaml",
        ".yml" to "yaml",
        ".toml" to "toml",
        ".xml" to "xml",
        ".xsl" to "xml",
        ".xslt" to "xml",
        ".svg" to "xml",
        ".sql" to "sql",
        ".graphql" to "graphql",
        ".gql" to "graphql",
        ".proto" to "protobuf",
        ".prisma" to "prisma",
        ".nix" to "nix",
        ".ini" to "ini",
        ".cfg" to "ini",
        ".conf" to "ini",
        ".properties" to "ini",
        ".desktop" to "ini",
        ".env.example" to "dotenv",
        ".env.template" to "dotenv",
        ".csv" to "csv",
        ".tsv" to "tsv",
        ".diff" to "diff",
        ".patch" to "diff",

        // Documentation & Markup
        ".md" to "markdown",
        ".markdown" to "markdown",
        ".mdown" to "markdown",
        ".mdx" to "markdown",
        ".txt" to "text",
        ".rst" to "rst",
        ".adoc" to "asciidoc",
        ".asciidoc" to "asciidoc",
        ".tex" to "latex",
        ".latex" to "latex",
        ".org" to "org"
    )

    private val readmeNames = setOf(
        "readme",
        "readme.md",
        "readme.markdown",
        "readme.mdown",
        "readme.mdx",
        "readme.rst",
        "readme.txt",
        "readme.org"
    )

    fun shouldIgnore(name: String, isDir: Boolean, skipLocks: Boolean): Boolean {
        if (isDir) {
            return name in defaultIgnoreDirs || name.startsWith(".")
        }
        if (name.startsWith(".")) {
            return true
        }
        val lower = name.lowercase(Locale.ROOT)
        if (lower == "f2md" || lower == "folder2md") {
            return true
        }
        if (lower == ".env" || lower.startsWith(".env.")) {
            return true
        }
        if (lower.endsWith(".map") || lower.endsWith(".min.js") || lower.endsWith(".min.css")) {
            return true
        }
        if (skipLocks && lower in lockFiles) {
            return true
        }
        return false
    }

    fun isReadmeLike(name: String): Boolean = name.lowercase(Locale.ROOT) in readmeNames

    fun pathDepth(relPath: String): Int {
        val norm = relPath.replace('\\', '/').trimEnd('/')
        if (norm.isEmpty() || norm == ".") return 0
        return norm.count { it == '/' }
    }

    fun fileExtensionLower(name: String): String {
        val dot = name.lastIndexOf('.')
        if (dot <= 0 || dot == name.length - 1) return ""
        return name.substring(dot).lowercase(Locale.ROOT)
    }

    fun detectText(name: String, openSniffStream: () -> InputStream?): Pair<Boolean, String> {
        val base = name.lowercase(Locale.ROOT)
        val ext = fileExtensionLower(name)
        if (ext in binaryExtensions) {
            return false to ""
        }
        commonTextFilenames[base]?.let { return true to it }
        textExtensions[ext]?.let { return true to it }

        val sniff = try {
            openSniffStream()?.use { stream ->
                val buf = ByteArray(512)
                var total = 0
                while (total < buf.size) {
                    val read = stream.read(buf, total, buf.size - total)
                    if (read <= 0) break
                    total += read
                }
                buf.copyOf(total)
            } ?: return false to ""
        } catch (_: Exception) {
            return false to ""
        }

        return sniffTextBytes(sniff, ext)
    }

    fun sniffTextBytes(buf: ByteArray, ext: String): Pair<Boolean, String> {
        if (buf.isEmpty()) return true to "text"
        for (b in buf) {
            if (b == 0.toByte()) return false to ""
        }

        // If we truncated mid-character at the 512-byte boundary, trim up to 3 trailing incomplete UTF-8 bytes
        var checkLen = buf.size
        if (buf.size == 512) {
            var tail = 0
            while (tail < 3 && checkLen > 0) {
                val lastByte = buf[checkLen - 1].toInt() and 0xFF
                if ((lastByte and 0x80) == 0) break // ASCII byte
                if ((lastByte and 0xC0) == 0xC0) {
                    // Start of a multi-byte sequence right at the end; trim it
                    checkLen--
                    break
                }
                // Continuation byte 10xxxxxx
                checkLen--
                tail++
            }
        }

        val decoder = StandardCharsets.UTF_8.newDecoder()
            .onMalformedInput(CodingErrorAction.REPORT)
            .onUnmappableCharacter(CodingErrorAction.REPORT)
        try {
            decoder.decode(ByteBuffer.wrap(buf, 0, checkLen))
        } catch (_: CharacterCodingException) {
            return false to ""
        }

        val trimmed = ext.removePrefix(".")
        return true to (trimmed.ifEmpty { "text" })
    }

    fun getCodeFence(content: ByteArray): String {
        var maxRun = 0
        var currentRun = 0
        val tick = '`'.code.toByte()
        for (b in content) {
            if (b == tick) {
                currentRun++
                if (currentRun > maxRun) maxRun = currentRun
            } else {
                currentRun = 0
            }
        }
        val len = if (maxRun >= 3) maxRun + 1 else 3
        return "`".repeat(len)
    }

    fun generateOutputFilename(
        rootDirName: String,
        useFolderName: Boolean,
        useDate: Boolean,
        useTime: Boolean,
        ext: String,
        now: Date = Date()
    ): String {
        val parts = mutableListOf<String>()
        val cleanRoot = sanitizeFilenameSegment(rootDirName)
        if (useFolderName && cleanRoot.isNotEmpty() && cleanRoot != "." && cleanRoot != "/") {
            parts.add(cleanRoot)
        } else {
            parts.add("project")
        }
        if (useDate) {
            parts.add(SimpleDateFormat("yyyy-MM-dd", Locale.US).format(now))
        }
        if (useTime) {
            parts.add(SimpleDateFormat("HH-mm-ss", Locale.US).format(now))
        }
        val finalExt = ext.ifEmpty { "md" }
        parts.add("context.$finalExt")
        return parts.joinToString("_")
    }

    private fun sanitizeFilenameSegment(raw: String): String {
        return raw.trim()
            .replace(Regex("[/\\\\:*?\"<>|]+"), "_")
            .trim('_')
    }

    data class TreeNode(
        val name: String,
        var isDir: Boolean,
        var size: Long = 0L,
        var createdMs: Long = 0L,
        var modifiedMs: Long = 0L,
        val children: MutableMap<String, TreeNode> = LinkedHashMap()
    ) {
        fun insert(relPath: String, isDirEntry: Boolean, size: Long, createdMs: Long, modifiedMs: Long) {
            val parts = relPath.split('/').filter { it.isNotEmpty() }
            var cur = this
            for (i in parts.indices) {
                val seg = parts[i]
                val last = i == parts.lastIndex
                if (!last || isDirEntry) {
                    val child = cur.children.getOrPut(seg) { TreeNode(seg, isDir = true) }
                    child.isDir = true
                    if (last && modifiedMs > 0) {
                        child.modifiedMs = modifiedMs
                        child.createdMs = if (createdMs > 0) createdMs else modifiedMs
                    }
                    cur = child
                } else {
                    cur.children[seg] = TreeNode(
                        name = seg,
                        isDir = false,
                        size = size,
                        createdMs = if (createdMs > 0) createdMs else modifiedMs,
                        modifiedMs = modifiedMs
                    )
                }
            }
        }
    }

    data class VirtualFileEntry(
        val relPath: String,
        val name: String,
        val lang: String,
        val size: Long,
        val createdMs: Long,
        val modifiedMs: Long,
        val openStream: () -> InputStream?
    )

    private fun compareFileMeta(
        aName: String,
        aRelPath: String,
        aSize: Long,
        aCreated: Long,
        aModified: Long,
        bName: String,
        bRelPath: String,
        bSize: Long,
        bCreated: Long,
        bModified: Long,
        spec: SortSpec,
        pinReadme: Boolean
    ): Int {
        if (pinReadme) {
            val ar = isReadmeLike(aName)
            val br = isReadmeLike(bName)
            if (ar != br) return if (ar) -1 else 1
            if (ar && br) {
                val da = pathDepth(aRelPath)
                val db = pathDepth(bRelPath)
                if (da != db) return da.compareTo(db)
            }
        }

        return when (spec.mode) {
            SortMode.CREATED -> {
                if (aCreated != bCreated) {
                    if (spec.ascending) aCreated.compareTo(bCreated) else bCreated.compareTo(aCreated)
                } else 0
            }
            SortMode.MODIFIED -> {
                if (aModified != bModified) {
                    if (spec.ascending) aModified.compareTo(bModified) else bModified.compareTo(aModified)
                } else 0
            }
            SortMode.SIZE -> {
                if (aSize != bSize) {
                    if (spec.ascending) aSize.compareTo(bSize) else bSize.compareTo(aSize)
                } else 0
            }
            SortMode.NAME -> 0
        }
    }

    fun renderTree(node: TreeNode, prefix: String, options: PackOptions, sb: StringBuilder) {
        if (node.children.isEmpty()) return
        val sorted = node.children.values.sortedWith { a, b ->
            if (a.isDir != b.isDir) {
                return@sortedWith if (a.isDir) -1 else 1
            }
            val pin = options.pinReadme && !a.isDir
            val cmp = compareFileMeta(
                a.name, a.name, a.size, a.createdMs, a.modifiedMs,
                b.name, b.name, b.size, b.createdMs, b.modifiedMs,
                options.sort, pin
            )
            if (cmp != 0) cmp else a.name.compareTo(b.name)
        }

        for (i in sorted.indices) {
            val child = sorted[i]
            val isLast = i == sorted.lastIndex
            val connector = if (isLast) "└── " else "├── "
            val extension = if (isLast) "    " else "│   "
            sb.append(prefix).append(connector).append(child.name).append('\n')
            if (child.isDir) {
                renderTree(child, prefix + extension, options, sb)
            }
        }
    }

    fun cleanZipEntryPath(raw: String): Pair<String, Boolean>? {
        val norm = raw.replace('\\', '/')
        val isDir = norm.endsWith("/")
        val trimmed = norm.trimStart('/')
        if (trimmed.isEmpty()) return null

        val parts = ArrayList<String>()
        for (seg in trimmed.split('/')) {
            if (seg.isEmpty() || seg == ".") continue
            if (seg == "..") return null
            parts.add(seg)
        }
        if (parts.isEmpty()) return null
        return parts.joinToString("/") to isDir
    }

    fun shouldIgnoreRelativePath(relPath: String, isDir: Boolean, skipLocks: Boolean): Boolean {
        val parts = relPath.split('/').filter { it.isNotEmpty() }
        if (parts.isEmpty()) return true
        val dirLimit = if (isDir) parts.size else parts.size - 1
        for (i in 0 until dirLimit) {
            if (shouldIgnore(parts[i], isDir = true, skipLocks = skipLocks)) return true
        }
        if (!isDir && shouldIgnore(parts.last(), isDir = false, skipLocks = skipLocks)) {
            return true
        }
        return false
    }

    fun detectZipWrapper(entries: List<Pair<String, Boolean>>, skipLocks: Boolean): String {
        var topDir: String? = null
        var hasNested = false
        for ((cleaned, isDir) in entries) {
            val parts = cleaned.split('/')
            val first = parts[0]
            val firstIsDir = isDir || parts.size > 1
            if (shouldIgnore(first, firstIsDir, skipLocks)) continue
            if (!firstIsDir) return ""
            if (topDir == null) {
                topDir = first
            } else if (topDir != first) {
                return ""
            }
            if (parts.size > 1) hasNested = true
        }
        return if (topDir != null && hasNested) topDir else ""
    }

    fun packZipFile(
        zipFile: File,
        archiveDisplayName: String,
        options: PackOptions,
        outputFile: File,
        onProgress: (String) -> Unit = {}
    ): PackResult {
        val startMs = System.currentTimeMillis()
        onProgress("Opening ZIP archive...")

        ZipFile(zipFile).use { zf ->
            val rawEntries = mutableListOf<Triple<ZipEntry, String, Boolean>>()
            val enumEntries = zf.entries()
            while (enumEntries.hasMoreElements()) {
                val ze = enumEntries.nextElement()
                val cleanedPair = cleanZipEntryPath(ze.name) ?: continue
                val (cleaned, endsSlash) = cleanedPair
                rawEntries.add(Triple(ze, cleaned, ze.isDirectory || endsSlash))
            }

            val wrapper = detectZipWrapper(
                rawEntries.map { it.second to it.third },
                options.skipLocks
            )
            val stripPrefix = if (wrapper.isNotEmpty()) "$wrapper/" else ""
            val fallbackStem = archiveDisplayName
                .substringAfterLast('/')
                .replace(Regex("\\.zip$", RegexOption.IGNORE_CASE), "")
                .ifEmpty { "archive" }
            val rootName = wrapper.ifEmpty { fallbackStem }

            onProgress("Indexing ${rawEntries.size} archive entries...")

            val treeRoot = TreeNode(rootName, isDir = true)
            val textFiles = mutableListOf<VirtualFileEntry>()
            var skippedBinaryOrIgnored = 0

            for ((ze, cleaned, isDir) in rawEntries) {
                val relPath = if (stripPrefix.isNotEmpty()) {
                    if (cleaned == wrapper) continue
                    if (!cleaned.startsWith(stripPrefix)) {
                        skippedBinaryOrIgnored++
                        continue
                    }
                    cleaned.removePrefix(stripPrefix)
                } else {
                    cleaned
                }
                if (relPath.isEmpty()) continue

                if (shouldIgnoreRelativePath(relPath, isDir, options.skipLocks)) {
                    skippedBinaryOrIgnored++
                    continue
                }

                val modMs = if (ze.time > 0) ze.time else 0L
                val size = if (ze.size >= 0) ze.size else 0L
                treeRoot.insert(relPath, isDir, size, modMs, modMs)

                if (isDir) continue

                val name = relPath.substringAfterLast('/')
                val (isText, lang) = detectText(name) { zf.getInputStream(ze) }
                if (!isText) {
                    skippedBinaryOrIgnored++
                    continue
                }

                textFiles.add(
                    VirtualFileEntry(
                        relPath = relPath,
                        name = name,
                        lang = lang,
                        size = size,
                        createdMs = modMs,
                        modifiedMs = modMs,
                        openStream = { zf.getInputStream(ze) }
                    )
                )
            }

            return packVirtualEntries(
                rootName = rootName,
                treeRoot = treeRoot,
                files = textFiles,
                skippedBinaryOrIgnored = skippedBinaryOrIgnored,
                options = options,
                outputFile = outputFile,
                sourceType = "zip",
                startMs = startMs,
                onProgress = onProgress
            )
        }
    }

    fun packDirectory(
        rootDir: File,
        options: PackOptions,
        outputFile: File,
        onProgress: (String) -> Unit = {}
    ): PackResult {
        val startMs = System.currentTimeMillis()
        val rootName = rootDir.name.ifEmpty { "project" }
        val treeRoot = TreeNode(rootName, isDir = true)
        val textFiles = mutableListOf<VirtualFileEntry>()
        var skippedBinaryOrIgnored = 0
        val absOutputPath = outputFile.canonicalPath

        fun walk(dir: File, relPrefix: String) {
            val children = dir.listFiles() ?: return
            for (child in children) {
                if (child.canonicalPath == absOutputPath) continue
                val name = child.name
                val isDir = child.isDirectory
                if (shouldIgnore(name, isDir, options.skipLocks)) {
                    skippedBinaryOrIgnored++
                    continue
                }
                val relPath = if (relPrefix.isEmpty()) name else "$relPrefix/$name"
                val modMs = child.lastModified()
                val size = if (isDir) 0L else child.length()
                treeRoot.insert(relPath, isDir, size, modMs, modMs)

                if (isDir) {
                    walk(child, relPath)
                } else {
                    val (isText, lang) = detectText(name) { child.inputStream() }
                    if (!isText) {
                        skippedBinaryOrIgnored++
                        continue
                    }
                    textFiles.add(
                        VirtualFileEntry(
                            relPath = relPath,
                            name = name,
                            lang = lang,
                            size = size,
                            createdMs = modMs,
                            modifiedMs = modMs,
                            openStream = { child.inputStream() }
                        )
                    )
                }
            }
        }

        onProgress("Scanning directory...")
        walk(rootDir, "")

        return packVirtualEntries(
            rootName = rootName,
            treeRoot = treeRoot,
            files = textFiles,
            skippedBinaryOrIgnored = skippedBinaryOrIgnored,
            options = options,
            outputFile = outputFile,
            sourceType = "folder",
            startMs = startMs,
            onProgress = onProgress
        )
    }

    fun packVirtualEntries(
        rootName: String,
        treeRoot: TreeNode,
        files: List<VirtualFileEntry>,
        skippedBinaryOrIgnored: Int,
        options: PackOptions,
        outputFile: File,
        sourceType: String,
        startMs: Long = System.currentTimeMillis(),
        onProgress: (String) -> Unit = {}
    ): PackResult {
        onProgress("Building directory tree...")
        val treeSb = StringBuilder()
        treeSb.append(rootName).append("/\n")
        renderTree(treeRoot, "", options, treeSb)
        val treeText = treeSb.toString()

        // Lexicographical walk order first, then stable sort for sortSpec / pinReadme
        val orderedFiles = files.sortedBy { it.relPath }.let { lex ->
            if (options.sort.isActive || options.pinReadme) {
                lex.sortedWith { a, b ->
                    compareFileMeta(
                        a.name, a.relPath, a.size, a.createdMs, a.modifiedMs,
                        b.name, b.relPath, b.size, b.createdMs, b.modifiedMs,
                        options.sort, options.pinReadme
                    )
                }
            } else {
                lex
            }
        }

        outputFile.parentFile?.mkdirs()
        val maxSizeBytes = options.maxSizeKb * 1024L

        val previewBuffer = ByteArrayOutputStream()
        var previewTruncated = false
        var totalOutputBytes = 0L
        var packedCount = 0
        var skippedLargeCount = 0
        val summaries = ArrayList<PackedFileSummary>(orderedFiles.size)

        BufferedOutputStream(FileOutputStream(outputFile)).use { bos ->
            fun emitBytes(bytes: ByteArray) {
                bos.write(bytes)
                totalOutputBytes += bytes.size
                if (!previewTruncated) {
                    val remaining = PREVIEW_MAX_BYTES - previewBuffer.size()
                    if (remaining >= bytes.size) {
                        previewBuffer.write(bytes)
                    } else {
                        if (remaining > 0) {
                            previewBuffer.write(bytes, 0, remaining)
                        }
                        previewTruncated = true
                    }
                }
            }

            fun emit(text: String) = emitBytes(text.toByteArray(StandardCharsets.UTF_8))

            if (options.format == OutputFormat.TEXT) {
                emit("Project Directory Structure\n\n")
                emit(treeText)
                emit("\nFile Contents\n\n")
            } else {
                emit("# Project Directory Structure\n\n```text\n")
                emit(treeText)
                emit("```\n\n---\n\n# File Contents\n\n")
            }

            val total = orderedFiles.size
            for ((index, entry) in orderedFiles.withIndex()) {
                if (index % 20 == 0 || index == total - 1) {
                    onProgress("Packing ${index + 1} / $total files...")
                }

                if (maxSizeBytes > 0 && entry.size > maxSizeBytes) {
                    emitSkipped(options.format, entry.relPath, entry.size, options.maxSizeKb, ::emit)
                    skippedLargeCount++
                    summaries.add(
                        PackedFileSummary(
                            relPath = entry.relPath,
                            lang = entry.lang,
                            sizeBytes = entry.size,
                            skipped = true,
                            reason = "Exceeds ${options.maxSizeKb} KB"
                        )
                    )
                    continue
                }

                val contentResult = readStreamCapped(entry.openStream, maxSizeBytes)
                when (contentResult) {
                    is ReadResult.Error -> {
                        emitReadError(options.format, entry.relPath, contentResult.message, ::emit)
                        summaries.add(
                            PackedFileSummary(
                                relPath = entry.relPath,
                                lang = entry.lang,
                                sizeBytes = entry.size,
                                skipped = true,
                                reason = contentResult.message
                            )
                        )
                    }
                    is ReadResult.TooLarge -> {
                        val reportedSize = maxOf(entry.size, contentResult.observedSize)
                        emitSkipped(options.format, entry.relPath, reportedSize, options.maxSizeKb, ::emit)
                        skippedLargeCount++
                        summaries.add(
                            PackedFileSummary(
                                relPath = entry.relPath,
                                lang = entry.lang,
                                sizeBytes = reportedSize,
                                skipped = true,
                                reason = "Exceeds ${options.maxSizeKb} KB"
                            )
                        )
                    }
                    is ReadResult.Ok -> {
                        val content = contentResult.bytes
                        emitFileBody(options.format, entry.relPath, entry.lang, content, ::emit, ::emitBytes)
                        packedCount++
                        summaries.add(
                            PackedFileSummary(
                                relPath = entry.relPath,
                                lang = entry.lang,
                                sizeBytes = content.size.toLong(),
                                skipped = false
                            )
                        )
                    }
                }
            }
            bos.flush()
        }

        val previewStr = decodeUtf8Safe(previewBuffer.toByteArray())
        val suggested = generateOutputFilename(
            rootDirName = rootName,
            useFolderName = options.useFolderName,
            useDate = options.includeDate,
            useTime = options.includeTime,
            ext = options.format.ext
        )

        val durationMs = (System.currentTimeMillis() - startMs).coerceAtLeast(1L)
        val estimatedTokens = (totalOutputBytes / 4L).coerceAtLeast(if (totalOutputBytes > 0) 1L else 0L)

        return PackResult(
            rootName = rootName,
            suggestedFilename = suggested,
            outputFile = outputFile,
            format = options.format,
            sourceType = sourceType,
            treeText = treeText,
            previewText = previewStr,
            previewTruncated = previewTruncated,
            packedFiles = packedCount,
            skippedLargeFiles = skippedLargeCount,
            skippedBinaryOrIgnored = skippedBinaryOrIgnored,
            outputBytes = totalOutputBytes,
            estimatedTokens = estimatedTokens,
            durationMs = durationMs,
            files = summaries
        )
    }

    private sealed class ReadResult {
        data class Ok(val bytes: ByteArray) : ReadResult()
        data class TooLarge(val observedSize: Long) : ReadResult()
        data class Error(val message: String) : ReadResult()
    }

    private fun readStreamCapped(openStream: () -> InputStream?, maxSizeBytes: Long): ReadResult {
        return try {
            val stream = openStream() ?: return ReadResult.Error("Unable to open stream")
            stream.use { input ->
                val out = ByteArrayOutputStream()
                val buf = ByteArray(16 * 1024)
                var total = 0L
                while (true) {
                    val read = input.read(buf)
                    if (read <= 0) break
                    total += read
                    if (maxSizeBytes > 0 && total > maxSizeBytes) {
                        return ReadResult.TooLarge(total)
                    }
                    out.write(buf, 0, read)
                }
                ReadResult.Ok(out.toByteArray())
            }
        } catch (e: Exception) {
            ReadResult.Error(e.message ?: "Read error")
        }
    }

    private fun emitSkipped(
        format: OutputFormat,
        relPath: String,
        sizeBytes: Long,
        maxKb: Long,
        emit: (String) -> Unit
    ) {
        if (format == OutputFormat.TEXT) {
            emit("----- $relPath -----\n")
            emit("[skipped: file size (${sizeBytes / 1024} KB) exceeds limit ($maxKb KB)]\n\n")
        } else {
            emit("## File: `$relPath`\n\n")
            emit("<!-- Skipped: file size (${sizeBytes / 1024} KB) exceeds limit ($maxKb KB) -->\n\n---\n\n")
        }
    }

    private fun emitReadError(
        format: OutputFormat,
        relPath: String,
        errorMsg: String,
        emit: (String) -> Unit
    ) {
        if (format == OutputFormat.TEXT) {
            emit("----- $relPath -----\n")
            emit("Error reading file: $errorMsg\n\n")
        } else {
            emit("## File: `$relPath`\n\n")
            emit("Error reading file: $errorMsg\n\n---\n\n")
        }
    }

    private fun emitFileBody(
        format: OutputFormat,
        relPath: String,
        lang: String,
        content: ByteArray,
        emit: (String) -> Unit,
        emitBytes: (ByteArray) -> Unit
    ) {
        val endsWithNewline = content.isNotEmpty() && content.last() == '\n'.code.toByte()
        if (format == OutputFormat.TEXT) {
            emit("----- $relPath -----\n")
            emitBytes(content)
            if (content.isNotEmpty() && !endsWithNewline) {
                emit("\n")
            }
            emit("\n")
            return
        }

        val fence = getCodeFence(content)
        emit("## File: `$relPath`\n\n")
        emit("$fence$lang\n")
        emitBytes(content)
        if (content.isNotEmpty() && !endsWithNewline) {
            emit("\n")
        }
        emit("$fence\n\n---\n\n")
    }

    private fun decodeUtf8Safe(bytes: ByteArray): String {
        return String(bytes, StandardCharsets.UTF_8)
    }
}
