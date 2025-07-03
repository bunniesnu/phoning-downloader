# Phoning Downloader

A CLI downloader for calls from Phoning. Written in Go.

## Installation


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
git clone https://github.com/phoning-tools/phoning-downloader.git
cd phoning-downloader
go build
```
Then, you should have a binary named ```phoning-downloader``` or an executable named ```phoning-downloader.exe```.