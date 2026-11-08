docker build -t mesh-mqtt:test ../

docker run \
    -e BROKER=localhost:1883 \
    -e TOPIC=meshstatic \
    mesh-mqtt:test
