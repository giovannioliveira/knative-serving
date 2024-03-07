# Tracexec
Execute HTTP requests based on a provided tracefile with high precision [O(uS)].
## Executing
The application source code is placed at [`tracexec.go`](tracexec.go). After compiled, it generates a `tracexec` executable binary.
You must run the script `run.ssh`, which can be edited to meet the desired arguments given as inline environmental variables. This program requires superuser permission to prioritize its process with `nice` and increase the benchmark's precision. 
## Cofiguring

| NAME    | DEFAULT                      | TYPE          | DESCRIPTION                                         |
|---------|------------------------------|---------------|-----------------------------------------------------|
| TRACE   | invokes.csv                  | str           | Trace source                                        |
| DUTY    | 0.5                          | float         | Busy wait duration ratio                            |
| URL     | [cloud svc endpoint]         | str           | FaaS public URL                                     |
| BEGIN   | 2 minutes ahead (floor secs) | str (RFC3339) | Experiment begin schedule datetime                  |
| DBGFUNC | None                         | str           | Ignore FID and call the same function for debugging |
| OUTDIR  | ./logs/                      | str           | Directory for output log files                      |
| INITRID | 0                            | int           | Initial RID for current execution                   |
| ENDRID  | -1 (last record)             | int           | Final RID for current execution                     |

## Response
Values from this section are integer.
- **rt0**=[init_func_unix_ns] := Request processing start in Unix nS
- **rtb**=[real_busy_time_ns] := Time spent at the busy stage in nS
- **rit**=[real_busy_iterations] := Number of iterations completed at the busy stage
- **rts**=[real_idle_time_ns] := Time spent at the idle stage in nS
- **rdt**=[real_duration_ns] := Total function execution time in nS
- **rtf**=[final_func_unix_ns] := Request processing end in nS
## Development
### Run
Run development versions locally with `func run` (the Knative Function tool).
### Testing
TODO: write tests.
### Commit changes
After testing, generate a new commit with updated source code and documentation.
## Build & Deployment
### Build & Push
- Compile the source code with `go build tracexec.go`.
- Push the built binary to the experiment environment.
  > **Note**: you can also download and use the latest public dist build for Linux x86-64 at https://github.com/giovannioliveira/knative-serving/raw/development/custom/tracexec/tracexec`
## Additional fields
Before saving the request log, which contains data from request parameters and response, we also include the following fields:

| NAME | TYPE | DESCRIPTION                                            |
|------|------|--------------------------------------------------------|
| dt0  | int  | ScheduledReqInit - MeasuredReqInit timestamp in UnixNs |
| dtd  | int  | TargetDuration - MeasuredDuration in nS                |
| t0c  | int  | MeasuredReqInit timestamp in UnixNs                    |
| tfc  | int  | MeasuredReqEnd timestamp in UnixNs                     |
