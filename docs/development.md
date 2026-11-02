# Development

The commands below need [Task](https://taskfile.dev); `task` alone lists all the tasks.

```bash
task check              # gofmt, go vet (also for macOS), golangci-lint, go test: the check before a commit
task test:update        # rewrite the golden files, then run the tests again
task build              # CGO_ENABLED=0 go build -o lacre .
task release:snapshot   # build the release archives into dist/, published nowhere
```

`task check` also needs golangci-lint, and `task release:snapshot` needs GoReleaser. The golden files in
`internal/app/testdata` fix the expected output. The binaries of `task release:snapshot` print a snapshot version, such
as `lacre v0.0.0-SNAPSHOT-<commit>` before the first tag.

On every push to a branch and on every pull request, continuous integration (CI) runs the checks of `task check` and the
build on Linux, and the tests and the build on Windows. No job runs on macOS: the macOS build is cross-compiled and
vetted, never run. A `v*` tag runs the same checks, then publishes a release with GoReleaser.
