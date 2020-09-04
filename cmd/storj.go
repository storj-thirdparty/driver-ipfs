package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"

	"storj.io/uplink"
)

// ConfigStorj depicts keys to search for within the stroj_config.json file.
type ConfigStorj struct {
	Key                  string `json:"key"`
	APIKey               string `json:"apikey"`
	Satellite            string `json:"satellite"`
	Bucket               string `json:"bucket"`
	UploadPath           string `json:"uploadPath"`
	EncryptionPassphrase string `json:"encryptionpassphrase"`
	SerializedAccess     string `json:"serializedAccess"`
	AllowDownload        string `json:"allowDownload"`
	AllowUpload          string `json:"allowUpload"`
	AllowList            string `json:"allowList"`
	AllowDelete          string `json:"allowDelete"`
	NotBefore            string `json:"notBefore"`
	NotAfter             string `json:"notAfter"`
}

// DownloadConfigStorj structure to store data from json file
type DownloadConfigStorj struct {
	Hash         string `json:"hash"`
	DownloadPath string `json:"downloadPath"`
	Key          string `json:"key"`
}

// LoadStorjConfiguration reads and parses the JSON file that contain Storj configuration information.
func LoadStorjConfiguration(fullFileName string) ConfigStorj {

	var configStorj ConfigStorj
	fileHandle, err := os.Open(filepath.Clean(fullFileName))
	if err != nil {
		log.Fatal("Could not load storj config file: ", err)
	}

	jsonParser := json.NewDecoder(fileHandle)
	if err = jsonParser.Decode(&configStorj); err != nil {
		log.Fatal(err)
	}

	// Close the file handle after reading from it.
	if err = fileHandle.Close(); err != nil {
		log.Fatal(err)
	}

	// Display storj configuration read from file.
	fmt.Println("\nRead Storj configuration from the ", fullFileName, " file")
	fmt.Println("API Key\t\t: ", configStorj.APIKey)
	fmt.Println("Satellite	: ", configStorj.Satellite)
	fmt.Println("Bucket		: ", configStorj.Bucket)

	// Convert the upload path to standard form.
	if configStorj.UploadPath != "" {
		if configStorj.UploadPath == "/" {
			configStorj.UploadPath = ""
		} else {
			checkSlash := configStorj.UploadPath[len(configStorj.UploadPath)-1:]
			if checkSlash != "/" {
				configStorj.UploadPath = configStorj.UploadPath + "/"
			}
		}
	}

	fmt.Println("Upload Path\t: ", configStorj.UploadPath)
	fmt.Println("Serialized Access Key\t: ", configStorj.SerializedAccess)
	return configStorj
}

// LoadStorjDownloadConfiguration reads and parses the JSON file that contain Storj configuration information.
func LoadStorjDownloadConfiguration(fullFileName string) DownloadConfigStorj { // fullFileName for fetching storj V3 credentials from  given JSON filename.

	var downloadConfigStorj DownloadConfigStorj

	fileHandle, err := os.Open(filepath.Clean(fullFileName))
	if err != nil {
		log.Fatal("Error in Opening file")
	}

	jsonParser := json.NewDecoder(fileHandle)
	if err = jsonParser.Decode(&downloadConfigStorj); err != nil {
		log.Fatal(err)
	}

	err = fileHandle.Close()
	if err != nil {
		log.Fatal(err)
	}

	// Display read information.
	fmt.Println("\nReading Download configuration from file: ", fullFileName)
	fmt.Println("Hash\t\t\t: ", downloadConfigStorj.Hash)
	fmt.Println("Download Path\t\t: ", downloadConfigStorj.DownloadPath)

	return downloadConfigStorj
}

