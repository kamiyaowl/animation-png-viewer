package apng

fun main(args: Array<String>) {
    // test dir set
    val path: String = if (args.isEmpty()) "/Users/user/Documents/apng-impl/png_sample/sample.png" else args[0]
    val reader = PngReader()
    reader.read(path)
}