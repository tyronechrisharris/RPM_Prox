#!/bin/bash

# Configuration Variables
SIM_IP="100.79.78.21"
TOTAL_LANES=50
RPM_START_PORT=10001
CAM_START_PORT=8554
PI_LISTEN_START_PORT=9001

echo "Generating mediamtx.yml..."
# Initialize the MediaMTX config with the global parameters
cat <<EOF > mediamtx.yml
api: yes
apiAddress: :9997
rtmp: no
hls: no
webrtc: no
srt: no
paths:
EOF

echo "Generating docker-compose.yml..."
# Initialize the Docker Compose config with the MediaMTX service
cat <<EOF > docker-compose.yml
services:
  mediamtx:
    image: bluenviron/mediamtx:1.8.0-ffmpeg
    network_mode: "host"
    restart: unless-stopped
    volumes:
      - ./mediamtx.yml:/mediamtx.yml
EOF

# Loop through all 50 lanes and append the configurations
for i in $(seq 1 $TOTAL_LANES); do
    # Calculate ports based on the current iteration (offset by 1)
    OFFSET=$((i - 1))
    RPM_PORT=$((RPM_START_PORT + OFFSET))
    CAM_PORT=$((CAM_START_PORT + OFFSET))
    LISTEN_PORT=$((PI_LISTEN_START_PORT + OFFSET))
    
    # Format lane number with a leading zero (e.g., 01, 02... 50)
    LANE_NUM=$(printf "%02d" $i)

    # Append the camera path to mediamtx.yml
    cat <<EOF >> mediamtx.yml
  lane${LANE_NUM}_cam:
    source: rtsp://${SIM_IP}:${CAM_PORT}/
    sourceOnDemand: yes
EOF

    # Append the TCP multiplexer service to docker-compose.yml
    cat <<EOF >> docker-compose.yml
  rpm-proxy-${LANE_NUM}:
    build: 
      context: .
      dockerfile: Dockerfile.multiplexer
    network_mode: "host"
    restart: unless-stopped
    command: ["${LISTEN_PORT}", "${SIM_IP}", "${RPM_PORT}"]
EOF
done

echo "Success! The 50-lane configuration files are ready."
