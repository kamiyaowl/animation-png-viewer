package apng

import java.io.File
import java.io.InputStream

val ByteArray.toIntArray get() = {
    val len = this.size / 4
    val list = mutableListOf<Int>()
    for(i in 0 until len) {
        val ptr = i * 4
        val data = (this[ptr + 0].toInt() shl 24) or (this[ptr + 1].toInt() shl 16) or (this[ptr + 2].toInt() shl 8) or (this[ptr + 3].toInt() shl 0)
        list.add(data)
    }
    list.toIntArray()
}
val Collection<Byte>.toIntArray get() = {
    this.toByteArray().toIntArray()
}

class PngData {
    companion object {
        private val SIGNATURE = byteArrayOf(0x89.toByte(), 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a)
        // Result, ErrorMessage
        fun read(path: String): PngData {
            val stream: InputStream = File(path).inputStream()
            fun readByte(len: Int) : ByteArray {
                val buf = ByteArray(len)
                stream.read(buf, 0, buf.size)
                return buf
            }
            val dst = PngData()
            val sig = readByte(SIGNATURE.size)
            if (!sig.contentEquals(SIGNATURE)) {
                throw IllegalArgumentException("PNGファイルではありません")
            }
            while(stream.available() > 0) {
                val (dataLen) = readByte(4).toIntArray()
                val readLen = 4 + dataLen + 1 // chunktype(1) + chunkdata + crc(1)
                val data = readByte(readLen)

                val chunkTypeStr = data.take(4).joinToString("") { "${it.toChar()}" }
                val chunkData = data.drop(4).take(dataLen).toIntArray()
                val crc = data.last() // chunkType + chunkDataで計算
            }
            // チャンクデータの読み出し
            stream.close()
            return dst
        }
    }
}