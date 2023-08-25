<!--
order: 0
title: "Metrics Overview"
parent:
  title: "metrics"
-->

# `metrics`


## Abstract

`x/metrics` is a stateless module that does not affect consensus. It captures chain metrics and emits them when the `instrumentation.prometheus` option is enabled in `config.toml`.

## Precision

The metrics emitted by `x/metrics` are `float64`s. They use `github.com/go-kit/kit/metrics` Prometheus gauges. Cosmos-sdk's `telemetry` package was not used because, at the time of writing, it only supports `float32`s and so does not maintain accurate representations of ints larger than ~16.8M. With `float64`s, integers may be accurately represented up to ~9e15.

## Metrics

The following metrics are defined:
* `cometbft_blocksync_latest_block_height` - this emulates the blocksync `latest_block_height` metric in CometBFT v0.38+. The `cometbft` namespace comes from the `instrumentation.namespace` config.toml value.

## Metric Labels

All metrics emitted have the labels defined in app.toml's `telemetry.global-labels` field. This is the same field used by cosmos-sdk's `telemetry` package.

example:
```toml
# app.toml
[telemetry]
global-labels = [
  ["chain_id", "kava_2222-10"],
  ["my_label", "my_value"],
]
```
