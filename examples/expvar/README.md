# metrics example

## Running the example
Build the example application using `go build`.

Supply additional flags `port` and `workers` to specify the port that the HTTP
server will run on and the number of workers.

1. `go build .`
2. `./expvar -port 8080 -workers 8`
3. Navigate to http://localhost:8080 to see the various metrics exposed by the expvar
package (including taskrunner's own metrics).
