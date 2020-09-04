package cmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"

	shell "github.com/ipfs/go-ipfs-api"
)

// ConfigIpfs defines the variables and types.
type ConfigIpfs struct {
	HostName  string `json:"hostName"`
	Port      string `json:"port"`
	Path      string `json:"path"`
	ChunkSize string `json:"chunkSize"`
}

// LoadIpfsProperty reads and parses the JSON file
// that contain a IPFS instance's property.
// and returns all the properties as an object.
func LoadIpfsProperty(fullFileName string) ConfigIpfs {
	var configIpfs ConfigIpfs

	// Open and read the file
	fileHandle, err := os.Open(filepath.Clean(fullFileName))
	if err != nil {
		log.Fatal("Can't open the file : ", err)
	}

	jsonParser := json.NewDecoder(fileHandle)
	if err = jsonParser.Decode(&configIpfs); err != nil {
		log.Fatal(err)
	}

	err = fileHandle.Close()
	if err != nil {
		log.Fatal(err)
	}

	// Display read information.
	fmt.Println("\nReading IPFS configuration from file: ", fullFileName, "file")
	fmt.Println("Host Name\t: ", configIpfs.HostName)
	fmt.Println("Port\t\t: ", configIpfs.Port)
	fmt.Println("Upload File Path: ", configIpfs.Path)

	return configIpfs
}

// ConnectToIpfs will connect to a IPFS instance,
// based on the read property from an external file.
// It returns a reference to an io.Reader with IPFS instance information.
func ConnectToIpfs(configIpfs ConfigIpfs) *shell.Shell {

	fmt.Println("\nConnecting to IPFS...")

	if configIpfs.HostName == "ipfsHostName" || configIpfs.HostName == "" {
		err1 := errors.New("Invalid HostName")
		log.Fatal("Invalid Hostname error : ", err1)
	}
	// Connect IPFS deamon to IPFS node.
	sh := shell.NewShell(configIpfs.HostName + ":" + configIpfs.Port)

	_, _, errVer := sh.Version()
	if errVer != nil {
		err1 := errors.New("Could not find Daemon running")
		log.Fatal("Daemon error : ", err1)
	}

	// Convert size of chunks into int64.
	givenSize, _ := strconv.ParseInt(configIpfs.ChunkSize, 10, 64)

	if givenSize <= 0 {
		err1 := errors.New("Invalid chunk size entered")
		log.Fatal("Invalid Chunk size : ", err1)
	}

	fmt.Println("Successfully connected to IPFS!")

	return sh
}

// CreateCID will connect to a IPFS instance
// based on the read property from an external file.
// It returns Created CID.
func CreateCID(sh *shell.Shell, data []byte) string {

	readers := bytes.NewReader(data)

	// Create encrypt chunk CID
	encryptChunkCID, err := sh.Add(readers, shell.OnlyHash(true))
	if err != nil {
		log.Fatal("Error in Encryption : ", err)
	}

	// Return IPFS connection object, chunk size and file path.
	return encryptChunkCID
}

// ConnectToIPFSForDownload will connect to a IPFS instance,
// based on the hash name of file on IPFS.
// It returns a reference to an io.Reader with IPFS instance information
func ConnectToIPFSForDownload(hash string, hostName string, port string) *bytes.Reader {

	fmt.Println("\nConnecting to IPFS...")
	if hostName == "ipfsHostName" || hostName == "" {
		err1 := errors.New("Invalid HostName")
		log.Fatal(err1)
	}

	if hash[0:2] != "Qm" || len(hash) != 46 {
		err1 := errors.New("Invalid Shareable Hash")
		log.Fatal(err1)
	}

	// Connect to IPFS daemon to IPFS node.
	sh := shell.NewShell(hostName + ":" + port)
	_, _, errVer := sh.Version()
	if errVer != nil {
		err1 := errors.New("Could not find Daemon running")
		log.Fatal(err1)
	}

	// Inform about successful connection.
	fmt.Println("\nSuccessfully connected to IPFS!")

	// Get data from ipfs node.
	fileReader, err := sh.Cat(hash)
	if err != nil {
		log.Fatal("IPFS data read error: ", err)
	}

	// Read all data recive from ipfs.
	readbytes, _ := ioutil.ReadAll(fileReader)
	reader := bytes.NewReader(readbytes)
	return reader
}

// GetReader returns a Reader of corresponding file whose path is specified.
// io.ReadCloser type of object returned is used to perform transfer of file to Storj.
func GetReader(sh *shell.Shell, configIpfs ConfigIpfs) io.ReadCloser {
	ipfsReader, err1 := os.Open(filepath.Clean(configIpfs.Path))
	if err1 != nil {
		err2 := errors.New("Invalid File path entered")
		log.Fatal("FIle path error : ", err2)
	}
	return ipfsReader
}

// GetReaderDownload returns a Reader of corresponding file whose path is specified.
// io.ReadCloser type of object returned is used to perform transfer of file to Storj.
func GetReaderDownload(sh *shell.Shell, hash string) *bytes.Reader {
	// Get data from ipfs node.
	fileReader, err := sh.Cat(hash)
	if err != nil {
		fmt.Println("IPFS data read error: ", err)
	}

	// // Read all data recive from ipfs.
	readbytes, _ := ioutil.ReadAll(fileReader)
	reader := bytes.NewReader(readbytes)
	return reader
}
