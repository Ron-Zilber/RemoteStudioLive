# RemoteStudioLive

RemoteStudioLive is a real-time audio communication application built using Golang. It enables two users to communicate live in a high-fidelity audio conversation with very low latency for musical jam session. The project leverages the Opus codec for audio encoding and decoding, and efficiently uses Go's features like channels and Goroutines.

This project is developed as my final project in the Electrical and Computer Engineering BSc program at Ben-Gurion University.

## Features

- **High-Fidelity Audio**: Utilizes the Opus codec to ensure high-quality audio communication.
- **Low Latency**: Designed to minimize latency for real-time musical jam sessions.
- **Efficient Golang Implementation**: Makes use of Go's concurrency features like channels and Goroutines for efficient audio processing and streaming.


## Installation

### Prerequisites

- Linux operating system for both server and client
- Go 1.15 or higher
- Opus library
- ALSA (for Linux users)
- PortAudio

### Steps

1. Clone the repository:

    ```sh
    git clone https://github.com/YourUsername/RemoteStudioLive.git
    cd RemoteStudioLive
    ```

2. Install dependencies:

    ```sh
    go get -u github.com/gordonklaus/portaudio
    go get -u github.com/hraban/opus
    ```

3. Build the project:

    ```sh
    go build -o remotestudiolive main.go
    ```

## Usage

To start the application, run the following commands:
1. In /RemoteStudioLive/LocalServer, run the server with:

```sh
./run_server

The execution will return a string in the form of: Listening udp on: <server ip>:<server port>

2. In /RemoteStudioLive/LocalServer/Client, run the client with:


```sh
./run_client <server ip>


