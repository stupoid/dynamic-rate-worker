# Summary

This sets up a toy project where a worker is running a process every _x_ seconds.
This rate _x_ can be controlled via a http server running at port `:8090` by calling `/up` or `/down`.
The current stats can be retrieved by calling `/stats` and finally the whole server can be gracefully shut down with `/stop`.

This is to demonstrate setting up another plane of control for a worker where the rate at which jobs are being handled can be dynamically tuned via an accessible external api (e.g. HTTP calls).
