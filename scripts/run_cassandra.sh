# /bin/bash

python3.4 generate_cassandra_yaml.py --host $(hostname) --cassandra_home $(pwd)/cassandra --seed unix4.andrew.cmu.edu --output ./cassandra.$(hostname).yaml
./cassandra/bin/cassandra -D cassandra.config=file:$(pwd)/cassandra.$(hostname).yaml