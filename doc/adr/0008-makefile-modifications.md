# 8. Makefile modifications

Date: 2021-07-29

## Status

Accepted

## Context

`operator-sdk` have created `Makefile` as part of scaffolding. Throughtout development cycle for operator we needed to modify
behaviour of existing `Makefile`. One of the options could be maintenance of separate `Makefile.gitlab` to isolate scaffolding code from our own changes.

## Decision

It has been decided to maintain single `Makefile` with all of the original scaffolding and our changes for the time being for simplicity. 

## Consequences

We will rely on git history to resolve potential future conflicts (perhaps as part of operator-sdk upgrade)
