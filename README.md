The core utilities of https://github.com/xplshn/Andes.

### Note
extendC.b contains a list of repositories that ./build.sh will clone, check out if they have a cbuild.sh script inside, if they do, it'll execute them and proceed to build.
extendGo.b is the same but for Go binaries, the Go repos don't need any kind of integration, they don't need to have cbuild.sh,
./build.sh builds the local ./cmd folder by default. You can use build.sh in other projects.
