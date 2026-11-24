package com.arcioth.f2md.engine

import android.content.ContentResolver
import android.net.Uri
import android.provider.DocumentsContract
import java.io.File

object SafFolderPacker {

    private val CHILD_PROJECTION = arrayOf(
        DocumentsContract.Document.COLUMN_DOCUMENT_ID,
        DocumentsContract.Document.COLUMN_DISPLAY_NAME,
        DocumentsContract.Document.COLUMN_MIME_TYPE,
        DocumentsContract.Document.COLUMN_SIZE,
        DocumentsContract.Document.COLUMN_LAST_MODIFIED
    )

    fun resolveTreeDisplayName(contentResolver: ContentResolver, treeUri: Uri): String {
        val rootDocId = DocumentsContract.getTreeDocumentId(treeUri)
        val rootDocUri = DocumentsContract.buildDocumentUriUsingTree(treeUri, rootDocId)
        try {
            contentResolver.query(
                rootDocUri,
                arrayOf(DocumentsContract.Document.COLUMN_DISPLAY_NAME),
                null,
                null,
                null
            )?.use { cursor ->
                if (cursor.moveToFirst()) {
                    val idx = cursor.getColumnIndex(DocumentsContract.Document.COLUMN_DISPLAY_NAME)
                    if (idx >= 0) {
                        val name = cursor.getString(idx)
                        if (!name.isNullOrBlank()) return name
                    }
                }
            }
        } catch (_: Exception) {
            // Fall through to docId fallback
        }
        return rootDocId
            .substringAfterLast('/')
            .substringAfterLast(':')
            .ifBlank { "project" }
    }

    fun packTreeUri(
        contentResolver: ContentResolver,
        treeUri: Uri,
        options: PackOptions,
        outputFile: File,
        onProgress: (String) -> Unit = {}
    ): PackResult {
        val startMs = System.currentTimeMillis()
        val rootDocId = DocumentsContract.getTreeDocumentId(treeUri)
        val rootName = resolveTreeDisplayName(contentResolver, treeUri)

        val treeRoot = F2mdPacker.TreeNode(rootName, isDir = true)
        val textFiles = mutableListOf<F2mdPacker.VirtualFileEntry>()
        var skippedBinaryOrIgnored = 0
        var scannedCount = 0

        fun walk(parentDocId: String, relPrefix: String) {
            val childrenUri = DocumentsContract.buildChildDocumentsUriUsingTree(treeUri, parentDocId)
            val dirsToVisit = mutableListOf<Pair<String, String>>()

            contentResolver.query(childrenUri, CHILD_PROJECTION, null, null, null)?.use { cursor ->
                val idIdx = cursor.getColumnIndex(DocumentsContract.Document.COLUMN_DOCUMENT_ID)
                val nameIdx = cursor.getColumnIndex(DocumentsContract.Document.COLUMN_DISPLAY_NAME)
                val mimeIdx = cursor.getColumnIndex(DocumentsContract.Document.COLUMN_MIME_TYPE)
                val sizeIdx = cursor.getColumnIndex(DocumentsContract.Document.COLUMN_SIZE)
                val modIdx = cursor.getColumnIndex(DocumentsContract.Document.COLUMN_LAST_MODIFIED)

                while (cursor.moveToNext()) {
                    val docId = if (idIdx >= 0) cursor.getString(idIdx) else continue
                    val name = if (nameIdx >= 0) cursor.getString(nameIdx) else continue
                    if (docId.isNullOrEmpty() || name.isNullOrEmpty()) continue

                    val mime = if (mimeIdx >= 0) cursor.getString(mimeIdx) else ""
                    val isDir = mime == DocumentsContract.Document.MIME_TYPE_DIR
                    val size = if (sizeIdx >= 0 && !cursor.isNull(sizeIdx)) cursor.getLong(sizeIdx) else 0L
                    val modMs = if (modIdx >= 0 && !cursor.isNull(modIdx)) cursor.getLong(modIdx) else 0L

                    if (F2mdPacker.shouldIgnore(name, isDir, options.skipLocks)) {
                        skippedBinaryOrIgnored++
                        continue
                    }

                    val relPath = if (relPrefix.isEmpty()) name else "$relPrefix/$name"
                    treeRoot.insert(relPath, isDir, size, modMs, modMs)

                    if (isDir) {
                        dirsToVisit.add(docId to relPath)
                    } else {
                        scannedCount++
                        if (scannedCount % 25 == 0) {
                            onProgress("Scanning folder ($scannedCount files)...")
                        }

                        val docUri = DocumentsContract.buildDocumentUriUsingTree(treeUri, docId)
                        val (isText, lang) = F2mdPacker.detectText(name) {
                            contentResolver.openInputStream(docUri)
                        }
                        if (!isText) {
                            skippedBinaryOrIgnored++
                            continue
                        }

                        textFiles.add(
                            F2mdPacker.VirtualFileEntry(
                                relPath = relPath,
                                name = name,
                                lang = lang,
                                size = size,
                                createdMs = modMs,
                                modifiedMs = modMs,
                                openStream = { contentResolver.openInputStream(docUri) }
                            )
                        )
                    }
                }
            }

            for ((childDocId, childRelPath) in dirsToVisit) {
                walk(childDocId, childRelPath)
            }
        }

        onProgress("Scanning folder \"$rootName\"...")
        walk(rootDocId, "")

        return F2mdPacker.packVirtualEntries(
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
}
