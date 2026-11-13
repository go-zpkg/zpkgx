package app

import (
	"fmt"
	"os"
)

// `zpks` — what the registry has, and what you have.
//
// # WHY A SECOND BINARY RATHER THAN NINE MORE SUBCOMMANDS
//
// Upstream zpkx has one grammar and it has no verbs at all: the first
// argument that is not a flag is a PACKAGE, and zpkx runs it.
//
//	zpkx node@22 --version
//	zpkx ls                  # GNU ls, from gnu.org/coreutils
//
// This fork had added nine verbs to that position — ls, search, why, store,
// browse, catalog, env, compat, completion — and every one of them took a
// name away from the packages. Two did it to packages that exist: measured
// against the published linux/x86-64 catalogue on 2026-10-18, `gnu.org/
// coreutils` provides both `ls` and `env`. So `zpkx ls` ran a catalogue
// listing where upstream runs GNU ls, and there was no way to ask for the
// other one.
//
// ⛔ A RESERVED WORD CANNOT FIX THAT, which is the part worth writing down.
// The obvious repair is a single word to hang the verbs off — `zpkx pkg ls`.
// It fails for the same reason: `vercel.com/pkg` provides the command `pkg`.
// Any bare word can become a package tomorrow, because the whole point of
// the first position is that the registry decides what is in it. Only a
// token the grammar already reserves is safe, and the two kinds are a flag
// (which is what upstream chose for its own query mode, `zpkx -Q`) and a
// different binary.
//
// # WHY `zpks`
//
// It reads as a subject followed by a verb — `zpks search json`, `zpks why
// curl.se openssl.org` — which is the grammar being restored, and it keeps
// the family's shape beside zpkx, zpkm and zpke. Checked before it was
// chosen: nothing in either published catalogue provides a command called
// `zpks`, while `pkgctl` is a real binary in Arch's devtools and `pkgtool`
// one in Slackware — collisions outside the catalogue that the catalogue
// cannot report.
//
// It is one binary beside zpkx and not one repository beside it, because
// the two share the resolver, the catalogue reader and the store walker. A
// fork of the resolver is a fork of what a version MEANS.
const zpksUsage = `zpks dev — what the registry has, and what you have

usage:
  zpks ls [--tree] [--depth N] [node]
                                     what is available under a node. A
                                     NAMESPACE node lists what it contains,
                                     a PACKAGE node what it needs; a
                                     trailing slash asks for the namespace
                                     where a node is both (curl.se is).
                                     --tree descends, --depth bounds it.
                                     Read from the catalogue, so no network
  zpks ls --dependents [--tree] <pkg>
                                     the other direction: who NEEDS this.
                                     What you ask before CHANGING something
  zpks why <project> <dependency>    the shortest path by which one needs the
                                     other, as "nix why-depends" answers it:
                                     an explanation, not every route
  zpks search [-n N] <text>          find a package by its name or by a COMMAND
                                     it provides: "zpks search rg" finds
                                     crates.io/ripgrep, which no guess at the
                                     project name reaches. Offline
  zpks browse [node]                 walk the tree full-screen: arrows to move
                                     and descend, tab cycles names,
                                     dependencies and dependents, / searches,
                                     space pins, q to quit printing what you
                                     pinned — so zpkx +$(zpks browse) composes
  zpks catalog                       which catalogue is here, and how old
  zpks catalog update                fetch the catalogue of what this platform
                                     has, into $ZPKG_DIR/catalog. THE ONLY
                                     command that asks the registry what
                                     exists; ls and <TAB> then read the file,
                                     so they are instant and work offline
  zpks catalog update --platform os/arch
                                     the same, for a platform you are not on:
                                     linux/x86-64 from a mac. Writes to that
                                     platform's own file, so ls and <TAB>
                                     still read the host's
  zpks catalog reindex               rebuild the name index <TAB> reads from
                                     the catalogue already on disk, with NO
                                     network — for an image that ships one
  zpks request [--url] <project>     a package is missing: what the registry
                                     actually has, as a report you can read
                                     and then open as an issue. It SENDS
                                     NOTHING — no argv, no environment, no
                                     paths; the fields are fixed and printed
  zpks journal [-n N] [--clear]      what has failed HERE and how often: the
                                     packages this registry did not have, most
                                     asked for first. A local file, written on
                                     failure and read by you. It SENDS NOTHING
                                     either; ZPKG_JOURNAL=0 records nothing
  zpks catalog platforms             which platforms the registry publishes a
                                     catalogue for, one per line, ready to hand
                                     back to --platform. Asks the registry: no
                                     local file knows what EXISTS
  zpks builds <project>@<version>    which BUILDS of that version exist, here
                                     and in the registry, with the hash to
                                     name one: "<project>@<version>/<hash>".
                                     Builds are siblings within a version, so
                                     a version is required.
                                     --sort=status|closure|hash|installed —
                                     closure puts a build made against a
                                     later toolchain first, which is what a
                                     CVE rebuild produces. There is no date
                                     key: a published timestamp is pinned so
                                     a re-publish is byte-identical
  zpks store [--by-size] [-n N]      what the artifact store holds: every version
                                     an ephemeral run ever pulled, and its
                                     size. Reports and deletes NOTHING
  zpks store --root <lock> ...       and what any of it is FOR: a lock pins
                                     a whole closure, so it is a root set.
                                     Live against dead, as guix gc --list-live
                                     does — and like it, removes NOTHING
  zpks compat +<pin>... -- <pkg>...  how much of a package set resolves under
                                     those base versions, and what each refusal
                                     blames. Projects on stdin when none given
  zpks env init [--module]           print the zpke shell function, for
                                     eval "$(zpks env init)" in a profile:
                                     zpke load|unload|purge|list|save|restore
                                     --module also defines module() and ml(),
                                     where no other module system exists
  zpks env avail                     list the declared environments
  zpks env show <environment>        what an environment brings, and from where
  zpks env import <modulefile>...    convert Lmod (.lua) or Environment Modules
                                     (TCL) modulefiles into HCL2 environments,
                                     refusing anything it would have to guess at
  zpks completion bash|zsh|fish      the shell snippet for <TAB> completion, for
                                     BOTH zpks and zpkx. It asks the binaries,
                                     so it follows the registry without being
                                     regenerated
  zpks -h, --help                    show this help
  zpks -v, --version                 show version

  Running a package is zpkx's job, and it keeps upstream's grammar:
  "zpkx node@22 --version", "zpkx +gnu.org/make -- build.sh", "zpkx -Q git".

env:
  ZPKG_DIR     artifact store (default: ~/.zpkg)
  ZPKG_CATALOG read the catalogue from this file instead of $ZPKG_DIR/catalog —
               for an air-gapped image, or to inspect one before publishing.
               Named and unreadable is an error, not a quiet fallback
  ZPKG_DIST    artifact source (default: oci://ghcr.io/go-pkgx/packages)
  ZPKG_VERIFY  verify artifact signatures, fail-closed (default: on)
  ZPKG_DEPRECATED
               a JSON file of end-of-life release lines per project, which
               REPLACES the list built into this binary; {} turns it off.
               "zpks compat" resolves, so it reads the same list zpkx does
  ZPKG_WITHDRAWN
               a JSON file of BUILDS a CVE has caught, per project. Shown in
               the WHAT column of "zpks builds", with the hash that replaces
               each one. Ranked last among its siblings, never removed

  The variables keep the ZPKG_ spelling: they name a store and a registry
  that both binaries share, not a program.
`

