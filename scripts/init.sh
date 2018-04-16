# /bin/source

# install python dependency
pip3.4 install ruamel.yaml --user

# install and configure cassandra
wget http://apache.claz.org/cassandra/3.11.2/apache-cassandra-3.11.2-bin.tar.gz
tar -xf apache-cassandra-3.11.2-bin.tar.gz
mv apache-cassandra-3.11.2 cassandra
rm apache-cassandra-3.11.2-bin.tar.gz
mv ./cassandra-env.sh ./cassandra/conf/cassandra-env.sh

# install redis
wget http://download.redis.io/releases/redis-4.0.9.tar.gz
tar xzf redis-4.0.9.tar.gz
mv redis-4.0.9 redis
cd redis
make
cd ..
rm redis-4.0.9.tar.gz

# install go
wget https://dl.google.com/go/go1.9.5.linux-amd64.tar.gz
tar xzf go1.9.5.linux-amd64.tar.gz
export GOROOT=$(pwd)/go
export PATH=$PATH:$GOROOT/bin
rm go1.9.5.linux-amd64.tar.gz

mkdir go_projects
cd go_projects
mkdir src
mkdir bin
mkdir pkg
cd ..
export GOPATH=$(pwd)/go_projects
go get github.com/go-redis/redis
go get github.com/gorilla/mux
go get github.com/gocql/gocql
cd $GOPATH/src
git clone https://github.com/markzy/Haystack.git
cd ~/haystack


