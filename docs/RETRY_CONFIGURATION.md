# Retry Configuration

The provider automatically retries rate-limited (429) and transient server-error (5xx) responses using exponential backoff with jitter. Connection errors on idempotent requests (GET, PUT, DELETE) are also retried.

## Default behaviour

| Setting | Default | Description |
|---|---|---|
| Strategy | `backoff` | Exponential backoff with ±25% jitter |
| Initial interval | 500 ms | Wait before the first retry |
| Max interval | 60 s | Ceiling on any single wait; `Retry-After` values are capped here too |
| Max elapsed time | 300 s (5 min) | Total budget across all attempts; give up when this is exhausted |
| Retry connection errors | true | Retries broken pipe / connection reset on idempotent methods |

When the provider is waiting it logs a warning at the Terraform log level:

```
WARN API rate limited, waiting before retry
  wait_seconds = 12.5
  attempt      = 2
```

Enable Terraform logging with `TF_LOG=WARN` (or `DEBUG`) to see these messages.

## Tuning

### Provider attributes

```hcl
provider "netskope" {
  retry_max_elapsed_time = 30   # seconds — fail fast in CI
  retry_disabled         = true # disable retries entirely
}
```

### Environment variables

```bash
export NETSKOPE_RETRY_MAX_ELAPSED=30   # seconds
export NETSKOPE_RETRY_DISABLED=true
```

Provider attributes take precedence over environment variables when both are set.

## Recommended settings by context

| Context | Setting | Reason |
|---|---|---|
| Interactive apply | default (300 s) | Allows time for transient rate limits to clear |
| CI pipeline | `retry_max_elapsed_time = 30` | Fails fast; surface errors rather than blocking the pipeline |
| Parallel acceptance tests | default | Tests run with `-parallel 8`; rate limits are normal and expected |
| Disabled (debugging) | `retry_disabled = true` | Returns raw API responses immediately for diagnosis |

## How the budget is enforced

Before sleeping, the provider checks whether the proposed wait would exceed the remaining budget. If it would, the request fails immediately rather than sleeping and then failing — this prevents a large `Retry-After` value from blocking a pipeline for its full duration when time has already run out.

The `Retry-After` header value is also capped to `MaxInterval` (60 s by default), so a misbehaving API cannot force a wait longer than the configured ceiling regardless of what it returns.
