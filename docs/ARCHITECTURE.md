# Architecture Notes

## Services

The system is two independently deployable services:

- **Go JSON API** (`cmd/server`, `internal/*`): stateless, builds the OSRM matrix and solves the order.
- **Next.js frontend** (`frontend/`): App Router dashboard that proxies `/api/*` to the Go service.

## Backend layering

```
cmd/server  ->  handlers  ->  services  ->  { osrm.Client, optimizer.Solver }
                                        ->  validation
```

- `cmd/server` performs dependency wiring only. It constructs the OSRM client, the solver, the service, and the router, then runs the HTTP server with graceful shutdown.
- `handlers` owns transport concerns: decoding requests, mapping domain errors to HTTP status codes and friendly messages, and encoding JSON responses.
- `services.RouteService` orchestrates the workflow and holds no transport knowledge.
- `osrm` and `optimizer` are leaf packages exposing interfaces; they can be imported and reused independently.

## Error classification

The domain returns wrapped sentinel errors (`osrm.ErrServiceUnavailable`, `validation.ErrInsufficientStops`, `optimizer.ErrTooFewNodes`, ...). The handler uses `errors.Is` to select the correct HTTP status and a human-readable message, keeping presentation out of the domain.

## Optimization model

Given a duration matrix `D` of size `n x n`, a fixed start `s`, and a fixed end `e`, find a permutation of the interior nodes that minimizes:

```
cost(order) = sum_{i=0}^{len-2} D[order[i]][order[i+1]]
```

Construction uses nearest-neighbour from the start; refinement uses 2-opt segment reversals restricted to interior positions so `s` and `e` remain pinned. The loop is bounded by an iteration cap and checks `ctx.Err()` each pass.

## Extending to a real OR-Tools solver

Implement `optimizer.Solver` in a new type that shells out to an OR-Tools sidecar (or a CGO build) and register it in `cmd/server`. No other package changes.
