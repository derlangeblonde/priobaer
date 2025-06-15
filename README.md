# Prio Solver
*A web-based application to calculate "optimal" assignments of participants to courses based on the participants priorities.*

This is a learning project for me to 
- get better at building stuff.
- play around with different technologies and concepts.
- make my own mistakes and learn from them.

# Running Project Locally (for Development) 

The project uses z3 for solving "optimal" assignments. You have to build z3 first. To do that navigate to `internal/z3` and execute:
```sh
make
```

After that you can define the required environment variables (for example via `source .dev-linux.env`) and finally run:
```sh
go run ./cmd/server
```

*Hint: the `.dev-linux.env` defines a directory for sqlite db-files ad `./db`. Make sure that directory exists if you use the `.env` file*

