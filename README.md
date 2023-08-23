# PacMacro 2023

This repository holds the source code for the PacMacro event during CSSS Frosh Week 2023 ([RETRO FROSH](https://sfucsss.org/events/frosh/2023)).

## Structure

This version of PacMacro consists of a **Go API** that is accessed under a `api/` directory, with the syntax `/api/(function)/(optional: inputs)`.
This API is interacted with and accessed through a simple and static **JavaScript frontend**.

## Deployment

For the deployment of PacMacro, the static files can be served wherever (e.g., GitHub Pages, or with the API), so long as the api calls are properly called.

### Example: Nginx on a Debian server

(To be written.)

## Building

### API

To build the PacMacro API, run `go build main` from the root directory of this repo.
Start the API in a detachable terminal (e.g., `tmux`), and ensure that your web server is proxying the API under a `api/` directory.

### Frontend

To build the PacMacro frontend, simply copy all files and directories under `htdocs` to wherever documents are served to the internet.
