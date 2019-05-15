# Sersan

Sersan is optimised selenium/webdriver hub for kubernetes cluster written in golang. It employes kubernetes engine to manage browsers container lifecycle and queue, ensure the test environment always clean and fresh.
Each time it receives new session request, sersan will ask kubernetes to create new pod based on the browser and version in the capabilities. Once the pod is running, the next request will be forwarded to the created pod. Pod will be deleted if sersan receives delete session request.
Session id is a jwt token contains some information includes original session id from webdriver, pod ip, selenium/webdriver port, and vnc port.

## Features

- Easy setup and maintenance. Only execute a few commands to get everything up and ready to receive high concurrent browser testing. Update browser version is easier as edit a yaml file.
- One pod for one test. Ensure the test environment always clean and fresh.
- Less memory consumption.
- Unlimited auto-scale. Only need a single cluster to handle test from all projects.
- Unified load distribution.
- Stateless. You can scale up or deploy new version of the sersan without worrying the running test.
- Compatible with selenium/webdriver test. No need modification of your existing test.
- Support VNC viewer to see the running browser. Use selenium chrome/firefox debug image to use VNC.


## Prerequisites

1. Kubernetes cluster. You may use [minikube](https://github.com/kubernetes/minikube) or [kind](https://github.com/kubernetes-sigs/kind) for local development.


## Installation

Create namespace for all sersan resources
```
kubectl create -f https://git.io/fjOiE
```

Create config map contains available browsers information. Refer to `config/browsers.yaml`
```
kubectl create -n sersan configmap sersan-browsers --from-file=https://git.io/fjOig
```

Deploy the sersan application
```
kubectl create -n sersan -f https://git.io/fjOi0
```

That's all. Now you can run your test just like you run it on selenium hub. Point your webdriver remote address to sersan service ip.
```
kubectl get services -lapp=sersan
# webdriver remote address
http://<service ip>:4444/wd/hub
```


## Development

Get sersan source
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

Then redeploy sersan application.


## Customisation

Sersan can be customised through environment variabel. Below is supported environment variables:
- **PORT**: Sersan port. Default is 4444.
- **BROWSER_CONFIG_FILE**: Custom browser config file. Default is `config/browsers.yaml`.
- **STARTUP_TIMEOUT**: Timeout from new session request until selenium/webdriver running. Default is 900000 (miliseconds).
- **NEW_SESSION_ATTEMPT_TIMEOUT**: Timeout from pod running and new session created. Default is 60000 (miliseconds).
- **RETRY_COUNT**: Number of create session attempt. Default is 30.
- **SIGNING_KEY**: Session id signing key. Default is secret_key.
- **GRID_LABEL**: Browser's pod label. Default is `dev`.
- **NODE_SELECTOR_KEY**: Node selector key. Default is empty.
- **NODE_SELECTOR_VALUE**: Node selector value. Default is empty.
- **SERSAN_GRID_TIMEOUT**: Pod will be deleted automatically if the age is more than timeout. Default is 300 (seconds).
- **CPU_LIMIT**: CPU limit of browsers container. Default is `600m`.
- **CPU_REQUEST**: CPU request of browsers container. Default is `400m`.
- **MEMORY_LIMIT**: Memory limit of browsers container. Default is `600Mi`.
- **MEMORY_REQUEST**: Memory request of browsers container. Default is `1000Mi`.


## Browsers Image

Sersan compatible with selenium standalone or selenoid browsers image:
- [Selenium Chrome](https://hub.docker.com/r/selenium/standalone-chrome)
- [Selenium Chrome Debug](https://hub.docker.com/r/selenium/standalone-chrome-debug)
- [Selenium Firefox](https://hub.docker.com/r/selenium/standalone-firefox)
- [Selenium Firefox Debug](https://hub.docker.com/r/selenium/standalone-firefox-debug)
- [Selenoid Chrome](https://hub.docker.com/r/selenoid/chrome)
- [Selenoid Firefox](https://hub.docker.com/r/selenoid/firefox)


## Todos

- Sersan UI/cli to manage running test
- Support Android and ios emulator


## Contributing

Please read [CONTRIBUTING.md](https://github.com/salestock/sersan/CONTRIBUTING.md) for details on our code of conduct.


## Authors

* **Dimas Aryo** - *Initial work* - [dimasaryo](https://github.com/dimasaryo)

See also the list of [contributors](https://github.com/salestock/sersan/contributors) who participated in this project.


## License

This project is licensed under the Apache License 2.0 License - see the [LICENSE.md](LICENSE.md) file for details


## Acknowledgments

* Inspired by [Selenoid](https://github.com/aerokube/selenoid)

