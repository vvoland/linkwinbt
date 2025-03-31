group "default" {
    targets = ["binary"]
}

group "validate" {
    targets = ["test", "lint", "vet"]
}

target "binary" {
    dockerfile = "Dockerfile"
    context = "."
    target = "binary"
    output = ["type=local,dest=build"]
    platforms = ["local"]
}

target "test" {
    dockerfile = "dev.Dockerfile"
    context = "."
    target = "test"
    output = ["type=cacheonly"]
}

target "lint" {
    dockerfile = "dev.Dockerfile"
    context = "."
    target = "lint"
    output = ["type=cacheonly"]
}

target "vet" {
    dockerfile = "dev.Dockerfile"
    context = "."
    target = "vet"
    output = ["type=cacheonly"]
}
