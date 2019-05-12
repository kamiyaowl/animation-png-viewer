package apng

import java.io.File

class PngReader() {
    val BUF_SIZE = 32

    fun read(path: String) {
        val stream = File(path).inputStream()
        val buf = ByteArray(this.BUF_SIZE)

        stream.close()
    }

}