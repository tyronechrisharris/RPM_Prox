import yaml

SIMULATOR_IP = "100.79.78.21"
TOTAL_LANES = 50

# Starting Ports
SIM_RPM_START_PORT = 10001
SIM_CAM_START_PORT = 8554
PI_LISTEN_START_PORT = 9001

def generate_mediamtx():
    config = {
        "api": True,
        "apiAddress": ":9997",
        "rtmp": False,
        "hls": False,
        "webrtc": False,
        "srt": False,
        "paths": {}
    }

    for i in range(TOTAL_LANES):
        lane_num = i + 1
        cam_port = SIM_CAM_START_PORT + i
        path_name = f"lane{lane_num:02d}_cam"
        
        config["paths"][path_name] = {
            "source": f"rtsp://{SIMULATOR_IP}:{cam_port}/",
            "sourceOnDemand": True
        }

    with open("mediamtx.yml", "w") as f:
        yaml.dump(config, f, default_flow_style=False, sort_keys=False)
    print("Generated mediamtx.yml")

def generate_docker_compose():
    compose = {
        "services": {
            "mediamtx": {
                "image": "bluenviron/mediamtx:1.8.0-ffmpeg",
                "network_mode": "host",
                "restart": "unless-stopped",
                "volumes": ["./mediamtx.yml:/mediamtx.yml"]
            }
        }
    }

    for i in range(TOTAL_LANES):
        lane_num = i + 1
        sim_rpm_port = SIM_RPM_START_PORT + i
        pi_listen_port = PI_LISTEN_START_PORT + i
        service_name = f"rpm-proxy-{lane_num:02d}"

        compose["services"][service_name] = {
            "build": {
                "context": ".",
                "dockerfile": "Dockerfile.multiplexer"
            },
            "network_mode": "host",
            "restart": "unless-stopped",
            "command": [str(pi_listen_port), SIMULATOR_IP, str(sim_rpm_port)]
        }

    with open("docker-compose.yml", "w") as f:
        yaml.dump(compose, f, default_flow_style=False, sort_keys=False)
    print("Generated docker-compose.yml")

if __name__ == "__main__":
    generate_mediamtx()
    generate_docker_compose()
