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

# Design & Concepts
This section documents some of the projets key concepts.

## Solving Algorithm
For finding an "optimal" assignments we employ the SMT-Solver Z3.

We can define the requirements for assignments as artithmetic or boolean constraints (e.g. a participant should be in exactly one course, courses should respect their minimum & maximium capacity). We can also define an objective functoin that should be maximized. For example we could score each created assignment based on the respective priority and then formulate a function that sums all scores obtained by created assignments. Z3 can find a solution for our problem instance so that all constraints are met and the objective function has the maximum possible result.

In theory this problem in NP-hard, but in practice Z3 will be able to calculate solutions for medium sized instances (N ~= 1000) in acceptable time.

## Per session database
This project was designed for the possibility of me actually operating this application. This means that it has to be GDPR-compliant. Also, since this is a side-project and not my day job, i wanted to keep maintenance effort and responsiblity for customer data to a minimum.
Therefore I came up with the following concept:

A user visiting the webpage is assigned as sesssion token (if they do not have one alre 


# Current State of the Project

Users can already:
- create, view, and delete participants, courses as well as their priorities.
- assign participants to courses via drag & drop.
- users can let the server solve for assignments of particpants to courses (currently respects course capacities, but not participants priorities).
- data can be imported & exported to & from excel files.

Important TODOs
- respect priorities in the solving algorithm.
- improve ui/ux for web-component that let's a user managa priorities - it currently is in a barely usable proof-of-concept state.
