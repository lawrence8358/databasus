//go:build windows

// Package winservice runs the launcher as a Windows service: it registers and removes the service
// with the recovery settings the installer relies on, maps service control requests onto the
// launcher's context, and reports start failures to the Windows event log.
package winservice
