# SRLS Dual-Proxy Edge Node

A high-performance, containerized proxy system designed to duplicate 50 lanes of video (RTSP) and Radiation Portal Monitor (RPM) telemetry (TCP) from the Sandia Radiation Portal Monitor Lane Simulator (SRLS). 

This architecture allows for the identical, real-time distribution of 640x480 H.264 camera feeds and live telemetry to multiple monitoring platforms simultaneously (e.g., OSCAR V3.5 and a standard Central Alarm Station) for true comparative testing.

## Architecture

This system is optimized to run on a Raspberry Pi 4 or 5 utilizing an ARM64 architecture, featuring:
* **MediaMTX:** Handles strict RTSP pass-through with zero transcoding to conserve CPU.
* **Custom Go Multiplexer:** A lightweight, compiled TCP duplicator that actively broadcasts incoming RPM data to all connected clients without load-balancing or dropping connections.
* **Host Networking:** Bypasses Docker's default bridged NAT to ensure the Pi's CPU is not overwhelmed by routing thousands of simultaneous packets across 50 simulated lanes.

## Prerequisites

* Raspberry Pi 4 or Raspberry Pi 5 (Gigabit Ethernet required).
* Docker and Docker Compose installed.
* Network routing to the SRLS instance (Default Simulator IP: `100.79.78.21`).

## Repository Contents

* `generate.sh`: A native Bash script to generate the required Docker Compose and MediaMTX configurations for all 50 lanes instantly.
* `main.go`: The source code for the high-performance TCP multiplexer.
* `Dockerfile.multiplexer`: The multi-stage build instructions to compile the Go binary into a zero-overhead `scratch` container.

## Deployment Instructions

1.  **Transfer Files:** Ensure `generate.sh`, `main.go`, and `Dockerfile.multiplexer` are located in the same directory on your Raspberry Pi.
2.  **Make Generator Executable:**
    ```bash
    chmod +x generate.sh
    ```
3.  **Generate Configurations:** Run the script to build `mediamtx.yml` and `docker-compose.yml`.
    ```bash
    ./generate.sh
    ```
4.  **Build and Deploy:** Start the proxy array. This will compile the Go binary for ARM64 and launch all 51 containers (1 for MediaMTX, 50 for the RPM proxies).
    ```bash
    docker-compose up --build -d
    ```

## Connecting Clients (OSCAR / CAS)

Point your monitoring platforms to the **Raspberry Pi's IP Address** (not the simulator's IP). Assuming the Pi is at `192.168.1.50`:

### Video Feeds (RTSP)
* **Lane 01:** `rtsp://192.168.1.50:8554/lane01_cam`
* **Lane 02:** `rtsp://192.168.1.50:8554/lane02_cam`
* ...
* **Lane 50:** `rtsp://192.168.1.50:8554/lane50_cam`

*Note: Streams are configured as `sourceOnDemand`. The proxy will only pull from the SRLS when a client is actively viewing the feed.*

### RPM Telemetry (TCP Sockets)
* **Lane 01 RPM:** `192.168.1.50 : 9001`
* **Lane 02 RPM:** `192.168.1.50 : 9002`
* ...
* **Lane 50 RPM:** `192.168.1.50 : 9050`

You can verify the data fan-out by connecting multiple terminal sessions to the same port using Netcat (e.g., `nc 192.168.1.50 9001`) and observing the duplicated byte streams.

## Maintenance Notes

* **Logging:** Minimal logging is enforced to prevent rapid degradation of the Raspberry Pi's SD card. For sustained production usage, redirect Docker logs to `tmpfs`.
* **Hardware Limits:** Pushing beyond 50 lanes or increasing camera resolutions above 640x480 may oversaturate the network interface or CPU.
