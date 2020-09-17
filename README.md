# simple-spa-server

_simple-spa-server_ is a simple static web server for serving SPAs (single page applications), allowing the `index.html` to use values provided through the environment. This allows you to deploy your UI as a [12-factor-app](https://12factor.net).

SPAs typically need to serve static files (JS, HTML, CSS, images) to a client. _simple-spa-server_ is a good fit, if your application only requires APIs external to the original host.

## Environment variables

To be usable in the spirit of a [12-factor-app](https://12factor.net), settings for simple-spa-server are taken from the environment:

| name             | default                     | description                                                                                                      |
| ---------------- | --------------------------- | ---------------------------------------------------------------------------------------------------------------- |
| SSS_APP_NAME     | `"simple-spa-server"`       | The application name displayed in the log during startup.                                                        |
| SSS_BINDADDR     | `:8080`                     | The address the server will bind to.                                                                             |
| SSS_DOCROOT      | `/data/docroot`             | The path to the document root, where the files are served from.                                                  |
| SSS_INDEXDOC     | `${SSS_DOCROOT}/index.html` | The path to the file used as the index document. (Inside this file, all environment variables will be replaced.) |
| SSS_USE_TLS      | `"0"`                       | Disable or enable TLS / HTTPS. Set to `"1"` to enable.                                                           |
| SSS_TLS_CERTPATH | `/data/conf/cert.pem`       | The path to the certificate used for HTTPS. Only used when `SSS_USE_TLS=1`.                                      |
| SSS_TLS_KEYPATH  | `/data/conf/key.pem`        | The path to the private key used for HTTPS. Only used when `SSS_USE_TLS=1`.                                      |

**Note:** Whilst _simple-spa-server_ does support HTTPS, it is much more likely, that you run it through a docker container, and your ingress controller or platform solution will take over the HTTPS part and proxy it to the container.

### Environment in the index document

The environment will also be used to substitute values inside the index document of the SPA. Substitution uses standard "shell like" rules, so you might have a section in your `index.html` that looks like this:

```html
<script>
  window.__EXAMPLE_APP__ = {
    api: "${API_URL}",
  };
</script>
```

**Note:** This substitution takes place during startup, and then the index document is cached for the runtime of _simple-spa-server_.

## Example scenario

These are the URLs involved:

- my-app.example.com: SPA serving the user interface for "my app"
- my-app-qa.example.com: SPA serving the internal staging/QA version of the user interface for "my app"
- api.example.com: Microservice (production) for the backend of "my app"
- api-qa.example.com: Microservice (staging/QA) for the backend of "my app"

Then `index.html` could look something like this:

```html
<!DOCTYPE html>
<html>
  <head>
    <!-- other stuff omitted for brevity -->
    <title>example app</title>
    <link rel="stylesheet" href="styles.css" />
  </head>
  <script>
    window.__EXAMPLE_APP__ = {
      api: "${API_URL}",
    };
  </script>
  <script src="example-app.min.js"></script>
</html>
```

```sh
# for production
SSS_DOCROOT=/path/to/example-app SSS_APP_NAME="example app (PROD)" API_URL=https://api.example.com simple-spa-server

# for QA
SSS_DOCROOT=/path/to/example-app-qa  SSS_APP_NAME="example app (QA)" API_URL=https://api-qa.example.com simple-spa-server
```

This works especially well inside docker containers.

## Using with docker

### Step 1: Build docker image for simple-spa-server

If you have access to the PSI internal docker registry, you can skip to step 3 and just use the image `docker.psi.ch:5000/simple-spa-server`.

Otherwise you need to build the image locally:

```sh
docker build . -t simple-spa-server
```

The above command will build `simple-spa-server` inside a container and just keep the resulting static binary in a minimal image. That docker image will be tagged as simple-spa-server for you locally, so you can use this name instead of the image ID (hash).

Verify that it worked by trying to run the docker image:

```sh
docker run --rm -i -t simple-spa-server
```

Now you **should see an error message**, that it cannot open `/data/docroot/index.html`. (Because that file does not exist inside the container.) That means, that the server did run successfully, so your image is ready for use.

### Step 2: Run the SPA through the docker image

Next, let's use this docker image we just built to run the SPA. For this example we'll assume, the complete SPA exists all bundled up and ready to be served from `$PWD/dist`.

```sh
docker run --rm -i -t -v $PWD/dist:/data/docroot -e SSS_APP_NAME="example SPA server" simple-spa-server
```

You should see a message saying "example SPA server starting up...".

### Step 3: Build docker image for the SPA

The last step is to use docker image as a base for running the SPA in.
This example assumes, that the SPA can be built by running `npm run build` (i.e. running the "build" script defined inside `package.json` of the SPA's npm project), and that the build creates a subdirectory `./dist` where all files are collected; so `./dist` would be the document root for a static web server.

Here's a `Dockerfile` that showcases this:

```
# example-spa Dockerfile

# ------------------------------------------------------------
# stage 1: build the ui
FROM node:12-alpine AS build-ui

COPY . /ui
WORKDIR /ui
RUN npm install
# assuming there is a "build" script in package.json that
# will build the SPA into a folder ./dist
RUN npm run build

# ------------------------------------------------------------
# stage 2: run the ui

# if you have access to PSI's private docker registry use this
# FROM docker.psi.ch:5000/simple-spa-server
FROM simple-spa-server

ENV SSS_APP_NAME "Example SPA"

COPY --from=build-ui /ui/dist /data/docroot

```

Then build an image by running the following command:

```sh
docker build . -t example-spa
```

Now you can run the SPA through simple-spa-server inside a container like so:

```sh
docker run --rm -i -t -p 3000:8080 example-spa
```
