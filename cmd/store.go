package cmd

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"

	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	shell "github.com/ipfs/go-ipfs-api"
	chunker "github.com/ipfs/go-ipfs-chunker"
	"github.com/spf13/cobra"
)

// storeCmd represents the store command.
var storeCmd = &cobra.Command{
	Use:   "store",
	Short: "Command to upload data to storj V3 network.",
	Long:  `Command to connect to desired IPFS account and back-up the complete data to given Storj Bucket.`,
	Run:   ipfsStore,
}

// DownCmd represents the download command.
var DownCmd = &cobra.Command{
	Use:   "download",
	Short: "Command to download data from storj V3 network.",
	Long:  `Command to download data from the Storj Bucket using the Hash.`,
	Run:   storjDownload,
}

func init() {

	// Setup the store command with its flags.
	rootCmd.AddCommand(storeCmd)
	rootCmd.AddCommand(DownCmd)
	var defaultIpfsFile string
	var defaultStorjFile string
	var defaultStorjDownloadFile string
	storeCmd.Flags().BoolP("accesskey", "a", false, "Connect to storj using access key(default connection method is by using API Key).")
	storeCmd.Flags().BoolP("share", "s", false, "For generating share access of the uploaded backup file.")
	storeCmd.Flags().StringVarP(&defaultIpfsFile, "ipfs", "i", "././config/ipfs_property_v01.json", "full filepath contaning IPFS configuration.")
	storeCmd.Flags().StringVarP(&defaultStorjFile, "storj", "u", "././config/storj_config_v01.json", "full filepath contaning storj V3 configuration.")
	DownCmd.Flags().StringVarP(&defaultIpfsFile, "ipfs", "i", "././config/ipfs_property_v01.json", "full filepath contaning IPFS configuration.")
	DownCmd.Flags().BoolP("accesskey", "a", false, "Connect to storj using access key(default connection method is by using API Key).")
	DownCmd.Flags().StringVarP(&defaultStorjFile, "storj", "u", "././config/storj_config_v01.json", "full filepath contaning storj V3 configuration.")
	DownCmd.Flags().StringVarP(&defaultStorjDownloadFile, "storjDown", "d", "././config/storj_download_v01.json", "Download data from stroj")
}

