The core utilities of https://github.com/xplshn/Andes.

### First-rate utilities for your Unix system!
A-Utils is a growing set of utilities, some are meant to replace the ones you are using right now while others are simple commands that are great to have, like `wttr` or `issue`.
The code in this repo is largely inspired by `u-root`, various commands like `ed` are direct modifications upon `u-root`'s implementations.

## Why?
I am tired of confusing BSD additions, GNU extension and flags added for compatibility with non-POSIX systems. I plan to implement every program that `u-root` lacks, or hasn't implemented as per the POSIX specifications, implementation deviations will be marked in the help pages/manual pages as a note stating: "This is NOT POSIX".
Also, I got bored, so I started writting utilities for fun.

#### Objectives
- Implement commands that are specific to SunOS/Solaris/Illumos/Plan9/OtherNonGNUThings.
- BSD commands that aren't in the spirit of BSD (bloat) should be implemented.
- Use "flag files", whenever feasible and in places that make sense. https://github.com/u-root/u-root/blob/pkg/uflag/flagfile.go
- The output of commands should be pretty, distinguishable and easy to follow-up if you've been in the terminal for way too much time, colors and other ANSI atributes should be used. Programs should not however use these special atributes when their output is being captured/redirected.
- Transforming extended flags of commands, like BSD's cat -v, into independent programs that execute the specific functionalities of those flags, so as to keep Unix clones in the spirit of Unix.

##### Rules
1. Avoid repetition. Won't implement commands which's functionality could be reduced to piping 2 or 3 commands together
2. Scripting is a priority, thus the commands MUST have reliable output

#### TODO
1. Implement interrupt channels (CTRL+C as defined behavior of the programs) in programs that may need them
2. The `cat -x` flag in cat.go should be removed, the program for visualizing text with syntax highlighting should be `ccat`

# ...
I am very much against bloat, but I enjoy challenges, which translates into me implementing things I shouldn't and slapping the label "Feature" on-top. If you ever find a so called "feature" like this, do tell me why it is feature creep.

##### Screenshots?
![image](https://github.com/user-attachments/assets/49469f2b-0ffc-4f81-961a-c08d5b470af1)
![image](https://github.com/user-attachments/assets/560dc83b-5354-4bf2-bf53-b110ab882237)

##### Future:
 - Should this project adopt all of U-Root's commands and improve upon them? Or should `a-utils`
 only have in its repo commands that have been modified/improved upon already?
 - Is making `u-root` commands behave as the POSIX implementation decreed them be something that should be upstreamed?

### Note on Building
A-Utils utilizes build.sh for building utilities. This script checks for cbuild.sh in cloned repositories listed in extendC.b and executes them accordingly. For Go binaries, extendGo.b serves a similar purpose but doesn't require cbuild.sh. By default, ./build.sh builds utilities in the local ./cmd folder but can be adapted for other projects.
