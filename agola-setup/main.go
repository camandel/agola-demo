package main

import (
	"context"
	"io/ioutil"
	"log"
	"os"
	"path"
	"text/template"

	gwapitypes "agola.io/agola/services/gateway/api/types"
	gwclient "agola.io/agola/services/gateway/client"

	"code.gitea.io/sdk/gitea"
)

const configTemplate = `
gateway:
  apiExposedURL: "http://{{.Ip}}:8000"
  webExposedURL: "http://{{.Ip}}:8000"
  runserviceURL: "http://localhost:4000"
  configstoreURL: "http://localhost:4002"
  gitserverURL: "http://localhost:4003"

  web:
    listenAddress: ":8000"
  tokenSigning:
    # hmac or rsa (it possible use rsa)
    method: hmac
    # key to use when signing with hmac
    key: supersecretsigningkey
    # paths to the private and public keys in pem encoding when using rsa signing
    #privateKeyPath: /path/to/privatekey.pem
    #publicKeyPath: /path/to/public.pem
  adminToken: "admintoken"

scheduler:
  runserviceURL: "http://localhost:4000"

notification:
  webExposedURL: "http://{{.Ip}}:8000"
  runserviceURL: "http://localhost:4000"
  configstoreURL: "http://localhost:4002"
  etcd:
    endpoints: "http://etcd:2379"

configstore:
  dataDir: /data/agola/configstore
  etcd:
    endpoints: "http://etcd:2379"
  objectStorage:
    type: posix
    path: /data/agola/configstore/ost
  web:
    listenAddress: ":4002"

runservice:
  #debug: true
  dataDir: /data/agola/runservice
  etcd:
    endpoints: "http://etcd:2379"
  objectStorage:
    type: posix
    path: /data/agola/runservice/ost
  web:
    listenAddress: ":4000"

executor:
  dataDir: /data/agola/executor
  # The directory containing the toolbox compiled for the various supported architectures
  toolboxPath: ./bin
  runserviceURL: "http://localhost:4000"
  web:
    listenAddress: ":4001"
  activeTasksLimit: 2
  driver:
    type: docker

gitserver:
  dataDir: /data/agola/gitserver
  gatewayURL: "http://localhost:8000"
  web:
    listenAddress: ":4003"
`
type agolaConfig struct{
        Ip string
}

func setup(ip string) {

	agolaAPIURL := "http://" + ip + ":8000"
	giteaAPIURL := "http://" + ip + ":3000"
	giteaClient := gitea.NewClient(giteaAPIURL, "")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// create gittea access token
	giteaToken, err := giteaClient.CreateAccessToken("user01", "password", gitea.CreateAccessTokenOption{Name: "token01"})
	if err != nil {
		log.Fatalf("unexpected err: %v\n", err)
		return
	}
	log.Printf("created gitea user token: %s\n", giteaToken.Token)

	// create LinkedAccount
	gwClient := gwclient.NewClient(agolaAPIURL, "admintoken")
	user, _, err := gwClient.CreateUser(ctx, &gwapitypes.CreateUserRequest{UserName: "user01"})
	if err != nil {
		log.Fatalf("unexpected err: %v\n", err)
		return
	}
	log.Printf("created agola user: %s\n", user.UserName)

	// create agola user token
	token, _, err := gwClient.CreateUserToken(ctx, "user01", &gwapitypes.CreateUserTokenRequest{TokenName: "token01"})
	if err != nil {
		log.Fatalf("unexpected err: %v\n", err)
		return
	}
	log.Printf("created agola user token: %s\n", token.Token)

	rs, _, err := gwClient.CreateRemoteSource(ctx, &gwapitypes.CreateRemoteSourceRequest{
		Name:                "gitea",
		APIURL:              giteaAPIURL,
		Type:                "gitea",
		AuthType:            "password",
		SkipSSHHostKeyCheck: true,
	})
	if err != nil {
		log.Fatalf("unexpected err: %v\n", err)
		return
	}
	log.Printf("created agola remote source: %s\n", rs.Name)

	// From now use the agola user token
	gwClient = gwclient.NewClient(agolaAPIURL, token.Token)

	_, _, err = gwClient.CreateUserLA(ctx, "user01", &gwapitypes.CreateUserLARequest{
		RemoteSourceName:          "gitea",
		RemoteSourceLoginName:     "user01",
		RemoteSourceLoginPassword: "password",
	})
	if err != nil {
		log.Fatalf("unexpected err: %v\n", err)
		return
	}
	log.Println("created user linked account")

	giteaClient = gitea.NewClient(giteaAPIURL, giteaToken.Token)

	// add ssh key
	key, err := ioutil.ReadFile("/tmp/id_rsa.pub")
	if err != nil {
		log.Fatalf("Unable to read id_rsa.pub: %v\n", err)
		return
	}
	_, err = giteaClient.AdminCreateUserPublicKey("user01", gitea.CreateKeyOption{
		Title:    "user01@localhost",
		Key:      string(key),
		ReadOnly: true,
	})
	if err != nil {
		log.Fatalf("unexpected err: %v\n", err)
		return
	}
	log.Println("ssh public key added")

	// create repo
	giteaRepo, err := giteaClient.CreateRepo(gitea.CreateRepoOption{
		Name:    "agola-example-go",
		Private: false,
	})
	if err != nil {
		log.Fatalf("unexpected err: %v\n", err)
		return
	}
	log.Printf("created gitea repo: %s\n", giteaRepo.Name)

	_, _, err = gwClient.CreateProject(ctx, &gwapitypes.CreateProjectRequest{
		Name:             "agola-example-go",
		ParentRef:        path.Join("user", "user01"),
		RemoteSourceName: "gitea",
		RepoPath:         path.Join("user01", "agola-example-go"),
		Visibility:       gwapitypes.VisibilityPublic,
	})
	if err != nil {
		log.Printf("unexpected err: %v\n", err)
		return
	}
	log.Println("Setup completed")
}

func generateConfig(ip string) {
	agolaConfig := agolaConfig{
		Ip: ip,
	}

	t := template.New("config-template")

	t, err := t.Parse(configTemplate)
	if err != nil {
		log.Fatal("parse template: ", err)
		return
	}

        f, err := os.Create("/data/agola/config.yml")
        if err != nil {
                log.Fatal("create /data/agola/config.yml file: ", err)
                return
        }

	err = t.Execute(f, agolaConfig)
	if err != nil {
		log.Fatal("generate config.yml from template: ", err)
		return
        }
        log.Println("Agola config file created in /data/agola/config.yml")
}

func main() {

        args:= os.Args
        if len(args) != 3 {
                log.Fatal("setup [ template | services ] <ip>")
                os.Exit(1)
        }

        ip := args[2]

        switch args[1] {
        case "template":
                generateConfig(ip)
        case "services":
                setup(ip)
        default:
            log.Fatal("setup [ template | services ] <ip>")
            os.Exit(1)
        }
}
