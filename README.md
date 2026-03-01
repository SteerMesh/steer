# SteerMesh Steer CLI

Go CLI for the [SteerMesh](https://steermesh.dev) platform: compile tool-agnostic steering packs into tool-specific formats (Kiro, Cursor, Amazon Q, etc.), validate against the [spec](https://github.com/SteerMesh/spec), and sync with SteerMesh Cloud.

- **Website:** https://steermesh.dev  
- **Spec:** https://github.com/SteerMesh/spec  

## Commands

| Command | Description |
|---------|-------------|
| `steer init` | Initialize a SteerMesh project (steer.yaml, .steer/) |
| `steer compile` | Compile packs into target artifacts and bundle manifest |
| `steer validate` | Validate project config and pack YAML against spec |
| `steer add pack@version` | Add a pack and update lockfile |
| `steer sync` | Sync with SteerMesh Cloud (stub when API not ready) |
| `steer doctor` | Check env, config, lockfile, and bundle consistency |

## Build and test

```bash
make build    # Build steer binary
make test     # Run tests
make lint     # Run linters
```

## Exit codes

- **0** — Success
- **1** — Validation or user error (e.g. invalid config, missing lockfile, invalid pack YAML)
- **2** — Internal/runtime error (use `cli.ErrRuntime` when returning from commands)

## Multi-pack and lockfile

- **steer.yaml** lists packs (name + version constraint). Optional `registryUrl` or env `STEER_REGISTRY_URL`. Example: `version: "1.0.0"` or `"^1.0.0"`.
- **steer.lock** stores resolved version and source (`file://./packs/<name>`). When a registry is used, the pack is downloaded to `packs/<name>/pack.yaml` once so compile stays offline.
- **Compile** loads each pack from the lockfile source; if a pack is missing, it is resolved (registry or local) and the lockfile is updated. Semver (^, ~) is resolved to an exact version.
- Place pack content under your project’s `packs/` directory (e.g. clone [SteerMesh/packs](https://github.com/SteerMesh/packs) or copy a pack folder).

## Determinism

Builds are deterministic: no timestamps in bundle manifest or rendered output; stable iteration order. SHA256 checksums are emitted in the bundle manifest for every generated file.

## Roadmap / TODO

- [x] Implement init, compile, validate, add, sync (stub), doctor
- [x] Embed spec pack schema for offline validation
- [x] Multi-pack merge and lockfile-driven resolution in compile
- [ ] Cloud sync client (real API)
- [ ] Achieve >90% unit test coverage for internal packages (run `go test -cover ./...`)

## License

See the repository license file.
