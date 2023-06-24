# Statigo

An experiment to see how small a memory footprint we can get to for a simple static HTTP server.

As features grow it is getting bigger but the goal is still to keep it as small as possible.

Docker images available on DockerHub: https://hub.docker.com/r/stut/statigo

For a React site or similar web application set `--not-found-filename index.html` on the command line.

Iterations:

* v1: Basic static HTTP server.
* v2: Added prometheus metrics, healthcheck URL (/health by default), and custom 404 content (404.html by default).
* v3: Return 404 for dodgy-looking requests.
* v4: More dodgy-looking requests now get a 404.
* v5: Return 404 for directory list requests. Added Apache-style request logging to stdout (enabled by default).
* v6: Log the IP address from X-Forwarded-For if present.
* v7: Improved (corrected) Apache-style request logging.
* v8: Default to not serving hidden files and folders.
