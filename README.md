The core utilities of https://github.com/xplshn/Andes.

### First-grade utilities for your Unix system!
A-Utils is a growing set of utilities, some are meant to replace the ones you are using right now while others are simple commands that are great to have, like `wttr` or `issue`.
The code in this repo is largely inspired by `u-root`, various commands like `ed` are direct modifications upon `u-root`'s implementations.

## Why?
I am tired of confusing BSD additions, GNU extension and flags added for compatibility with non-POSIX systems. I plan to implement every program that `u-root` lacks, or hasn't implemented as per the POSIX specifications, implementation deviations will be marked in the help pages/manual pages as a note stating: "This is NOT POSIX".
Also, I got bored, so I started writting utilities for fun.

#### Objectives
- Implement commands that are specific to SunOS/Solaris/Illumos.
- BSD commands that aren't in the spirit of BSD (bloat) should be implemented.
- Use "flag files", whenever feasible and in places that make sense. https://github.com/u-root/u-root/blob/pkg/uflag/flagfile.go
- The output of commands should be pretty, distinguishable and easy to follow-up if you've been in the terminal for way too much time, colors and other ANSI atributes should be used. Programs should not however use these special atributes when their output is being captured/redirected.

##### Rules
1. Avoid repetition. Won't implement commands which's functionality could be reduced to piping 2 or 3 commands together
2. Scripting is a priority, thus the commands MUST have reliable output

#### TODO
1. Implement interrupt channels (CTRL+C as defined behavior of the programs)

# ...
I am very much against bloat, but I enjoy challenges, which translates into me implementing things I shouldn't and slapping the label "Feature" on-top. If you ever find a so called "feature" like this, do tell me why it is feature creep.

##### Screenshots?
![image](https://github.com/user-attachments/assets/5eb85af5-e477-4b45-b2f9-8be342ab6e3e)
![image](https://github.com/user-attachments/assets/e6707b7e-d7b0-4c08-bf37-1fbcfbf55803)
![image](https://github.com/user-attachments/assets/2cd4b402-c189-4f30-b978-f828480087fd)


### Note
extendC.b contains a list of repositories that ./build.sh will clone, check out if they have a cbuild.sh script inside, if they do, it'll execute them and proceed to build.
extendGo.b is the same but for Go binaries, the Go repos don't need any kind of integration, they don't need to have cbuild.sh,
./build.sh builds the local ./cmd folder by default. You can use build.sh in other projects.
