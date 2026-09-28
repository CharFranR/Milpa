package repository

import "embed"

// Migrations carries every SQL migration compiled into the binary.
//
// The embed has to live in this package, next to the migrations directory:
// //go:embed patterns may not contain ".." and are resolved against the
// directory of the file that declares them, so the database package — which
// owns the migration runner — cannot reach a sibling subtree without going
// through here.
//
// Compiling the migrations in is what makes booting independent of the working
// directory. The runner used to read them from a relative filesystem path,
// which only resolved when the process happened to start inside the module
// root; any other working directory failed to boot, and shipping the SQL files
// alongside the binary in the container image was required to make that work.
//
//go:embed migrations/*.sql
var Migrations embed.FS
