# Sersan

Sersan is optimised selenium/webdriver hub for kubernetes cluster written in golang.  


## Features
- Easy setup. Only need to deploy single binary to get everything up, running and ready to receive high concurrent browser testing.
- One pod for one test. Ensure the test environment always clean and fresh.
- Less memory consumption.
- Unlimited auto-scale.
- Unified load distribution.
- Stateless. You can scale up or deploy new version of the sersan without worrying the running test.
- Compatible with selenium/webdriver test. No need modification of your existing test.
- Support VNC viewer to see the running browser. Use selenium chrome/firefox debug image to use VNC. 


## Getting Started

These instructions will get you a copy of the project up and running on your local machine for development and testing purposes. See deployment for notes on how to deploy the project on a live system.


### Prerequisites

```
1. Kubernetes cluster. [minikube](https://github.com/kubernetes/minikube) or [kind](https://github.com/kubernetes-sigs/kind)
```


### Installation

Get the source
```
go get -u https://github.com/salestock/sersan

```

Create config map contains available browsers information. Refer to `config/browsers.yaml`
```
cd $GO_PATH/scr/github.com/salestock/sersan && \
kubectl create configmap --from-file=config/browsers.yaml
```

Deploy the sersan application
```
kubectl create -f deployment.yaml
```

That's all. Now you can run your test just like you run it on selenium hub. Point your webdriver remote address to sersan service.
```
kubectl get services -lapp=sersan
# webdriver remote address
http://<service ip>:4444/wd/hub
```


## Development

Install all dependencies
```
cd $GO_PATH/src/github.com/salestock/sersan
glide install
```
If you have (realize)[https://github.com/oxequa/realize] installed, simply run bellow command.
```
realize start --build --run --no-config
```


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

