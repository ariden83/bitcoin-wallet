# bitcoin-wallet
HD bitcoin wallet using golang

## Command

- `make local` - launch the app
- `make test` - launch test


### Metrics

Metrics can be get on http://127.0.0.1:8082/metrics depending on configuration
The namespace is *private_api_video_ref_notifier*

*private_api_video_ref_notification_duration is the duration of a page total traitement
*private_api_video_ref_upsert is the number of traitements with label
    * `unusable`: invalid data received
    * `error`: error during traitement
    * `created`: all site segment created
  
### Healthz

Status of server can be get by url 

`
http://127.0.0.1:8082/liveness
`

and 

`
http://127.0.0.1:8082/readiness

## Resources

* See tutorial [levelup.gitconnected.com/bitcoin-hd-wallet-with-golang-and-grpc-part](https://levelup.gitconnected.com/bitcoin-hd-wallet-with-golang-and-grpc-part-l-56d8df05c602)
* See original code [github.com/LuisAcerv/btchdwallet](https://github.com/LuisAcerv/btchdwallet)
 

## PROTOC installation

##### Make sure you grab the latest version
```
curl -OL https://github.com/google/protobuf/releases/download/v3.2.0/protoc-3.2.0-linux-x86_64.zip
``` 

##### Unzip
```  
unzip protoc-3.2.0-linux-x86_64.zip -d protoc3
``` 

##### Move protoc to /usr/local/bin/
```
sudo mv protoc3/bin/* /usr/local/bin/
``` 

##### Move protoc3/include to /usr/local/include/
```
sudo mv protoc3/include/* /usr/local/include/
``` 

##### Optional: change owner
```
sudo chwon [user] /usr/local/bin/protoc
sudo chwon -R [user] /usr/local/include/google

