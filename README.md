# WASAText

This repository contains the completed project for the [Web and Software Architecture](http://gamificationlab.uniroma1.it/en/wasa/) homework.

The application was built starting from the "Fantastic coffee (decaffeinated)" basic structure provided in class, which has been fully implemented and adapted into a functional messaging platform for the course evaluation.

## Project structure

* `cmd/` contains all executables; Go programs here should only do "executable-stuff", like reading options from the CLI/env, etc.
	* `cmd/healthcheck` is an example of a daemon for checking the health of servers daemons; useful when the hypervisor is not providing HTTP readiness/liveness probes (e.g., Docker engine)
	* `cmd/webapi` contains an example of a web API server daemon
* `demo/` contains a demo config file
* `doc/` contains the documentation (usually, for APIs, this means an OpenAPI file)
* `service/` has all packages for implementing project-specific functionalities
	* `service/api` contains an example of an API server
	* `service/globaltime` contains a wrapper package for `time.Time` (useful in unit testing)
* `vendor/` is managed by Go, and contains a copy of all dependencies
* `webui/` is an example of a web frontend in Vue.js; it includes:
	* Bootstrap JavaScript framework
	* a customized version of "Bootstrap dashboard" template
	* feather icons as SVG
	* Go code for release embedding

Other project files include:
* `open-node.sh` starts a new (temporary) container using `node:20` image for safe and secure web frontend development (you don't want to use `node` in your system, do you?).

## Go vendoring

This project uses [Go Vendoring](https://go.dev/ref/mod#vendoring). You must use `go mod vendor` after changing some dependency (`go get` or `go mod tidy`) and add all files under `vendor/` directory in your commit.

For more information about vendoring:

* https://go.dev/ref/mod#vendoring
* https://www.ardanlabs.com/blog/2020/04/modules-06-vendoring.html

## Node/YARN vendoring

This repository uses `yarn` and a vendoring technique that exploits the ["Offline mirror"](https://yarnpkg.com/features/caching). As for the Go vendoring, the dependencies are inside the repository.

You should commit the files inside the `.yarn` directory.

## How to build

If you're not using the WebUI, or if you don't want to embed the WebUI into the final executable, then:

```shell
go build ./cmd/webapi/
```

If you're using the WebUI and you want to embed it into the final executable:

```shell
./open-node.sh
# (here you're inside the container)
yarn run build-embed
exit
# (outside the container)
go build -tags webui ./cmd/webapi/
```

## How to run (in development mode)

You can launch the backend only using:

```shell
go run ./cmd/webapi/
```

If you want to launch the WebUI, open a new tab and launch:

```shell
./open-node.sh
# (here you're inside the container)
yarn run dev
```

## How to build for production

```shell
./open-node.sh
# (here you're inside the container)
yarn run build-prod
```

## Known issues

### My build works when I use `yarn run dev`, however there is a Javascript crash in production/grading

Some errors in the code are somehow not shown in `vite` development mode. To preview the code that will be used in production/grading settings, use the following commands:

```shell
./open-node.sh
# (here you're inside the container)
yarn run build-prod
yarn run preview
```

## License

See [LICENSE](LICENSE).
