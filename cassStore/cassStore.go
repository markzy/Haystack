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
var cluster_addresses = [...]string{"unix4.andrew.cmu.edu:", "unix5.andrew.cmu.edu:"}

// const CASSPORT = 7269
// const CASSPORT = 9161
const REDISPORT = "6969"
const CASSPORT = "25538"
const PORT = 4000
const keyspace = "store"
const table = "photos"
const key = "photoid"
const value = "data"
const metadata = "status"

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
				var meta string
				var nouse string
				err = session.Query("SELECT * "+" FROM photos WHERE "+key+"="+"'"+pid+"'").Scan(&nouse, &data, &meta)
				if err != nil {
					fmt.Fprintf(w, "Failed reading from Cassandra")
					fmt.Println(err)
				} else {
					if meta != "0" {
						w.Write([]byte(data))
						err = rclient.Set(pid, string(data), 0).Err()
						if err != nil {
							fmt.Println("Failed writing to cache")
							fmt.Println(err)
						}
					} else {
						fmt.Fprintf(w, "File Deleted")
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
				err = session.Query("INSERT INTO "+table+" ("+key+","+value+","+metadata+") VALUES(?,?,?)", pid, d, "1").Exec()
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

	} else if r.Method == "DELETE" && r.URL.Path[1:] != "favicon.ico" {
		fmt.Println("POST METHOD")
		pathInfo := strings.Split(r.URL.Path[1:], "/")
		pid := pathInfo[0]
		if len(pathInfo) != 1 {
			fmt.Fprintf(w, "Incorrect URL format")
		} else {
			session, err := cass.CreateSession()
			if err != nil {
				fmt.Fprintf(w, "Failed Creating Cassandra Session")
			}
			defer session.Close()
			err = session.Query("UPDATE " + table + " SET " + metadata + "='0' WHERE " + key + "='" + pid + "'").Exec()
			if err != nil {
				fmt.Fprintf(w, "Failed to write to Cassandra")
				fmt.Println(err)
			} else {
				fmt.Fprintf(w, "Delete Successful")
				err = rclient.Del(pid).Err()
				if err != nil {
					fmt.Println("Failed deleting to cache")
					fmt.Println(err)
				}
			}

		}
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

	err = session.Query(`DROP KEYSPACE IF EXISTS ` + keyspace).Exec()
	if err != nil {
		fmt.Println(err)
	}

	err = session.Query(fmt.Sprintf(`CREATE KEYSPACE %s
	WITH replication = {
		'class' : 'SimpleStrategy',
		'replication_factor' : %d
	}`, keyspace, 2)).Exec()

	if err != nil {
		fmt.Println(err)
	}
}

func main() {
	cass = gocql.NewCluster(cluster_addresses[0]+CASSPORT, cluster_addresses[1]+CASSPORT)
	cass.Keyspace = keyspace
	cass.Timeout = 5 * time.Second
	cass.ProtoVersion = 4
	i, _ := strconv.Atoi(CASSPORT)
	cass.Port = i
	createKeyspace(cass, "store")

	rclient = redis.NewClient(&redis.Options{Addr: cluster_addresses[0] + REDISPORT, Password: "", DB: 0})
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
		err = session.Query("CREATE TABLE " + table + " (" + key + " text PRIMARY KEY," + value + " blob," + metadata + " text);").Exec()
		if err != nil {
			fmt.Println("Error creating table")
			fmt.Println(err)
			// log.Fatal(err)
		}
	}
	session.Close()
	http.HandleFunc("/", handler)
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(PORT), nil))

}