func ipfsStore(cmd *cobra.Command, args []string) {

	// Process arguments from the CLI.
	ipfsConfigfilePath, _ := cmd.Flags().GetString("ipfs")
	fullFileNameStorj, _ := cmd.Flags().GetString("storj")
	useAccessKey, _ := cmd.Flags().GetBool("accesskey")
	useAccessShare, _ := cmd.Flags().GetBool("share")

	// Read IPFS instance's configurations from an external file and create an IPFS configuration object.
	configIpfs := LoadIpfsProperty(ipfsConfigfilePath)

	// Read storj network configurations from and external file and create a storj configuration object.
	storjConfig := LoadStorjConfiguration(fullFileNameStorj)

	// Connect to storj network using the specified credentials.
	access, project := ConnectToStorj(fullFileNameStorj, storjConfig, useAccessKey)

	// Connect to IPFS using the specified credentials
	ipfsShell := ConnectToIpfs(configIpfs)

	fileHandle := GetReader(ipfsShell, configIpfs)

	fmt.Println("\nReading content from the file:", configIpfs.Path)

	// Get file name from the file path from configration file.
	_, lastFileName := filepath.Split(configIpfs.Path)

	// Create encrypt Base CID
	encryptCID, _ := ipfsShell.Add(fileHandle, shell.OnlyHash(true))

	// Open the uploaded file
	file, err1 := os.Open(configIpfs.Path)
	if err1 != nil {
		fmt.Println(err1)
	}

	// Get total size of uploaded file
	statFile, err4 := file.Stat()
	if err4 != nil {
		fmt.Println(err4)
	}
	fileSize := statFile.Size()

	givenSize, _ := strconv.ParseInt(configIpfs.ChunkSize, 0, 64)

	// Generate the number of chunk files
	noOfChunkFiles := int(fileSize/givenSize) + 1

	// Divided total uploaded file data into chunks DAG.
	chunkFile := chunker.NewSizeSplitter(file, givenSize)

	var metaFile *os.File
	metaFileName := "./metadata.txt"
	os.Remove(metaFileName)

	for i := 0; i < noOfChunkFiles; i++ {

		// Get the chunks data from the chunks DAG.
		storeChunkFile, _ := chunkFile.NextBytes()

		//Encrypt the chunk data by the given key
		key := []byte("This is a storj ipfs private key")

		encryptData, err := encrypt(key, storeChunkFile)
		if err != nil {
			log.Fatal(err)
		}

		// Create chunk CID using bytes data
		encryptChunkCID := CreateCID(ipfsShell, encryptData)

		fileName := encryptCID + "/" + encryptChunkCID

		reader := bytes.NewReader(encryptData)

		// Upload chunk data on storj Network with baseCID/chunkCID name.
		UploadData(project, storjConfig, fileName, reader)

		// Write all chunks CID into loacl disk file in append mode.
		metaFile, _ = os.OpenFile(metaFileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if _, err := metaFile.WriteString(encryptChunkCID + ","); err != nil {
			log.Fatal(err)
		}

		// Close meta file after write all chunks into the file
		err = metaFile.Close()
		if err != nil {
			log.Fatal(err)
		}
	}

	// Open meta file from local disk
	openMetaFile, _ := os.Open(metaFileName)
	metaBytes, err := ioutil.ReadAll(openMetaFile)
	if err != nil {
		log.Fatal("Reader Error:", err)
	}
	metaReader := bytes.NewReader(metaBytes)
	// Close Meta file
	openMetaFile.Close()
	// Remove meta file from local disk.
	err = os.Remove(metaFileName)
	if err != nil {
		log.Fatal(err)
	}

	metaFileStoreName := encryptCID + "/" + encryptCID + ".txt"

	// Store meta file data on storj network with baseCID/baseCID.txt
	UploadData(project, storjConfig, metaFileStoreName, metaReader)

	fmt.Println("\nAdding configuration data to IPFS: Initiated...")

	ipfsStorjData := storjConfig.Bucket + "," + storjConfig.UploadPath + "," + lastFileName

	//Encrypt the storj configration data
	enkey := []byte(storjConfig.Key)

	ipfsStorjDataBytes := []byte(ipfsStorjData)
	storjEncryptData, err := encrypt(enkey, ipfsStorjDataBytes)
	if err != nil {
		log.Fatal(err)
	}

	var hash []byte
	hash = []byte(encryptCID)

	// Create buffer for Chunk CID and encrypted Storj configurations.
	var encryptedStorjConfig []byte
	encryptedStorjConfig = append(hash, storjEncryptData...)

	// Create the CID from encrypted chunk data and encrypted
	// storj configration and enrypted private key.
	configHash, _ := ipfsShell.Add(bytes.NewReader(encryptedStorjConfig))
	fmt.Println("Shareable Hash:", configHash)

	// Create restricted shareable serialized access if share is provided as argument.
	if useAccessShare {
		ShareAccess(access, storjConfig)
	}

}

func storjDownload(cmd *cobra.Command, args []string) {

	// Process arguments from the CLI.
	ipfsConfigfilePath, _ := cmd.Flags().GetString("ipfs")
	fullFileNameDownload, _ := cmd.Flags().GetString("storjDown")
	useAccessKey, _ := cmd.Flags().GetBool("accesskey")
	fullFileNameStorj, _ := cmd.Flags().GetString("storj")

	// Read storj network configurations from and external file and create a storj configuration object.
	storjConfig := LoadStorjConfiguration(fullFileNameStorj)

	// Read IPFS instance's configurations from an external file and create an IPFS configuration object.
	configIpfs := LoadIpfsProperty(ipfsConfigfilePath)

	// Connect to storj network using the specified credentials.
	_, project := ConnectToStorj(fullFileNameStorj, storjConfig, useAccessKey)

	// Read storj network cofiguration related to download.
	downloadConfig := LoadStorjDownloadConfiguration(fullFileNameDownload)

	// Connect to ipfs network using specified credentials.
	ipfsDownloadShell := ConnectToIpfs(configIpfs)

	reader := GetReaderDownload(ipfsDownloadShell, downloadConfig.Hash)

	DownloadData(project, downloadConfig, reader)

}

// Encrypt Function to encrypt data with specified key
func encrypt(key, text []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	b := base64.StdEncoding.EncodeToString(text)
	ciphertext := make([]byte, aes.BlockSize+len(b))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}
	cfb := cipher.NewCFBEncrypter(block, iv)
	cfb.XORKeyStream(ciphertext[aes.BlockSize:], []byte(b))
	return ciphertext, nil
}
