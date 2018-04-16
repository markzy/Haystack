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
	"github.com/gorilla/mux"
)

type Config struct {
	ServerAddresses []string
	SequenceNumber  int
}

type Photo struct {
	PhotoID string
	// 0 deleted, 1 ready, 2 uploading
	State   int
	Content []byte
}

type PhotoDAO struct{}

var photoDAO = PhotoDAO{}
var SystemConfig Config
var cass *gocql.ClusterConfig
var rclient *redis.Client
var cluster_addresses = [...]string{"unix4.andrew.cmu.edu:", "unix5.andrew.cmu.edu:"}

const REDISPORT = "6969"
const CASSPORT = "25538"
const PORT = 4000
const keyspace = "store"
const table = "photos"
const key = "photoid"
const value = "data"

func (m *PhotoDAO) Insert(pm Photo) error {
	return nil
}

func (m *PhotoDAO) FindById(id string) (Photo, error) {
	return Photo{}, nil
}

func (m *PhotoDAO) Update(photo Photo) error {
	return nil
}

func (m *PhotoDAO) Delete(photo Photo) error {
	return nil
}

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

//func getPhotoURL(w http.ResponseWriter, r *http.Request) {
//	params := mux.Vars(r)
//	id := params["id"]
//	result, err := photoDAO.FindById(id)
//
//	if err != nil || result.State == 0 {
//		w.WriteHeader(404)
//		return
//	}
//
//	// get from cassandra
//	//url := "http://" + SystemConfig.ServerAddresses[machineID] + "/" + result.PhotoID
//	http.Redirect(w, r, url, 302)
//}
//
//func uploadPhoto(w http.ResponseWriter, r *http.Request) {
//	id := generateUniqueID()
//	photo := Photo{PhotoID: id, State: 2}
//	photoDAO.Insert(photo)
//
//	body, parseErr := ioutil.ReadAll(r.Body)
//	if parseErr != nil {
//		http.Error(w, parseErr.Error(), http.StatusInternalServerError)
//		return
//	}
//
//	for _, element := range SystemConfig.ServerAddresses {
//		res, err := http.Post("http://"+element+"/"+id, r.Header.Get("Content-type"), bytes.NewReader(body))
//		if err != nil {
//			http.Error(w, err.Error(), http.StatusInternalServerError)
//			return
//		}
//	}
//
//	photo.State = 1
//	updateErr := photoDAO.Update(photo)
//	if updateErr != nil {
//		panic(updateErr)
//	}
//
//	w.Write([]byte(photo.PhotoID))
//}

func getTimestamp() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

func generateUniqueID() string {
	// TODO: need to involve host id and sequence number here
	return strconv.FormatInt(getTimestamp(), 10)
}

func init() {
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
		err = session.Query("CREATE TABLE " + table + " (" + key + " text PRIMARY KEY," + value + " blob);").Exec()
		if err != nil {
			fmt.Println("Error creating table")
			log.Fatal(err)
		}
	}
	session.Close()
}

func main() {
	SystemConfig.ServerAddresses = []string{"localhost:4000"}
	r := mux.NewRouter()
	r.HandleFunc("/:id", photoGetHandler).Methods("GET")
	r.HandleFunc("/", photoPostHandler).Methods("POST")
	r.HandleFunc("/:id", photoDeleteHandler).Methods("DELETE")

	//r.HandleFunc("/", handler)
	//r.HandleFunc("/photo/{id}", getPhotoURL).Methods("GET")
	//r.HandleFunc("/photo", uploadPhoto).Methods("POST")

	if err := http.ListenAndServe(":"+strconv.Itoa(PORT), r); err != nil {
		log.Fatal(err)
	}

}
