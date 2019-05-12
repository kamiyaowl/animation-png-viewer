package apng

import java.io.File

class PngData {

    companion object {
        val BUF_SIZE = 32
        val SIGNATURE = byteArrayOf(0x89.toByte(), 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a)
        // Result, ErrorMessage
        fun read(path: String): PngData {
            val data = PngData()
            val stream = File(path).inputStream()
            val buf = ByteArray(this.BUF_SIZE)
            stream.read(buf, 0, SIGNATURE.size)
            if (!buf.take(SIGNATURE.size).toByteArray().contentEquals(SIGNATURE)) {
                throw IllegalArgumentException("PNGファイルではありません")
            }
            stream.close()
            return data
        }
    }
}