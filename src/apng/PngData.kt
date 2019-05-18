package apng

import java.io.File
import java.io.InputStream

val ByteArray.toBigEndian get() = {
    (this[0].toInt() shl 24) or (this[1].toInt() shl 16)or (this[2].toInt() shl 8) or this[3].toInt()
}

class PngData {
    companion object {
        private val SIGNATURE = byteArrayOf(0x89.toByte(), 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a)
        // Result, ErrorMessage
        fun read(path: String): PngData {
            val stream: InputStream = File(path).inputStream()
            fun readByte(byteLen: Int) : ByteArray {
                val buf = ByteArray(byteLen)
                stream.read(buf, 0, buf.size)
                return buf
            }
            fun readInt(len: Int) : IntArray {
                val list = mutableListOf<Int>()
                val buf = ByteArray(len * 4)
                stream.read(buf, 0, buf.size)
                for(i in 0 until len) {
                    val ptr = i * 4
                    val data = (buf[ptr + 0].toInt() shl 24) or (buf[ptr + 1].toInt() shl 16) or (buf[ptr + 2].toInt() shl 8) or (buf[ptr + 3].toInt() shl 0)
                    list.add(data)
                }
                return list.toIntArray()
            }
            val dst = PngData()
            val sig = readByte(SIGNATURE.size)
            if (!sig.contentEquals(SIGNATURE)) {
                throw IllegalArgumentException("PNGファイルではありません")
            }
            while(stream.available() > 0) {
                val (byteLen) = readInt(1)
                val chunkType = readByte(4).map { it.toChar() }.joinToString(separator = "")
                val chunkData = readInt(byteLen / 4)
                val (crc) = readInt(1)
            }
            // チャンクデータの読み出し
            stream.close()
            return dst
        }
    }
}