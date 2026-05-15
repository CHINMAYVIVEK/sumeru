# Tooling

Commands and generators used every day while developing Sumeru.

## Run the server

From **`sumeru/`**:

```bash
make run
# or
go run ./cmd/sumeru -- -c sumeru.conf
```

Common flags: **`-i`** install, **`-u`** update metadata from disk, **`-p`** / **`--http-port`**, **`-d`** database override, **`--stop-after-init`** (run install/update then exit). See **`sumeru/README.md`** for the full table.

## Refresh addon Go imports (`zimports.go`)

Whenever you add an addon package that must register models in **`init()`**, regenerate blank imports:

```bash
cd sumeru
make generate
# equivalent: go generate ./cmd/sumeru
```

This runs **`cmd/sumeru-import-gen`** using an INI’s **`addons_path`** (in **`sumeru/`**, `go generate ./cmd/sumeru` uses the tracked **`sumeru.conf.example`** by default). External workspaces (**`sumeru_custom_addons`**) use **`make generate`** there with their own **`-out`** / **`-package`** — see **`sumeru_custom_addons/README.md`**.

## Scaffold a new addon (`sumeru-bp`)

From the **`sumeru`** repo root:

```bash
go run ./cmd/sumeru-bp -- -bp my_module
```

Creates **`addons/my_module/`** with manifest, **`init.go`**, **`models/`**, starter **`views/`**, **`security/`** (including **`sys.access.csv`**), and static placeholders. Module name must satisfy **`^[a-z][a-z0-9_]*$`**.

Then **`make generate`** and **`-i my_module`** once.

## Tests

```bash
cd sumeru
go test ./...
```

Add focused tests next to packages you touch (`render`, `orm`, `module`, …).

## Formatting

```bash
gofmt -w path/to/file.go
```

CI and local hygiene should keep **`gofmt`** clean on all committed Go files.