// runVerb dispatches the verbs `zpks` owns, and reports whether argv[0] was
// one of them.
//
// ⛔ ONE DISPATCH, NOT TWO. It lives here rather than inline in Pkgs because
// `zpkx` has to be able to say something useful about these words without
// running them — see unknownVerb in main.go, which answers `zpkx ls` by
// pointing at `zpks ls` ONLY when the registry does not offer a package by
// that name. Two lists of nine words would disagree the first time one
// gained a tenth.
func runVerb(argv []string) (int, bool) {
	switch argv[0] {
	case "ls":
		return runLs(argv[1:], os.Stdout, os.Stderr), true
	case "catalog":
		return runCatalogCmd(argv[1:], os.Stdout, os.Stderr), true
	case "store":
		return runStore(argv[1:], os.Stdout, os.Stderr), true
	case "why":
		return runWhy(argv[1:], os.Stdout, os.Stderr), true
	case "builds":
		return runBuilds(argv[1:], os.Stdout, os.Stderr), true
	case "search":
		return runSearch(argv[1:], os.Stdout, os.Stderr), true
	case "request":
		return runRequest(argv[1:], os.Stdout, os.Stderr), true
	case "journal":
		return runJournal(argv[1:], os.Stdout, os.Stderr), true
	case "browse":
		return runBrowse(argv[1:], os.Stdin, os.Stdout, os.Stderr), true
	case "completion":
		if len(argv) < 2 {
			fmt.Fprintln(os.Stderr, pfx()+"usage: "+tool+" completion bash|zsh|fish")
			return 2, true
		}
		return runCompletionSnippet(argv[1], os.Stdout, os.Stderr), true
	case "compat":
		// How far does a chosen base reach? See runCompat.
		plus, rest := splitPlus(argv[1:])
		if err := runCompat(plus, rest, os.Stdin, os.Stdout); err != nil {
			if err != errSilent {
				fmt.Fprintln(os.Stderr, pfx()+err.Error())
			}
			return 1, true
		}
		return 0, true
	case "env":
		return runEnvCmd(argv), true
	}
	return 0, false
}

