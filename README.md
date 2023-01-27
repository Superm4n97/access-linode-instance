# access-linode-instance

### run:
* Instead of SSH key you can also export `USERNAME` and `PASSWORD` to access linode.
```bash
export LINODE_CLI_TOKEN=???????
export USERNAME=??????
export PASSWORD=??????
go run main.go
```
* For custom SSH key, you can export the ssh file path to environment as `SSH_PATH`
```bash
export LINODE_CLI_TOKEN=???????
export SSH_PATH=???????
go run main.go
```
* For default SSH key:
```bash
export LINODE_CLI_TOKEN=???????
go run main.go
```


### linode-credential
CredentialOptions is a structure which contains the fields Username, Password, and PrivateKey. It will look for the credential sequentially as follows:
* check for `username` and `password`
* if username and password are not given then look for the `private key`
* if private key is not given then use the local `HomeDir/.ssh/id_rsa` to create a private key

to use the private key you have to add SSH to linode

### to.do
* restructure the credential structure: instead of username options will be user(root)
* You can remove the `CopyOption` structure as you don't have any default dependencies
* You may need to update the `README.md` file
* test create instance, get instance with id, and copy a file to remote.