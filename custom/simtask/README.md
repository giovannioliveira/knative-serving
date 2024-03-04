# Simtask
Simulate a task with parameters for idle and busy wait during its execution.
## Request
The application source code is placed at [`handle.go`](handle.go). After deployed, a function is called using a HTTP request of the form:
```
curl -v --header "Host: [function_id].default.knative.dev"\
 "http://200.144.244.220:10080/?cl=[clent_id]&id=[request_id]&t0=[experiment_ts_ns]&ts=[sleep_duration_ns]&tb=[busy_duration_ns]&[custom_arg_x]=[custom_value_x]"
```
## Parameters
- **function_id** := Called function unique identifier
- **200.144.244.220:10080** := Public IP for FaaS serving
  - **10.4.0.143:10080** := Cloud private IP for FaaS serving
- **cl**=[client_id] := Client unique identifier
- **rid**=[request_id] := Request unique identifier
- **t0**=[experiment_ts_ns] := Virtual timestamp with experiment begin offset
- **ts**=[sleep_time_ns] := Target idle wait duration in nanoseconds
- **tb**=[busy_time_ns] := Target busy wait duration in nanoseconds
  - XOR **it** := Target iteration number.
- **custom_key_x**=[custom_value_x] := Client defined key-value pairs (it can be used multiple times for the distinct keys)
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
Develop new features by adding a test to [`handle_test.go`](handle_test.go) for
each feature, and confirm it works with `go test`. TODO: update tests.
### Commit changes
After testing, generate a new commit with updated source code and documentation.
## Build & Deployment
### Build & Push
- Build the new source code with `func build`.
- Push the built version to DockerHub with `docker push docker.io/[your-account]/simtask`.
    > **Note**: you can also pull and use the public image from DockerHub with `docker push docker.io/giovanniapsoliveira/simtask`
### Deploy at you Knative cluster
- Create a new Knative service with the pulled image:
   ```
    kn service create svcname \
	 --image docker.io/giovanniapsoliveira/simtask \
	 --port 8080 
   ```
  - you can change "svcname" to the desired service name and "giovanniapsoliveira" to your own DockerHub account name.
- Test with the curl command as described at the usage section.
