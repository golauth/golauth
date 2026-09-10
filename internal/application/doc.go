// Package application is the use-case layer. Its subpackages (token, user,
// role, keys, ...) orchestrate domain entities and repository interfaces and
// must not depend on the delivery mechanism: nothing under pkg/application may
// import pkg/infra. layering_test.go enforces that.
package application
