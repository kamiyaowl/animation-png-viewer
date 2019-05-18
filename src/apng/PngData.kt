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
        val crcTablble: ByteArray? = null
        fun crc32(src: ByteArray) {
            // 事前テーブルの計算
            for(i in 0 until 256) {
                var c: UInt = i.toUInt()
                for(j in 0 until 8) {
                    c = if (c and 1u != 0u) { 0xedb88320 } else { 0x0 }
                }
            }
            //  実装についてはwiki参照

        }
    }
    private val fileHeadSignature = byteArrayOf(0x89.toByte(), 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a)


    // 指定されたファイルを読み込んで、ファイルに記載されたパラメータを内部変数にセットします
    fun read(path: String) {
        val stream: InputStream = File(path).inputStream()
        fun readByte(len: Int) : ByteArray {
            val buf = ByteArray(len)
            stream.read(buf, 0, buf.size)
            return buf
        }
        val sig = readByte(fileHeadSignature.size)
        if (!sig.contentEquals(fileHeadSignature)) {
            throw IllegalArgumentException("PNGファイルではありません")
        }
        while(stream.available() > 0) {
            val (dataLen) = readByte(4).toIntArray()
            val readLen = 4 + dataLen + 1 // chunktype(1) + chunkdata + crc(1)
            val data = readByte(readLen)

            val chunkTypeStr = data.take(4).joinToString("") { "${it.toChar()}" }
            val chunkData =
                if (dataLen > 0)  { data.drop(4).take(dataLen).toByteArray() }
                else { ByteArray(0) } // IENDの場合、データは空
            val crc = data.drop(4 + dataLen).take(4).toIntArray() // chunkType + chunkDataで計算
            // CRC検査
        }
        // チャンクデータの読み出し
        stream.close()
    }
}