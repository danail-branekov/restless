# Restless

`restless` is a utility to temporarily disable system suspend while a process
is running. This is useful to ensure that long running tasks (e.g. compiling a
huge application) are not be paused while you are away living your real life.

It sends [dbus](https://www.freedesktop.org/wiki/Software/dbus/) messages,
therefore it only works on Linux (unless other operating systems support dbus).

I have only tested it on KDE Plasma but it should also work on other desktop
environments.

## How to build it

Clone this repository and run `make`. The binary is built into the current
directory.

## How to use it

`restless your-long-running-app arg1 arg2...`

## How does it work

1. It connects to dbus
1. Invokes the `org.freedesktop.PowerManagement.Inhibit.Inhibit` dbus method
1. Runs the long running application as a child process making sure to wire
   standard IO correctly
1. Once the long running application exits or is aborted, `restless` would
   invoke `org.freedesktop.PowerManagement.Inhibit.UnInhibit` method to let the
   system know that it can go back to the configured power management policy

An important detail for systems using
[powerdevil](https://github.com/KDE/powerdevil) (KDE's power management
facility) is that dbus connection must be kept open while the background
process runs. If it gets closed immediately after calling the `Inhibit` method,
`powerdevil` ignores the inhibit request. Thanks to `d_ed`'s answer on this
[reddit
thread](https://www.reddit.com/r/kde/comments/hruubo/unable_to_inhibit_suspend/)
to bring me to [this
commit](https://github.com/KDE/powerdevil/commit/d21102cc6c7a4db204a29f376ce5eb316ef57a6e)
in powerdevil.

## What to do while waiting the task to complete

You may give Leprous' [Restless](https://www.youtube.com/watch?v=986iAyQpr1U) a
listen. It inspired the name of this project.
