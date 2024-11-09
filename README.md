The core utilities of https://github.com/xplshn/Andes.

### First-rate utilities for your Unix system!
A-Utils is a growing set of utilities, some are meant to replace the ones you are using right now while others are simple commands that are great to have, like `wttr` or `issue`.
The code in this repo is largely inspired by `u-root`, various commands like `ed` are direct modifications upon `u-root`'s implementations. A-Utils won't implement commands without any reason. If the command is already available in `u-root` and it is a suitable, POSIX compliant implementation, there isn't much reason to add it here without modifications.

## Why?
I am tired of confusing BSD additions, GNU extension and flags added for compatibility with non-POSIX systems. I plan to implement every program that `u-root` lacks, or hasn't implemented as per the POSIX specifications, implementation deviations will be marked in the help pages/manual pages as a note stating: "This is NOT POSIX".
Also, I got bored, so I started writting utilities for fun.

#### Objectives
- Implement commands that are specific to SunOS/Solaris/Illumos/Plan9/OtherNonGNUThings/NonBSDThings.
- BSD & GNU commands that aren't in the spirit of BSD nor GNU (bloat) should be implemented.
- Use "flag files", whenever feasible and in places that make sense. https://github.com/u-root/u-root/blob/pkg/uflag/flagfile.go. Config files suck.
- The output of commands should be pretty, distinguishable and easy to follow-up if you've been in the terminal for way too much time, colors and other ANSI atributes should be used. Programs should not however use these special atributes when their output is being captured/redirected.
- Transforming extended flags of commands, like BSD's cat -v, into independent programs that execute the specific functionalities of those flags, so as to keep the functionality, but in the spirit of Unix.
- Extend Unix commands without using flags nor changing behavior. For example, I might add colors to enhance the readability of some commands, but have that automatically turn off when our output is being piped or captured by/into another program.

##### Rules
1. Avoid repetition. Won't implement commands which's functionality could be reduced to piping 2 or 3 commands together
2. Scripting is a priority, thus the commands MUST have reliable output

#### TODO
1. Implement interrupt channels (CTRL+C as defined behavior of the programs) in programs that may need them

# ...
I am very much against bloat, but I enjoy challenges, which translates into me implementing things I shouldn't and slapping the label "Feature" on-top. If you ever find a so called "feature" like this, do tell me why it is feature creep.

##### Screenshots?
![image](https://github.com/user-attachments/assets/49469f2b-0ffc-4f81-961a-c08d5b470af1)
![image](https://github.com/user-attachments/assets/560dc83b-5354-4bf2-bf53-b110ab882237)

##### Future:
 - Should this project adopt all of U-Root's commands and improve upon them? Or should `a-utils`
 only have in its repo commands that have been modified/improved upon already?
 - Is making `u-root` commands behave as the POSIX implementation decreed them be something that should be upstreamed?

### Currently implemented:
1. Ed with syntax highlighting (ed from u-root but with syntax highlighting)
2. `catv`, the harmful features of `cat`
3. `ccat`, syntax highlighted cat
4. `fortune`, the cookie reader program, it gives you a random quote each time you execute it
5. `getconf`, the POSIX `getconf` program
6. `issue`, reads an "issue" file (/etc/issue or a user-suplied one)
7. `printf`, the POSIX `printf` command
8. `test`, the POSIX `test` command with some extensions for compatibility with "extended" shells and their scripts
9. `hpwd`, an implementation of `pwd` that replaces your $HOME with a prefix such as `~` or something else. For use with PS1 on strict POSIX shells
10. `wttr`, get's weather info from `wttr.in`
11. `cal`, an implementation of the POSIX `cal` command, as described in the OpenBSD manpage, that strives to have the same alignment as Busybox's and has (or tries to have) the Toybox's `cal` option to highlight specified days
12. `sed` implementation in Go, didn't add extensions. (doesn't comply, fails standard tests)
13. `fin`, file info command, from Torgo, created by "as" and now maintained by @xplshn (me)
14. `walk`, walk command, from Torgo, created by "as" and now maintained by @xplshn (me), this command walks a directory based on various rules and it print each file's full path to STDOUT
15. `listen`, listens and announces on a network port, ported from Torgo
16. `dial`, dials a network endpoint, ported from Torgo
17. `relf`, a program that shows you the sections of an ELF file and can also guess what the unknown "sections" contain
18. `isainfo`, like the SunOS command, it will show info about your architecture, capabilities, bits, etc.
19. `envsubst`, works like the built-in that many shells have. https://manned.org/man/alpine/envsubst
20. `importenv`, lets you access variables defined in the namespace/scope of another process
22. `noroot-do`, lets you `chroot` into a `rootfs` without root, it comes with various default presets of options, for different levels of sandboxing, etc, you can see those using the `--info` flag and you can also create/redifine your own `modes` :) (note: It doesn't create any external files, this program saves its config/options WITHIN itself, without rewritting itself, it uses XATTR arguments and a bit of magick)
23. `ntpdate` will simply update your HW clock based on `NTP`
24. `demo`, just a demo of the `ccmd` pkg/library

#### Contributions
I am looking for contributors, in fact. I NEED THEM. This is an ambicious project. And if you wish to contribute, please do! By openning a PR here to add a POSIX compliant utility that you've implemented, or bringing up to my attention bugs in our programs.

### Note on Building
A-Utils utilizes build.sh for building utilities. This script checks for cbuild.sh in cloned repositories listed in extendC.b and executes them accordingly. For Go binaries, extendGo.b serves a similar purpose but doesn't require cbuild.sh. By default, ./build.sh builds utilities in the local ./cmd folder but can be adapted for other projects. To build a single-file binary, you'll need: https://github.com/xplshn/a-utils_gobusybox

##### TODO:
- Publish my `grep` implementation
- Make `ccmd` be more useful and work like a replacement for the `"flag"` package, but with similar syntax. (as opposed to other alternatives which SUCK/are completely invasive)
- We need a package manager suite. One shall be developed, parting from an specification, and providing at least the following: a command used to create packages, and a command used to install, uninstall and get to KNOW available packages in a remote or local repository
- Transparently NETWORK everything
- Create a library for Go that works like overlayfs & uses fuse, but for use with the "embed" package.
