# vpcflow-grapherd #

## Overview ##

## Modules ##

The service struct in the grapherd package contains the modules used by this application. If none of these modules are configured, 
the built-in modules will be used.

```
func main() {
	...
	
	// Service created with default modules
	service := &grapherd.Service{
		Middleware: middleware,
	}
	
	...
} 
```

### Logging ###

This project uses [logevent](https://bitbucket.org/atlassian/logevent) as its logging interface. Structured logs that this project emits
can be found in the `logs` package. This project comes with a couple of default logging implementations that can be found in the plugins
package. These loggers are injected via HTTP middleware on the request context.

```
func main() {
	router := chi.NewRouter()
	middleware := []func(http.Handler) http.Handler{
		plugins.DefaultLogMiddleware(), // injects a logger which sends to os.Stdout
	}
	service := &grapherd.Service{Middleware: middleware}
	if err := service.BindRoutes(router); err != nil {
		panic(err.Error())
	}
}
```
Please note that this project will not run without some sort of logger being installed. While it's not recommended, if you wish to omit
logging, use the `NopLogMiddleware`.

### Stats ###

This project uses [xstats](https://github.com/rs/xstats) as the stats client. It supports a decent range of backends. The default stats
backend for the project is statsd using the datadog tagging extensions. The default backend will send stats to "localhost:8126". To change
the destination or the backend install the `CustomStatMiddleware` with your own xstats client.

### ExitSignals ###

Exit signals in this project are used to signal the service to perform a graceful shutdown. The built-in exit signal listens for SIGTERM and SIGINT and signals to the main routine to shutdown the service.

## Setup ##

## Contributing ##

### License ###

This project is licensed under Apache 2.0. See LICENSE.txt for details.

### Contributing Agreement ###

Atlassian requires signing a contributor's agreement before we can accept a
patch. If you are an individual you can fill out the
[individual CLA](https://na2.docusign.net/Member/PowerFormSigning.aspx?PowerFormId=3f94fbdc-2fbe-46ac-b14c-5d152700ae5d).
If you are contributing on behalf of your company then please fill out the
[corporate CLA](https://na2.docusign.net/Member/PowerFormSigning.aspx?PowerFormId=e1c17c66-ca4d-4aab-a953-2c231af4a20b).