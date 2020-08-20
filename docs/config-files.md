# Config Files

> There are two config files that contain Storj network and IPFS connection information. The tool is designed so you can specify a config file as part of your tooling/workflow.

## `ipfs_property.json`

Inside the `./config` directory there is a `ipfs_property.json` file, with following information about your IPFS instance:

* hostName 	:- Host Name connect to IPFS
* port	   	:- Port Number connect to IPFS
* path	   	:- Path of file to be uploaded
* chunkSize	:- Size of chunks to be created for uploading

## `storj_config.json`

Inside the `./config` directory a `storj_config.json` file, with Storj network configuration information in JSON format:

* key - This is a storj ipfs private key used to encrypt data being uploaded to Storj.
* apiKey - API Key created in Storj Satellite GUI (mandatory)
* satelliteURL - Storj Satellite URL (mandatory)
* encryptionPassphrase - Storj Encryption Passphrase (mandatory)
* bucketName - Name of the bucket to upload data into (mandatory)
* uploadPath - Path on Storj Bucket to store data (optional) or "" or "/" (mandatory)
* serializedAccess - Serialized access shared while uploading data used to access bucket without API Key (mandatory)
* allowDownload - Set *true* to create serialized access with restricted download (mandatory while using *share* flag)
* allowUpload - Set *true* to create serialized access with restricted upload (mandatory while using *share* flag)
* allowList - Set *true* to create serialized access with restricted list access
* allowDelete - Set *true* to create serialized access with restricted delete
* notBefore - Set time that is always before *notAfter*
* notAfter - Set time that is always after *notBefore*

## `storj_download.json`

Inside the `./config` directory there is a `storj_download.json` file, with following information about your file to be downloaded:

* hash 			:- Hash of file to be download
* downloadPath	:- Port Number connect to IPFS
* key 			:- This is the same storj ipfs private key used to decrypt data uploaded to Storj earlier.
