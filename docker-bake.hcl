group "default" {
    targets = ["binary"]
}

target "binary" {
    dockerfile = "Dockerfile"
    context = "."
    target = "binary"
    output = ["type=local,dest=_build"]
    attest = [
        { type = "provenance", mode = "max" },
        { type = "sbom" }
    ]
    platforms = ["local"]
}