// ShareAccess generates and prints the shareable serialized access
// as per the restrictions provided by the user.
func ShareAccess(access *uplink.Access, configStorj ConfigStorj) {

	allowDownload, _ := strconv.ParseBool(configStorj.AllowDownload)
	allowUpload, _ := strconv.ParseBool(configStorj.AllowUpload)
	allowList, _ := strconv.ParseBool(configStorj.AllowList)
	allowDelete, _ := strconv.ParseBool(configStorj.AllowDelete)
	notBefore, _ := time.Parse("2006-01-02_15:04:05", configStorj.NotBefore)
	notAfter, _ := time.Parse("2006-01-02_15:04:05", configStorj.NotAfter)

	permission := uplink.Permission{
		AllowDownload: allowDownload,
		AllowUpload:   allowUpload,
		AllowList:     allowList,
		AllowDelete:   allowDelete,
		NotBefore:     notBefore,
		NotAfter:      notAfter,
	}

	// Create shared access.
	sharedAccess, err := access.Share(permission)
	if err != nil {
		log.Fatal("Could not generate shared access: ", err)
	}

	// Generate restricted serialized access.
	serializedAccess, err := sharedAccess.Serialize()
	if err != nil {
		log.Fatal("Could not serialize shared access: ", err)
	}
	fmt.Println("Shareable serialized access: ", serializedAccess)
}

// ConnectToStorj reads Storj configuration from given file
// and connects to the desired Storj network.
// It then reads data property from an external file.
func ConnectToStorj(fullFileName string, configStorj ConfigStorj, accesskey bool) (*uplink.Access, *uplink.Project) {

	var access *uplink.Access
	var cfg uplink.Config

	// Configure the UserAgent.
	cfg.UserAgent = "IPFS"
	ctx := context.Background()
	var err error

	if accesskey {
		fmt.Println("\nConnecting to Storj network using Serialized access.")
		// Generate access handle using serialized access.
		access, err = uplink.ParseAccess(configStorj.SerializedAccess)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		fmt.Println("\nConnecting to Storj network.")
		// Generate access handle using API key, satellite url and encryption passphrase.
		access, err = cfg.RequestAccessWithPassphrase(ctx, configStorj.Satellite, configStorj.APIKey, configStorj.EncryptionPassphrase)
		if err != nil {
			log.Fatal(err)
		}
	}

	// Open a new porject.
	project, err := cfg.OpenProject(ctx, access)
	if err != nil {
		log.Fatal(err)
	}
	defer project.Close()

	// Ensure the desired Bucket within the Project.
	_, err = project.EnsureBucket(ctx, configStorj.Bucket)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Successfully connected to Storj network.")
	return access, project
}

// UploadData uploads the backup file to storj network.
func UploadData(project *uplink.Project, configStorj ConfigStorj, uploadFileName string, reader io.Reader) {

	ctx := context.Background()

	// Create an upload handle.
	upload, err := project.UploadObject(ctx, configStorj.Bucket, configStorj.UploadPath+uploadFileName, nil)
	if err != nil {
		log.Fatal("Could not initiate upload : ", err)
	}
	fmt.Printf("\nUploading %s to %s.", configStorj.UploadPath+uploadFileName, configStorj.Bucket)

	// Upload data on storj.
	_, err = io.Copy(upload, reader)
	if err != nil {
		abortErr := upload.Abort()
		log.Fatal("Could not upload data to storj: ", err, abortErr)
	}
	// Commit the upload after copying the complete content of the backup file to upload object.
	fmt.Println("\nPlease wait while the upload is being committed to Storj.")
	err = upload.Commit()
	if err != nil {
		log.Fatal("Could not commit object upload : ", err)
	}
}

