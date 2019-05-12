package apng

fun main(args: Array<String>) {
    // test dir set
    val path: String = if (args.isEmpty()) "/Users/user/Documents/apng-impl/png_sample/sample.png" else args[0]
    val result: Result<PngData> = runCatching {
        PngData.read(path)
    }
    result.onSuccess {
        val data: PngData = result.getOrThrow()
        println("Success")
        println(data)
        // TODO: なんか表示する
    }.onFailure(Throwable::printStackTrace)

}