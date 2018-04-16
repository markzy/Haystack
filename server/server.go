package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"
	"github.com/go-redis/redis"
	"github.com/gocql/gocql"
	"github.com/gorilla/mux"
)

var cluster_addresses = [...]string{"unix4.andrew.cmu.edu", "unix5.andrew.cmu.edu"}

var cass *gocql.ClusterConfig
var rclient *redis.Client

const REDISPORT = "25540"
const CASSPORT = "25538"
const PORT = 25555
const keyspace = "store"
const table = "photos"
const key = "photoid"
const value = "data"
const metadata = "status"

func deleteHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	pid := params["id"]

	if pid == "favicon.ico" {
		return
	}

	fmt.Println("DELETE METHOD")

	session, err := cass.CreateSession()
	if err != nil {
		fmt.Fprintf(w, "Failed Creating Cassandra Session")
		return
	}
	defer session.Close()

	err = session.Query("UPDATE " + table + " SET " + metadata + "='0' WHERE " + key + "='" + pid + "'").Exec()

	if err != nil {
		fmt.Fprintf(w, "Failed to write to Cassandra")
		fmt.Println(err)
		return
	}

	fmt.Fprintf(w, "delete Successful")
	err = rclient.Del(pid).Err()
	if err != nil {
		fmt.Println("Failed deleting cache")
		fmt.Println(err)
	}
}

func postHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path[1:] == "favicon.ico" {
		return
	}

	fmt.Println("POST METHOD")

	session, err := cass.CreateSession()
	if err != nil {
		fmt.Fprintf(w, "Failed Creating Cassandra Session")
		return
	}

	defer session.Close()
	//POST TO STORE
	pid := generateUniqueID()
	d, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Fprintf(w, "Couldn't Read Post request")
		fmt.Println(err)
		return
	}

	err = session.Query("INSERT INTO "+table+" ("+key+","+value+","+"status) VALUES(?,?,?)", pid, d, "1").Exec()
	if err != nil {
		fmt.Fprintf(w, "Failed to write to Cassandra")
		fmt.Println(err)
		return
	}

	fmt.Fprintf(w, "Write Successful, use this photo id to access photo:"+pid)
	err = rclient.Set(pid, string(d), 0).Err()
	if err != nil {
		fmt.Println("Failed writing to cache")
		fmt.Println(err)
	}
}

func getHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	pid := params["id"]

	if pid == "favicon.ico" {
		return
	}

	fmt.Println("GET METHOD")
	val, err := rclient.Get(pid).Result()
	if err == nil && val != "" {
		fmt.Println("Got from cache")
		w.Write([]byte(val))
		return
		// fmt.Fprintf(w, val)
	}

	//GET FROM Store
	session, err := cass.CreateSession()
	if err != nil {
		fmt.Fprintf(w, "Failed Creating Cassandra Session")
		return
	}

	defer session.Close()
	var id string
	var data string
	var status string
	err = session.Query("SELECT * " + " FROM photos WHERE " + key + "=" + "'" + pid + "'").Scan(&id, &data, &status)
	if err != nil {
		fmt.Fprintf(w, "Failed reading from Cassandra")
		fmt.Println(err)
		return
	}

	if status == "0" {
		fmt.Fprintf(w, "File Deleted")
		return
	}

	w.Write([]byte(data))
	err = rclient.Set(pid, string(data), 0).Err()
	if err != nil {
		fmt.Println("Failed writing to cache")
		fmt.Println(err)
	}
}

func createKeyspace(cluster *gocql.ClusterConfig, keyspace string) {
	c := *cluster
	c.Keyspace = "system"
	c.Timeout = 20 * time.Second
	session, err := c.CreateSession()
	if err != nil {
		fmt.Println("createSession:", err)
	}

	err = session.Query(fmt.Sprintf(`CREATE KEYSPACE IF NOT EXISTS %s
	WITH replication = {
		'class' : 'SimpleStrategy',
		'replication_factor' : %d
	}`, keyspace, 2)).Exec()

	if err != nil {
		fmt.Println(err)
	}
}

func getTimestamp() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

func generateUniqueID() string {
	// TODO: need to involve host id and sequence number here
	return strconv.FormatInt(getTimestamp(), 10)
}

func init() {
	cass = gocql.NewCluster(cluster_addresses[0]+":"+CASSPORT, cluster_addresses[1]+":"+CASSPORT)
	cass.Keyspace = keyspace
	cass.Timeout = 5 * time.Second
	cass.ProtoVersion = 4
	i, _ := strconv.Atoi(CASSPORT)
	cass.Port = i
	createKeyspace(cass, "store")

	rclient = redis.NewClient(&redis.Options{Addr: "unix4.andrew.cmu.edu:" + REDISPORT, Password: "", DB: 0})
	_, err := rclient.Ping().Result()
	if err != nil {
		fmt.Println("Can't ping cache")
		fmt.Println(err)
		// log.Fatal(err)
	}
	session, err := cass.CreateSession()
	if err != nil {
		fmt.Println("Failed Creating Cassandra Session")
		log.Fatal(err)
	}
	keyspacemeta, err := session.KeyspaceMetadata(keyspace)
	if err != nil {
		fmt.Println("Keyspace Error")
		log.Fatal(err)
	}
	_, exists := keyspacemeta.Tables[table]
	if exists != true {
		err = session.Query("CREATE TABLE IF NOT EXISTS " + table + " (" + key + " text PRIMARY KEY," + value + " blob, " + metadata + " text);").Exec()
		if err != nil {
			fmt.Println("Error creating table")
			log.Fatal(err)
		}
	}
	session.Close()
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/{id}", getHandler).Methods("GET")
	r.HandleFunc("/", postHandler).Methods("POST")
	r.HandleFunc("/{id}", deleteHandler).Methods("DELETE")
	if err := http.ListenAndServe(":"+strconv.Itoa(PORT), r); err != nil {
		log.Fatal(err)
	}
}
