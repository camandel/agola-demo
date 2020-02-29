# Agola Demo
## All-in-one Agola demo in a Vagrant machine 
[Agola](https://agola.io/) is a modern, distributed, cloud-native CI/CD software that relies on external services like key-value store (etcd) and git platforms (github, gitlab or gitea) to work.

There is already an official [agolademo image](https://hub.docker.com/r/sorintlab/agolademo) and detailed instructions to [try it](https://agola.io/tryit/) but I needed a **quick-to-deploy, ready-to-use, pre-configured Agola demo environment** to show people how it works. That's why I created this project.

What it does:
- spins-up a Linux Vagrant machine (Fedora 30)
- installs docker-ce and docker-compose
- starts agola, etcd and gitea services as containers
- creates user01 and agola-example-go repo on gitea
- configures gitea as an agola remote source
- creates tokens for user01 on agola and gitea
- creates user01 as an agola linked account
- creates agola-example-go as agola project
- creates local user01 and deploys public ssh key to gitea
- pushes local repo to remote to trigger the first agola run
- prepares a local environment for user01 to interact with agola and gitea from CLI

**DECLAIMER**: this project is intended **only** for demo purposes. Follow the official [documentation](https://agola.io/doc/) to deploy agola in a real environment.
## Requirements
- [vagrant](https://www.vagrantup.com/)
- a vagrant provider (tested with [libvirt](https://libvirt.org/) on Linux and [virtualbox](https://www.virtualbox.org/) on Windows)
## How to use
Clone or download the repository and start the Vagrant machine:
```bash
$ git clone https://github.com/camandel/agola-demo.git && cd agola-demo
$ vagrant up
```
after a while you should see the final message with the URLs:
```
<...>
agola-demo: #################################################################################
agola-demo: #
agola-demo: # AGOLA DEMO
agola-demo: #
agola-demo: # WEB access:
agola-demo: #
agola-demo: #     agola: http://<IP>:8000      (user01/password)
agola-demo: #     gitea: http://<IP>:3000      (user01/password)
agola-demo: #
agola-demo: # SSH access with vagrant key:
agola-demo: #
agola-demo: #     [me@host] $ vagrant ssh
agola-demo: #     [vagrant@agola-demo] $ sudo su - user01
agola-demo: #     [user01@agola-demo] $ agola ...
agola-demo: #
agola-demo: # SSH access with user01 password:
agola-demo: #
agola-demo: #     [me@host] $ ssh user01@<IP>  (password)
agola-demo: #     [user01@agola-demo] $ agola ...
agola-demo: #
agola-demo: #################################################################################
```
Open the agola URL in a browser, click on the "Login" button and enter the following credential:

`user01/password`

Click on the project "agola-example-go" and you'll see the result of the first `run`.

## Push example
Modify the source code, commit the change and push it to remote:
```sh
[me@host] $ vagrant ssh
[vagrant@agola-demo ~]$ sudo su - user01

[user01@agola-demo ~]$ cd agola-example-go

[user01@agola-demo agola-example-go]$ sed -i -e 's/quote.Hello()/quote.Go()/' main.go

[user01@agola-demo agola-example-go]$ git add main.go

[user01@agola-demo agola-example-go]$ git commit -m "add Go quote"
[master 9fbd513] add Go quote
 1 file changed, 1 insertion(+), 1 deletion(-)

[user01@agola-demo agola-example-go]$ git remote -v
origin	ssh://git@<IP>:2222/user01/agola-example-go.git (fetch)
origin	ssh://git@<IP>:2222/user01/agola-example-go.git (push)

[user01@agola-demo agola-example-go]$ git push
<...>
To ssh://<IP>:2222/user01/agola-example-go.git
   74298c8..9fbd513  master -> master
```
The push will trigger a new `run`. Check it as you did before.

## Direct Run example
Generate a new agola token from UI (`User Settings` -> `User Tokens`) or from CLI (check [Create a user api token](https://agola.io/tryit/) section) for user01. Then set it to a variable and change your code example:
```sh
[user01@agola-demo ~]$ TOKEN=<user01-token>

[user01@agola-demo ~]$ cd agola-example-go

[user01@agola-demo agola-example-go]$ sed -i -e 's/quote.Hello()/quote.Go()/' main.go

[user01@agola-demo agola-example-go]$ agola directrun start --gateway-url http://<IP>:8000 --token $TOKEN
```
It will trigger a new `direct run` with the content of your working directory even if nothing has been committed or pushed to the repo. Check it in agola UI (`Direct Runs` tab).
