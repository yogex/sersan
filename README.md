# Sersan

Sersan is an optimised Selenium/WebDriver hub for Kubernetes cluster written in Golang. It employs Kubernetes engine to manage the lifecycle and queue of browser containers, and ensures the test environment is always clean and fresh.
Each time it receives a new session request, Sersan will ask Kubernetes to create a new pod based on the browser and version that matches the requested capabilities. Once the pod is running, the next request will be forwarded to the created pod. The pod will be deleted if Sersan receives a delete session request.
The session ID is a JWT token containing, among other things, the original session ID from WebDriver, the pod's IP address, the Selenium/WebDriver port, and VNC port.

## Features

- Easy setup and maintenance. Only execute a few commands to get everything up and ready to receive high concurrent browser testing. Updating the browser version is as easy as editing a YAML file.
- One pod for one test. This ensures the test environment is always clean and fresh.
- Less memory consumption.
- Unlimited auto-scale. Only a single cluster is required to handle tests from all projects.
- Unified load distribution.
- Stateless. You can scale up or deploy a new version of Sersan without worrying about the currently running test.
- Compatible with Selenium/WebDriver test. No need to modify any of your existing tests.
- Support VNC viewer to observe the running browser. Use Selenium Chrome/Firefox debug image to use VNC.

## Prerequisites

1. Kubernetes cluster. You can use [minikube](https://github.com/kubernetes/minikube) or [kind](https://github.com/kubernetes-sigs/kind) for local development.

## Installation

To install simply use `kubectl` to apply resources.

```
kubectl apply -f manifests/
```

Check the sersan namespace (or the namespace you have specific in namespace: ) and make sure the pods are running.

That's all. You can now run your tests just like you would run it on Selenium hub. Point your WebDriver remote address to Sersan service ip.
```
kubectl get services -lapp=sersan
# WebDriver remote address
http://<service ip>:4444/wd/hub
```

## Development

Get Sersan source
```
go get -u https://github.com/salestock/sersan
```

Install all dependencies using glide
```
cd $GO_PATH/src/github.com/salestock/sersan
glide install
```

If you have (realize)[https://github.com/oxequa/realize] installed, simply run bellow command.
```
realize start --build --run --no-config
```

Build docker image
```
docker build -t <your domain>/sersan:latest .
```

Finally, redeploy Sersan application.

## Customisation

Sersan can be customised through environment variables. Supported environment variables are:

| Variable | Description | Default Value |
|----------|-------------|---------------|
|**PORT**|Sersan port.|`4444`|
|**BROWSER_CONFIG_FILE**|Custom browser config file.|`config/browsers.yaml`
|**STARTUP_TIMEOUT**|Timeout from new session request until Selenium/WebDriver is running.|`900000` (miliseconds)|
|**NEW_SESSION_ATTEMPT_TIMEOUT**|Timeout from pod running and new session created.|`60000` (miliseconds)|
|**RETRY_COUNT**|Number of attempts to create a session.|`30`|
|**SIGNING_KEY**|Session ID signing key|`secret_key`|
|**GRID_LABEL**|Browser's pod label.|`dev`|
|**NODE_SELECTOR_KEY**|Node selector key.||
|**NODE_SELECTOR_VALUE**|Node selector value.||
|**SERSAN_GRID_TIMEOUT**|Maximum age of the pod, after which it will be deleted automatically.|`300` (seconds)|
|**CPU_LIMIT**|CPU limit of browser containers.|`600m`|
|**CPU_REQUEST**|CPU request of browser containers.|`400m`|
|**MEMORY_LIMIT**|Memory limit of browser containers.|`600Mi`|
|**MEMORY_REQUEST**|Memory request of browser containers.|`1000Mi`|

## Browser Images

Sersan is compatible with the following Selenium standalone or selenoid browser images:
- [Selenium Chrome](https://hub.docker.com/r/selenium/standalone-chrome)
- [Selenium Chrome Debug](https://hub.docker.com/r/selenium/standalone-chrome-debug)
- [Selenium Firefox](https://hub.docker.com/r/selenium/standalone-firefox)
- [Selenium Firefox Debug](https://hub.docker.com/r/selenium/standalone-firefox-debug)
- [Selenoid Chrome](https://hub.docker.com/r/selenoid/chrome)
- [Selenoid Firefox](https://hub.docker.com/r/selenoid/firefox)

## Todos

- Sersan UI/CLI to manage running tests
- Support for Android and iOS emulator

## Contributing

Please read [CONTRIBUTING.md](https://github.com/salestock/sersan/CONTRIBUTING.md) for details on our code of conduct.

## Authors

* **Dimas Aryo** - *Initial work* - [dimasaryo](https://github.com/dimasaryo)

See also the list of [contributors](https://github.com/salestock/sersan/contributors) who participated in this project.

## License

This project is licensed under the Apache License 2.0 License - see the [LICENSE.md](LICENSE.md) file for details


## Acknowledgments

* Inspired by [Selenoid](https://github.com/aerokube/selenoid)
