package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-redis/redis"
	"github.com/gocql/gocql"
)

var cass *gocql.ClusterConfig
var rclient *redis.Client
var cluster_addresses = [...]string{"unix4.andrew.cmu.edu", "unix5.andrew.cmu.edu", "unix6.andrew.cmu.edu", "unix7.andrew.cmu.edu", "unix8.andrew.cmu.edu"}

// const CASSPORT = 7269
// const CASSPORT = 9161
const REDISPORT = 6969
const CASSPORT = 9040
const PORT = 4000
const keyspace = "store"
const table = "photos"
const key = "photoid"
const value = "data"

func handler(w http.ResponseWriter, r *http.Request) {
	//GET

	fmt.Println(strings.Split(r.URL.Path[1:], "/"))
	if r.Method == "GET" && r.URL.Path[1:] != "favicon.ico" {
		fmt.Println("GET METHOD")
		pathInfo := strings.Split(r.URL.Path[1:], "/")
		if len(pathInfo) != 1 {
			fmt.Fprintf(w, "Incorrect URL format")
		} else {
			//GET FROM CACHE??
			pid := pathInfo[0]
			val, err := rclient.Get(pid).Result()
			if err == nil && val != "" {
				fmt.Println("Got from cache")
				w.Write([]byte(val))
				// fmt.Fprintf(w, val)
			} else {
				//FINE GET FROM Store
				session, err := cass.CreateSession()
				if err != nil {
					fmt.Fprintf(w, "Failed Creating Cassandra Session")
				}
				defer session.Close()
				var data string
				err = session.Query("SELECT " + value + " FROM photos WHERE " + key + "=" + "'" + pid + "'").Scan(&data)
				if err != nil {
					fmt.Fprintf(w, "Failed reading from Cassandra")
					fmt.Println(err)
				} else {
					w.Write([]byte(data))
					err = rclient.Set(pid, string(data), 0).Err()
					if err != nil {
						fmt.Println("Failed writing to cache")
						fmt.Println(err)
					}
				}
			}
		}
	} else if r.Method == "POST" && r.URL.Path[1:] != "favicon.ico" {

		fmt.Println("POST METHOD")
		pathInfo := strings.Split(r.URL.Path[1:], "/")
		if len(pathInfo) != 1 {
			fmt.Fprintf(w, "Incorrect URL format")
		} else {
			session, err := cass.CreateSession()
			if err != nil {
				fmt.Fprintf(w, "Failed Creating Cassandra Session")
			}
			defer session.Close()
			//POST TO STORE
			pid := pathInfo[0]
			d, err := ioutil.ReadAll(r.Body)
			if err != nil {
				fmt.Fprintf(w, "Couldn't Read Post request")
				fmt.Println(err)
			} else {
				err = session.Query("INSERT INTO "+table+" ("+key+","+value+") VALUES(?,?)", pid, d).Exec()
				if err != nil {
					fmt.Fprintf(w, "Failed to write to Cassandra")
					fmt.Println(err)
				} else {
					fmt.Fprintf(w, "Write Successful")
					err = rclient.Set(pid, string(d), 0).Err()
					if err != nil {
						fmt.Println("Failed writing to cache")
						fmt.Println(err)
					}
				}
			}
		}

	} else {
		fmt.Fprintf(w, "Not a valid request")
	}

}

func main() {
	cass = gocql.NewCluster(cluster_addresses[0] + ":" + strconv.Itoa(CASSPORT))
	cass.Keyspace = keyspace
	cass.Timeout = 5 * time.Second
	cass.ProtoVersion = 4
	rclient = redis.NewClient(&redis.Options{Addr: cluster_addresses[0] + ":" + strconv.Itoa(REDISPORT), Password: "", DB: 0})
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
		err = session.Query("CREATE TABLE " + table + " (" + key + " text PRIMARY KEY," + value + " blob);").Exec()
		if err != nil {
			fmt.Println("Error creating table")
			log.Fatal(err)
		}
	}
	session.Close()
	http.HandleFunc("/", handler)
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(PORT), nil))

}
