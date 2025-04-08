# OpenTelemetry Golang Lib

[![License](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)

## Description

OpenTelemetry Golang Lib is a Golang library created to facilitate the use of OpenTelemetry, a set of tools and standards for instrumentation and observability of applications.

## Usage

It is necessary to configure the environment variables. You can find an example of these variables in the `env.sample` file. Make sure to configure the variables correctly according to your environment.

To use the library, you must always reference it through a specific version, never via the `main` branch. This is important to avoid backwards compatibility breaks if the `main` branch is changed/evolved. To do this, simply type the url adding `@<release>`, as in the example: 'go get "github.com/asamedeiros/opentelemetry-golang-lib/otelconfig@v1.0.0"'

We recommend always using the latest version of the library to get the latest features and bug fixes.

## Contribution

If you want to contribute to the OpenTelemetry Golang Lib, feel free to submit a pull request.