# Official Orax Client

Source code of the official client for the [Orax PegNet mining pool](https://www.oraxpool.com).

## Build

Cross compile the complete prod distribution:

```bash
# Build outputs located in build/prod folder
make clean && make prod             # Build with both Go 1.13 and 1.12
make clean && make prod-go-latest   # Build with Go >=1.13
make clean && make prod-go-1.12     # Build with Go 1.12
```

Or build a specific target:

Targets compiled with Go >=1.13:

- `orax-cli-linux-amd64`
- `orax-cli-linux-arm64`
- `orax-cli-linux-arm7`
- `orax-cli-darwin-amd64`
- `orax-cli-windows-amd64.exe`
- `orax-cli-windows-386.exe`

Targets compiled with Go 1.12:

- `orax-cli-linux-amd64-go-1.12`
- `orax-cli-linux-arm64-go-1.12`
- `orax-cli-linux-arm7-go-1.12`
- `orax-cli-darwin-amd64-go-1.12`
- `orax-cli-windows-amd64.exe-go-1.12`
- `orax-cli-windows-386.exe-go-1.12`
