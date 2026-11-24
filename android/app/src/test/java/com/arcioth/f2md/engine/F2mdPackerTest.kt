package com.arcioth.f2md.engine

import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Assert.assertTrue
import org.junit.Rule
import org.junit.Test
import org.junit.rules.TemporaryFolder
import java.io.File
import java.io.FileOutputStream
import java.util.zip.ZipEntry
import java.util.zip.ZipOutputStream

class F2mdPackerTest {

    @get:Rule
    val tempFolder = TemporaryFolder()

    private fun createZip(name: String, entries: List<Triple<String, ByteArray, Long>>): File {
        val zipFile = File(tempFolder.root, name)
        ZipOutputStream(FileOutputStream(zipFile)).use { zos ->
            for ((path, body, modMs) in entries) {
                val entry = ZipEntry(path).apply {
                    if (modMs > 0) time = modMs
                }
                zos.putNextEntry(entry)
                if (body.isNotEmpty()) {
                    zos.write(body)
                }
                zos.closeEntry()
            }
        }
        return zipFile
    }

    @Test
    fun packsZipWithSingleWrapperAndPinsReadme() {
        val zip = createZip(
            "download.zip",
            listOf(
                Triple("repo-main/", ByteArray(0), 0L),
                Triple("repo-main/zzz.kt", "fun z() = 1\n".toByteArray(), 1_700_000_000_000L),
                Triple("repo-main/README.md", "# Repo\n".toByteArray(), 1_600_000_000_000L),
                Triple("repo-main/src/App.kt", "fun main() {}\n".toByteArray(), 1_710_000_000_000L),
                Triple("repo-main/.git/config", "secret\n".toByteArray(), 0L),
                Triple("repo-main/node_modules/a/index.js", "x\n".toByteArray(), 0L),
                Triple("repo-main/package-lock.json", "{}\n".toByteArray(), 0L),
                Triple("repo-main/icon.png", byteArrayOf(0x89.toByte(), 0x50, 0x4E, 0x47, 0x00), 0L),
                Triple("__MACOSX/repo-main/._README.md", byteArrayOf(0x00, 0x05), 0L)
            )
        )

        val outFile = File(tempFolder.root, "out.md")
        val res = F2mdPacker.packZipFile(
            zipFile = zip,
            archiveDisplayName = "download.zip",
            options = PackOptions(
                format = OutputFormat.MARKDOWN,
                useFolderName = true,
                pinReadme = true,
                skipLocks = true
            ),
            outputFile = outFile
        )

        assertEquals("repo-main", res.rootName)
        assertEquals("repo-main_context.md", res.suggestedFilename)
        assertEquals(3, res.packedFiles)

        val out = outFile.readText()
        assertTrue(out.contains("repo-main/\n"))
        assertFalse(out.contains("__MACOSX"))
        assertFalse(out.contains(".git"))
        assertFalse(out.contains("node_modules"))
        assertFalse(out.contains("package-lock.json"))
        assertFalse(out.contains("## File: `icon.png`"))

        val readmeAt = out.indexOf("## File: `README.md`")
        val appAt = out.indexOf("## File: `src/App.kt`")
        val zzzAt = out.indexOf("## File: `zzz.kt`")
        assertTrue(readmeAt in 0 until appAt && readmeAt < zzzAt)
    }

    @Test
    fun packsFlatZipPlainTextSortedByModifiedAndSkipsOversized() {
        val oldMs = 1_640_995_200_000L
        val newMs = 1_767_225_600_000L
        val zip = createZip(
            "flatproj.zip",
            listOf(
                Triple("old.go", "package old\n".toByteArray(), oldMs),
                Triple("new.go", "package new\n".toByteArray(), newMs),
                Triple("huge.txt", "a".repeat(3000).toByteArray(), oldMs)
            )
        )

        val outFile = File(tempFolder.root, "out.txt")
        val res = F2mdPacker.packZipFile(
            zipFile = zip,
            archiveDisplayName = "flatproj.zip",
            options = PackOptions(
                format = OutputFormat.TEXT,
                sort = SortSpec(SortMode.MODIFIED, ascending = false),
                pinReadme = false,
                maxSizeKb = 1
            ),
            outputFile = outFile
        )

        assertEquals("flatproj", res.rootName)
        assertEquals(2, res.packedFiles)
        assertEquals(1, res.skippedLargeFiles)

        val out = outFile.readText()
        assertFalse(out.contains("```"))
        assertTrue(out.indexOf("----- new.go -----") < out.indexOf("----- old.go -----"))
        assertTrue(out.contains("[skipped: file size (2 KB) exceeds limit (1 KB)]"))
    }

    @Test
    fun dynamicFenceAndBinarySniffingWorkOnDirectory() {
        val dir = tempFolder.newFolder("sample")
        File(dir, "doc.md").writeText("```go\nfmt.Println()\n```\n")
        File(dir, "raw.custom").writeBytes(byteArrayOf(0x41, 0x42, 0x00, 0x43))
        File(dir, "plain.custom").writeText("hello utf8 world\n")

        val outFile = File(tempFolder.root, "sample_out.md")
        val res = F2mdPacker.packDirectory(
            rootDir = dir,
            options = PackOptions(format = OutputFormat.MARKDOWN),
            outputFile = outFile
        )

        assertEquals(2, res.packedFiles)
        val out = outFile.readText()
        assertTrue(out.contains("````markdown\n"))
        assertTrue(out.contains("raw.custom")) // Present in directory tree
        assertFalse(out.contains("## File: `raw.custom`")) // Binary content skipped
        assertTrue(out.contains("## File: `plain.custom`"))
    }
}
