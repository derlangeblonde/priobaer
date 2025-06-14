# Running Project Locally (for Development) 

The project uses z3 for solving "optimal" assignments. You have to build z3 first. To do that navigate to `internal/z3` and execute:
```sh
make```

After that you can define the required environment variables (for example via `source .dev-linux.env`) and finally run:
```sh
go run ./cmd/server```

