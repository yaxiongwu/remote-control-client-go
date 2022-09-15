# remote-control-client-go
功能探讨：
1.提供视频、音频、直播等功能，可配置
2.能在网页上配置树莓派的端口，第一次连接后，发送配置信息到网页，网页能提供相应的控制界面

1.远程能控制终端，如设置某一个IO口为输出，这需要一个高权限的datachannnel
2.一路是视屏加控制，另一路是旁观
3.需要完整的控制流程，有加入、断开、断线处理
4.人员管理，排队想控制的人，排队想旁观的人，如何管理？


工作流程：
1.视频源向服务器发register,role=videosoure登记新视频源，这个client需要标记为视频源,client需要state来标记是视频源、等待的控制端、正在控制、等待旁观、正在旁观等
2.控制方（APP）向服务器发join,不带sdp,轮到连接的时候才发送sdp,但是这个sdp需要设置检测超时，所有要跟其他的sdp分开处理
3.旁观方(App)的控制跟控制方类似

声音：

oto

## Prerequisite

On some platforms you will need a C/C++ compiler in your path that Go can use.

- macOS: On newer macOS versions type `clang` on your terminal and a dialog with installation instructions will appear if you don't have it
  - If you get an error with clang use xcode instead `xcode-select --install`
- Linux and other Unix systems: Should be installed by default, but if not try [GCC](https://gcc.gnu.org/) or [Clang](https://releases.llvm.org/download.html)

### macOS

Oto requires `AudioToolbox.framework`, but this is automatically linked.

### iOS

Oto requires these frameworks:

- `AVFoundation.framework`
- `AudioToolbox.framework`

Add them to "Linked Frameworks and Libraries" on your Xcode project.

### Linux

ALSA is required. On Ubuntu or Debian, run this command:

```sh
apt install libasound2-dev
```

In most cases this command must be run by root user or through `sudo` command.

### FreeBSD, OpenBSD

BSD systems are not tested well. If ALSA works, Oto should work.

## Usage

The two main components of Oto are a `Context` and `Players`. The context handles interactions with
the OS and audio drivers, and as such there can only be **one** context in your program.

From a context you can create any number of different players, where each player is given an `io.Reader` that
it reads bytes representing sounds from and plays.

Note that a single `io.Reader` must **not** be used by multiple players.

opus

### API Docs

Go wrapper API reference:
https://godoc.org/gopkg.in/hraban/opus.v2

Full libopus C API reference:
https://www.opus-codec.org/docs/opus_api-1.1.3/

For more examples, see the `_test.go` files.

## Build & Installation

This package requires libopus and libopusfile development packages to be
installed on your system. These are available on Debian based systems from
aptitude as `libopus-dev` and `libopusfile-dev`, and on Mac OS X from homebrew.
They are linked into the app using pkg-config.

Debian, Ubuntu, ...:
```sh
sudo apt-get install pkg-config libopus-dev libopusfile-dev
```

Mac:
```sh
brew install pkg-config opus opusfile
```




### Building Without `libopusfile`

This package can be built without `libopusfile` by using the build tag `nolibopusfile`.
This enables the compilation of statically-linked binaries with no external
dependencies on operating systems without a static `libopusfile`, such as
[Alpine Linux](https://pkgs.alpinelinux.org/contents?branch=edge&name=opusfile-dev&arch=x86_64&repo=main).

**Note:** this will disable all file and `Stream` APIs.

To enable this feature, add `-tags nolibopusfile` to your `go build` or `go test` commands:

```sh
# Build
go build -tags nolibopusfile ...

# Test
go test -tags nolibopusfile ./...
```