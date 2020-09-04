# Run

> Back-up is uploaded by streaming to the Storj network.

The following flags can be used with the `store` command:

* `accesskey` - Connects to the Storj network using a serialized access key instead of an API key, satellite url and encryption passphrase.
* `share` - Generates a restricted shareable serialized access with the restrictions specified in the Storj configuration file.

Once you have built the project you can run the following:

## Get help

```
$ ./driver-ipfs --help
```

## Check version

```
$ ./driver-ipfs --version
```

## Connect to IPFS and upload the files using their corresponding hash to Storj

```
$ ./driver-ipfs store --local <path_to_ipfs_config_file> --storj <path_to_storj_config_file>
```

## Connect to IPFS and upload the files using their corresponding hash to Storj bucket using Access Key

```
$ ./driver-ipfs store --accesskey
```

## Connect to IPFS and upload the files using their corresponding hash to Storj and generate a Shareable Access Key based on restrictions in `storj_config.json`

```
$ ./driver-ipfs store --share
```

## Connect to IPFS and download the files using their corresponding hash from Storj

```
$ ./driver-IPFS download --ipfs <path_to_ipfs_config_file> --storj <path_to_storj_config_file> --storjDown <path_to_storj_download_config_file>
```

## Connect to IPFS and download the files using their corresponding hash from Storj bucket using Access Key

```
$ ./driver-IPFS download --accesskey
```