// isVerb reports whether a word is one of zpks's verbs.
//
// ⛔ IT READS subcommandCompletions, the list completion already offers,
// rather than a second slice of the same nine words. I wrote that second
// slice first, two functions after a comment saying two lists would
// disagree the first time one gained a tenth entry. The one that exists is
// the one <TAB> shows, so it is the one that cannot be forgotten.
func isVerb(s string) bool {
	for _, c := range subcommandCompletions {
		if c.value == s {
			return true
		}
	}
	return false
}

// runEnvCmd is `zpks env …`, the environment front-end's one entry point. A
// verb rather than four top-level flags: the flags are not options on
// anything, they are a different job.
func runEnvCmd(argv []string) int {
	if len(argv) == 1 {
		fmt.Fprintln(os.Stderr, pfx()+"usage: "+tool+" env init|load|unload|purge|avail|show|import")
		return 2
	}
	var err error
	switch argv[1] {
	case "init":
		err = modeShellInit(len(argv) > 2 && argv[2] == "--module", os.Stdout)
	case "load", "unload", "purge":
		err = modeShell(argv[1], argv[2:], composeSpecs, osLookupEnv, os.Stdout)
	case "avail":
		err = modeAvail(os.Stdout)
	case "import":
		if len(argv) < 3 {
			fmt.Fprintln(os.Stderr, pfx()+"usage: "+tool+" env import <modulefile>...")
			return 2
		}
		err = modeImport(argv[2:], os.Stdout, os.Stderr)
	case "show":
		if len(argv) < 3 {
			fmt.Fprintln(os.Stderr, pfx()+"usage: "+tool+" env show <environment>")
			return 2
		}
		err = modeShow(argv[2], os.Stdout)
	default:
		fmt.Fprintln(os.Stderr, pfx()+"unknown env command "+argv[1]+" (init|load|unload|purge|avail|show|import)")
		return 2
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, pfx()+err.Error())
		return 1
	}
	return 0
}

// Pkgs is the `zpks` command.
//
// It sets `tool` first and before anything can print, because every
// diagnostic in this package names the binary the user typed and there is
// exactly one moment at which that is known.
func Pkgs(argv []string) int {
	tool = "zpks"
	artifactWarn()
	if err := configError(); err != nil {
		fmt.Fprintln(os.Stderr, pfx()+"warning: ignoring ~/.zpkg/config.hcl2: "+err.Error())
	}
	if len(argv) == 0 {
		fmt.Fprint(os.Stderr, zpksUsage)
		return 2
	}
	// Before anything else, for the same reason zpkx does it: a completion
	// must not install, fetch or run a thing.
	if maybeComplete(argv, os.Getenv, os.Stdout) {
		return 0
	}
	if code, handled := runVerb(argv); handled {
		return code
	}
	switch argv[0] {
	case "-h", "--help":
		fmt.Print(zpksUsage)
		return 0
	case "-v", "--version":
		fmt.Println("zpks " + version)
		return 0
	}
	// ⛔ IT DOES NOT FALL BACK TO RUNNING THE PACKAGE. `zpks node` is a
	// question nobody asked — zpks answers ABOUT packages and zpkx runs
	// them — and guessing which was meant is how the grammar got muddled in
	// the first place. The refusal names the other binary.
	fmt.Fprintf(os.Stderr, "%s%q is not a zpks command. To RUN a package: zpkx %s\n", pfx(), argv[0], argv[0])
	fmt.Fprint(os.Stderr, zpksUsage)
	return 2
}
