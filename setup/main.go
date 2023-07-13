package main

import (
	"log"
	"os"
	"os/exec"

	"code.gitea.io/sdk/gitea"
)

func main() {

	ip := os.Getenv("IP_VM")
	if ip == "" {
		log.Fatal("IP_VM env variable is not set")
		os.Exit(1)
	}

	agolaAPIURL := "http://" + ip + ":8000"
	agolaRedirectURL := agolaAPIURL + "/oauth2/callback"
	giteaAPIURL := "http://" + ip + ":3000"

	// gitea client
	giteaClient, _ := gitea.NewClient(giteaAPIURL)
	giteaClient.SetBasicAuth("user01", "password")

	// create gitea oauth2 application
	giteaOauth2, _, err := giteaClient.CreateOauth2(gitea.CreateOauth2Option{
		Name:               "Agola",
		RedirectURIs:       []string{agolaRedirectURL},
		ConfidentialClient: true,
	})
	if err != nil {
		log.Fatalf("[gitea] error creating oauth2 application: %v\n", err)
		return
	}
	log.Printf("[gitea] created oauth2 application: %s\n", giteaOauth2.Name)

	// add ssh key on gitea
	key, err := os.ReadFile("/tmp/id_rsa.pub")
	if err != nil {
		log.Fatalf("[gitea] Unable to read id_rsa.pub: %v\n", err)
		return
	}
	_, _, err = giteaClient.CreatePublicKey(gitea.CreateKeyOption{
		Title:    "user01@localhost",
		Key:      string(key),
		ReadOnly: false,
	})
	if err != nil {
		log.Fatalf("[gitea] error adding ssh key: %v\n", err)
		return
	}
	log.Println("[gitea] ssh public key added")

	// create repo
	giteaRepo, _, err := giteaClient.CreateRepo(gitea.CreateRepoOption{
		Name:    "agola-example-go",
		Private: false,
	})
	if err != nil {
		log.Fatalf("[gitea] error creating repo: %v\n", err)
		return
	}
	log.Printf("[gitea] created repo: %s\n", giteaRepo.Name)

	// create remote source
	args := []string{"--token", "admintoken", "remotesource", "create", "--gateway-url", agolaAPIURL, "--name", "gitea", "--type", "gitea", "--api-url", giteaAPIURL, "--auth-type", "oauth2", "--clientid", giteaOauth2.ClientID, "--secret", giteaOauth2.ClientSecret, "--skip-ssh-host-key-check"}

	cmd := exec.Command("agola", args...)
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		log.Fatalf("[agola] error creating remote source; %v\n", err)
	}
	log.Println("[agola] created gitea remote resource")
}
