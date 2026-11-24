package com.arcioth.f2md.engine

import java.io.File

enum class OutputFormat(val ext: String, val mimeType: String) {
    MARKDOWN("md", "text/markdown"),
    TEXT("txt", "text/plain");

    companion object {
        fun parse(raw: String?): OutputFormat = when (raw?.trim()?.lowercase()) {
            "txt", "text", "plain", "plaintext" -> TEXT
            else -> MARKDOWN
        }
    }
}

enum class SortMode {
    NAME,
    SIZE,
    CREATED,
    MODIFIED;

    companion object {
        fun parse(raw: String?): SortMode = when (raw?.trim()?.lowercase()) {
            "size", "s" -> SIZE
            "created", "date-created", "dc", "birth" -> CREATED
            "modified", "date-modified", "dm", "mtime", "m" -> MODIFIED
            else -> NAME
        }
    }
}

data class SortSpec(
    val mode: SortMode = SortMode.NAME,
    val ascending: Boolean = false
) {
    val isActive: Boolean get() = mode != SortMode.NAME
}

data class PackOptions(
    val format: OutputFormat = OutputFormat.MARKDOWN,
    val sort: SortSpec = SortSpec(),
    val useFolderName: Boolean = true,
    val pinReadme: Boolean = true,
    val includeDate: Boolean = false,
    val includeTime: Boolean = false,
    val maxSizeKb: Long = 500,
    val skipLocks: Boolean = true
)

data class PackedFileSummary(
    val relPath: String,
    val lang: String,
    val sizeBytes: Long,
    val skipped: Boolean,
    val reason: String? = null
)

data class PackResult(
    val rootName: String,
    val suggestedFilename: String,
    val outputFile: File,
    val format: OutputFormat,
    val sourceType: String,
    val treeText: String,
    val previewText: String,
    val previewTruncated: Boolean,
    val packedFiles: Int,
    val skippedLargeFiles: Int,
    val skippedBinaryOrIgnored: Int,
    val outputBytes: Long,
    val estimatedTokens: Long,
    val durationMs: Long,
    val files: List<PackedFileSummary>
)
