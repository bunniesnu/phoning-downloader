# Phoning Downloader

<a href="https://github.com/bunniesnu/phoning-downloader/releases"><img src="https://img.shields.io/github/release/bunniesnu/phoning-downloader.svg" alt="Latest Release"></a>
<a href="https://github.com/bunniesnu/phoning-downloader/actions"><img src="https://github.com/bunniesnu/phoning-downloader/actions/workflows/release.yml/badge.svg" alt="Build Status"></a>

A CLI downloader for calls from Phoning. Written in Go.

Please feel free to open an issue or a pull request.

## Installation

Download the [latest release](https://github.com/bunniesnu/phoning-downloader/releases/latest) binary/executable. Choose your OS and architecture to download.

* Supported OS: Windows, MacOS(darwin), Linux

* Supported Architectures: amd64(x86_64), arm64

On Windows, the system might detect the executable as virus. If you are concerned, try checking out the source code.

## Usage

You should get the API keys. Please [contact me](mailto:support@newjeans.app) for this.

If you have them, place them in the .env file in the following format.

```
API_KEY="*****"
SDK_KEY="*****"
```

The file should be in the folder where you execute the binary/executable. It will automatically verify whether you did it properly.

No further configuration required. You can change the download path such as:
```
phoning-downloader -o "your_download_path"
```

## Build

You can compile the binary/executable yourself. First, install [Go](https://go.dev/dl/) 1.24.4 on your system. Then, run the following commands.
```
git clone https://github.com/bunniesnu/phoning-downloader.git
cd phoning-downloader
go build
```
Then, you should have a binary named ```phoning-downloader``` or an executable named ```phoning-downloader.exe```.