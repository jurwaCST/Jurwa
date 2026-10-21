# How we work

## Branches

Never commit straight to `main`. Make one branch per Jira card, named with the card key:

```
SCRUM-73-service-template
SCRUM-17-report-submission
```

Start every branch from an up-to-date `main`:

```
git checkout main
git pull
git checkout -b SCRUM-17-report-submission
```

## Commits and pushing

- Commit small and often, and push at least once a day so teammates can see progress.
- Start commit messages with the Jira key: `SCRUM-17 add create-report endpoint`.

## Pull requests

- Open a pull request into `main` as soon as something is reviewable. Put the Jira key in the title.
- Another member reviews and approves it. Nobody merges their own pull request.
- Keep pull requests small. One card is one pull request where possible.
- Unit tests go in the same pull request as the code. Routing, SLA and status-flow logic is written test-first.
- After a merge, delete the branch and move the card to Done in Jira.

## Contracts (.proto files)

- Files in `proto/` are shared. Do not change a contract without telling its owner and getting a review from every service that uses it.
- Keep changes backward compatible where possible.

## Jira

- Move your own cards to In Progress and Done.
- Log your hours on each card. The weekly EVM report uses them.
- Anything not on the board waits for a team decision (scope is locked in the charter).

## Service layout (hexagonal)

Each service follows the same folder layout, which the shared Go template will provide:

```
services/<name>/
  domain/        business rules, no framework imports
  application/   use cases (inbound ports)
  ports/         interfaces the core depends on (repositories, publishers)
  adapters/
    in/grpc/     gRPC handlers
    out/         database, Redis and other adapters
```
