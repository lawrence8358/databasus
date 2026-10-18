//go:build windows

// Package stopevent lets the Windows launcher ask the backend to shut down gracefully. Windows has
// no SIGTERM, and a console control event only reaches processes sharing a console; a console
// shared with a service also delivers CTRL_LOGOFF_EVENT whenever any user signs out, which Go
// reports as SIGTERM. An inheritable event handed to the child needs no console at all.
package stopevent
