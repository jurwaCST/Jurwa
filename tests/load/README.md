# Load test: report submission (JMeter)

Card: SCRUM-109. Owner: Sonam Zangmo. Requirement NFR-02 and quality plan Q6: report submission is load-tested with Apache JMeter through the api-gateway.

| File | What it is |
|---|---|
| `report-submission.jmx` | The JMeter plan: each virtual user sends `POST /reports` with its own Student token, in a loop, for a fixed time |
| `run.sh` | Signs up one Student per virtual user (not timed), runs JMeter without the GUI, prints the summary |
| `summarize.py` | Count, errors, average, median, p90, p95, p99, max, throughput and HTTP codes from the results file, and whether the target is met |
| `out/` | Results of each run (ignored by git): `results.jtl`, `summary.txt` and the JMeter HTML report in `html/` |

## Target

50 users at the same time for 2 minutes, each sending one report a second: 95% of `POST /reports` answered in under 500 ms, and under 1% errors. Proposed to the team on 7 November 2026 for NFR-02.

## Run it

Install JMeter once (`brew install jmeter`; it needs Java 17 or later). Then start the four services the route needs, each in its own terminal, from the repo root:

```bash
JWT_SECRET=<any long random string> make run SERVICE=identity
make run SERVICE=authorization
make run SERVICE=report-submission
RATE_LIMIT_PER_SECOND=1000 make run SERVICE=api-gateway
```

`RATE_LIMIT_PER_SECOND=1000` is only for the load run: all virtual users come from one laptop, so the normal limit of 20 a second per IP would turn most requests into 429. On campus every student has their own IP.

```bash
bash tests/load/run.sh                                  # the target: 50 users, 2 minutes
USERS=5 DURATION=10 bash tests/load/run.sh              # quick check that everything works
```

Other settings: `GATEWAY` (default `http://localhost:8080`), `RAMPUP` (10 s), `PER_USER_PER_MINUTE` (60).

To see the rate limit work, start the gateway without `RATE_LIMIT_PER_SECOND` and send more than 20 a second:

```bash
USERS=5 DURATION=15 RAMPUP=1 PER_USER_PER_MINUTE=600 bash tests/load/run.sh
```

About 20 requests a second get 201 and the rest get 429.

Close other heavy programs first (Docker builds, a running k3d cluster): the numbers are only as good as the laptop is quiet.

## Results

`docs/evidence/sonam-zangmo/sprint3/jmeter-report-submission.txt`. On 7 November 2026: 5,617 reports in 2 minutes (47 a second), 0 errors, p95 252 ms. Target met.
