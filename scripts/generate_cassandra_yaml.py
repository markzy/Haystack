import argparse

import ruamel.yaml as YAML

parser = argparse.ArgumentParser(prog='cassandra yaml processor')
parser.add_argument("--host", help="current unix machine host")
parser.add_argument("--cassandra_home", help="home directory of cassandra")
parser.add_argument("--output", help="home directory of cassandra")
parser.add_argument("--seed", help="home directory of cassandra")

args = parser.parse_args()

file_path = args.cassandra_home + '/conf/cassandra.yaml'

stream = open(file_path, 'r')
code = YAML.load(stream, Loader=YAML.RoundTripLoader, preserve_quotes=True)
stream.close()

code['cluster_name'] = 'HayStackStoreCluster'
code['data_file_directories'] = [args.cassandra_home + '/data/' + args.host]
code['commitlog_directory'] = args.cassandra_home + '/commitlog/' + args.host
code['cdc_raw_directory'] = args.cassandra_home + '/cdc_raw/' + args.host
code['saved_caches_directory'] = args.cassandra_home + '/saved_caches/' + args.host
code['seed_provider'][0]['parameters'][0]['seeds'] = args.seed
code['storage_port'] = 25536
code['ssl_storage_port'] = 25537
code['listen_address'] = args.host
code['native_transport_port'] = 25538
code['rpc_address'] = args.host

with open(args.output, 'w') as out:
    YAML.dump(code, out, Dumper=YAML.RoundTripDumper)

