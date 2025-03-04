# Audio Player in Go

This is a lightweight audio player built in Go that utilizes the `faiface` package for audio decoding and playback, along with the `tcell` package for handling terminal-based user interaction.

## Features
- Decodes and plays audio using the `faiface` package
- Provides a terminal-based interface with `tcell`
- Supports common audio formats, mp3 and wav.
- Minimal dependencies for a lightweight experience

## Installation
1. Ensure you have [Go](https://go.dev/dl/) installed on your system.
2. Clone this repository:
   ```sh
   git clone <repository-url>
   cd <repository-name>
   ```
3. Install dependencies:
   ```sh
   go get github.com/faiface/beep
   go get github.com/gdamore/tcell/v2
   ```
4. Build and run the project:
   ```sh
   go run main.go
   ```

## Usage
- Run the player in the terminal and use the provided key bindings to control playback.

## Dependencies
- [`faiface/beep`](https://github.com/faiface/beep) - Audio processing and playback
- [`tcell`](https://github.com/gdamore/tcell) - For handling Key events and input from user

## Contributing
Feel free to open issues or submit pull requests to improve the project.