// DownloadData function downloads the data from storj bucket after upload to verify data is uploaded successfully.
func DownloadData(project *uplink.Project, downloadConfigStorj DownloadConfigStorj, readFile *bytes.Reader) {

	makebuffer := make([]byte, 46)

	// Read data from IPFS
	readHashData, err := readFile.Read(makebuffer)
	if err != nil {
		fmt.Println("Error reading file:", err)
	}

	// Seperate the Hash and configration data
	data := []byte(makebuffer[0:readHashData])
	downloadFileName := string(data)
	downloadFileSize := readFile.Size()

	restDataBuf := make([]byte, downloadFileSize-46)
	readRestFile, err := readFile.Read(restDataBuf)
	if err != nil {
		fmt.Println("Error reading file:", err)
	}

	dataEnc := []byte(restDataBuf[0:readRestFile])
	pkey := []byte(downloadConfigStorj.Key)

	// Decrypt the configration data
	decryptData, err := decrypt(pkey, dataEnc)
	if err != nil {
		log.Fatal(err)
	}

	// Split the configration data
	splitStorjData := strings.Split(string(decryptData), ",")
	downloadBucket := splitStorjData[0]
	downloadPath := splitStorjData[1]
	lastFileName := splitStorjData[2]

	ctx := context.Background()
	var dataDownload []byte
	var lastIndex int64
	var buf = make([]byte, 32768)
	lastIndex = 0

	fmt.Printf("Downloading %s...\n", downloadFileName)

	// Loop to read the object in chunks and store the read data in a byte array.
	download, err3 := project.DownloadObject(ctx, downloadBucket, downloadPath+downloadFileName+"/"+downloadFileName+".txt", &uplink.DownloadOptions{Offset: lastIndex, Length: int64(cap(buf))})
	if err3 != nil {
		fmt.Println("Error: ", err3)
	}

	var err2 error
	dataDownload, err2 = ioutil.ReadAll(download)
	if err2 != nil {
		log.Fatal(err)
	}

	//Convert byte array into String
	receiveContentsMeta := string(dataDownload)

	receiveContentsMeta = strings.TrimSuffix(receiveContentsMeta, ",")

	downloadFileNamesDEBUG := strings.Split(receiveContentsMeta, ",")

	var fileNameDownload = downloadConfigStorj.DownloadPath + "/" + lastFileName

	_, err = os.Stat(fileNameDownload)
	if err == nil {
		err = os.Remove(fileNameDownload)
		if err != nil {
			log.Fatal(err)
		}
	}

	hmkey := []byte("This is a storj ipfs private key")

	for _, filename := range downloadFileNamesDEBUG {
		downloadObj, err := project.DownloadObject(ctx, downloadBucket, downloadPath+downloadFileName+"/"+filename, &uplink.DownloadOptions{Offset: lastIndex, Length: int64(cap(buf))})
		if err != nil {
			log.Fatalf("Could not open object at %q: %v", downloadPath+downloadFileName+"/"+filename, err)
		}
		if _, err = os.Stat("./debug"); os.IsNotExist(err) {
			err1 := os.Mkdir("./debug", 0750)
			if err1 != nil {
				log.Fatal("Could not create debug folder: ", err1)
			}
		}

		// Read everything from the stream.
		receivedContents, err := ioutil.ReadAll(downloadObj)
		if err != nil {
			log.Fatal("Could not Read All content in stream:", err)
		}

		//Decryt the downloaded file data from storj
		dec, err := decrypt(hmkey, receivedContents)
		if err != nil {
			log.Fatal("Could not decrypt received data:", err)
		}
		// Create/open file in append mode.
		downloadFileDisk, err := os.OpenFile(filepath.Clean(filepath.Join("./debug", lastFileName)), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
		if err != nil {
			log.Fatal(err)
		}

		// read data
		dataReader := bytes.NewReader(dec)
		_, err = io.Copy(downloadFileDisk, dataReader)
		if err != nil {
			log.Fatal(err)
		}

		// Close the file handle after reading from it.
		if err = downloadFileDisk.Close(); err != nil {
			log.Fatal(err)
		}
	}
	fmt.Printf("File downloading: Complete!\n")
	fmt.Printf("\n file \"%s\" downloaded to \"%s\"\n", lastFileName, "debug/")
}

// Function to decrypt data based on given key.
func decrypt(key, text []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	if len(text) < aes.BlockSize {
		return nil, errors.New("ciphertext too short")
	}
	iv := text[:aes.BlockSize]
	text = text[aes.BlockSize:]
	cfb := cipher.NewCFBDecrypter(block, iv)
	cfb.XORKeyStream(text, text)
	data, err := base64.StdEncoding.DecodeString(string(text))
	if err != nil {
		return nil, err
	}
	//nolint:ineffassign
	iv = nil
	text = nil
	cfb = nil
	return data, nil
}
