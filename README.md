# Golang Docker Container Link Example

I found many examples online wherein developers use the 
[Go client for the Docker Engine API](https://pkg.go.dev/github.com/docker/docker/client)
to start and stop containers. I couldn't find an example wherein two containers are
started and networked together such that they're able to communicate. So, this is that
example.

In `main.go` I am starting two containers, each of which is attached to a Docker
[user defined network](https://docs.docker.com/network/). One container is running
[go-httpbin](https://github.com/mccutchen/go-httpbin), which is a clone of [Kenneth Reitz][kr]'s
[httpbin][httpbin-org] service. The other container runs the [curl](https://hub.docker.com/r/curlimages/curl) HTTP client. `curl` is used to place an HTTP request to the `go-httpbin`
container. After the request, both containers are shut down.

## Running the code

You can run the code with `go run main.go` and you should see output like

```
{"status":"Pulling from curlimages/curl","id":"latest"}
{"status":"Digest: sha256:a3e534fced74aeea171c4b59082f265d66914d09a71062739e5c871ed108a46e"}
{"status":"Status: Image is up to date for curlimages/curl:latest"}
{"status":"Pulling from mccutchen/go-httpbin","id":"latest"}
{"status":"Digest: sha256:7e1cb9d38fe89c00359c5253e9841664c5b84c9df2bb0494e9c949f0e6aab46e"}
{"status":"Status: Image is up to date for mccutchen/go-httpbin:latest"}
  % Total    % Received % Xferd  Average Speed   Time    Time     Time  Current
                                 Dload  Upload   Total   Spent    Left  Speed
100   179  100   179    0     0  44750      0 --:--:-- --:--:-- --:--:-- 44750
{"args":{"foo":["bar"]},"headers":{"Accept":["*/*"],"Host":["httpbin:8080"],"User-Agent":["curl/7.74.0-DEV"]},"origin":"192.168.0.3:40894","url":"http://httpbin:8080/get?foo=bar"}
```

The final like is the JSON-formatted HTTP response from `curl`.