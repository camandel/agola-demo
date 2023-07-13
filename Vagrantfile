# -*- mode: ruby -*-
# vi: set ft=ruby :

Vagrant.configure("2") do |config|
    
  config.vm.box = "fedora/38-cloud-base"
  config.vm.box_version = "38.20230413.1"

  config.vm.define "agola-demo"
  config.vm.hostname = "agola-demo"
  
  config.vm.provider "libvirt" do |lv|
    lv.qemu_use_session = false
    lv.memory = 2048
    lv.cpus = 2
  end

  config.vm.provider "virtualbox" do |vb|
    vb.customize ['modifyvm', :id, '--memory', '2048', '--cpus', '2']
  end

  config.vm.synced_folder ".", "/vagrant", disabled: true

  config.vm.provision "file", source: "./templates", destination: "./templates"
  config.vm.provision "file", source: "./setup", destination: "setup"

  config.vm.provision "shell", privileged: false, inline: <<-'SHELL'
    ## get vm ip
    export IP_VM=$(ip addr show eth0 | grep "inet\b" | awk '{print $2}' | cut -d/ -f1)
    echo "export IP_VM=${IP_VM}" | sudo tee /etc/profile.d/agola.sh && source /etc/profile.d/agola.sh

    # install requirements
    sudo dnf config-manager --add-repo https://download.docker.com/linux/fedora/docker-ce.repo
    sudo dnf install -y docker-ce docker-compose-plugin vim-enhanced bind-utils bash-completion git
    sudo systemctl enable docker --now
    sudo mkdir -p /data/agola /data/gitea && sudo chown vagrant:vagrant /data/agola /data/gitea        
    # prepare user01 local environment
    sudo useradd user01
    echo password | sudo passwd --stdin user01
    sudo su - user01 -c 'ssh-keygen -t rsa -b 4096 -C "user01@example.com" -f /home/user01/.ssh/id_rsa -N ""'
    sudo su - user01 -c "echo -e \"Host ${IP_VM}\n\tStrictHostKeyChecking no\" > /home/user01/.ssh/config && chmod 600 /home/user01/.ssh/config"
    sudo su - user01 -c "git config --global user.email 'user01@example.com' && git config --global  user.name 'User01'"
 
    # create image for agola-setup
    sudo docker build -t local-setup ./setup
    # generate agola config file in /data/agola/config.yml
    cat templates/agola-config.yml | envsubst > /data/agola/config.yml

    # generate docker compose config
    cat templates/docker-compose.yml | envsubst > docker-compose.yml

    # start containers
    sudo docker compose -p agola-demo up -d

    # wait for gitea is up & running
    until sudo docker logs gitea | grep Listen; do sleep 3; done
 
    # create user01 on gitea
    sudo docker exec gitea su git -c '/usr/local/bin/gitea admin user create --username user01 --password password --email user01@example.com --admin'

    # setup gitea and agola services
    sudo docker run --rm -e IP_VM=${IP_VM} -v /home/user01/.ssh/id_rsa.pub:/tmp/id_rsa.pub:ro local-setup

    # setup local repo
    sudo su - user01 -c "git clone https://github.com/agola-io/agola-example-go.git"
    sudo su - user01 -c "cd agola-example-go && git remote remove origin && git remote add origin ssh://git@${IP_VM}:2222/user01/agola-example-go.git"

    # copy agola binary
    sudo docker cp agola:/bin/agola /usr/local/bin
  
  SHELL

  config.vm.provision "shell", privileged: false, run: "always", inline: <<-'SHELL'
    ## get vm ip
    export IP_VM=$(ip addr show eth0 | grep "inet\b" | awk '{print $2}' | cut -d/ -f1)
    
    # post-up message
    echo "#################################################################################"
    echo "#"
    echo "# AGOLA DEMO"
    echo "#"
    echo "# WEB access:"
    echo "#"
    echo "#     agola: http://${IP_VM}:8000      (user01/password)"
    echo "#     gitea: http://${IP_VM}:3000      (user01/password)"
    echo "#"
    echo "# SSH access with vagrant key:"
    echo "#"
    echo "#     [me@host] $ vagrant ssh"
    echo "#     [vagrant@agola-demo] $ sudo su - user01"
    echo "#     [user01@agola-demo] $ agola ..."
    echo "#"
    echo "# SSH access with user01 password:"
    echo "#"
    echo "#     [me@host] $ ssh user01@${IP_VM}    (password)"
    echo "#     [user01@agola-demo] $ agola ..."
    echo "#"
    echo "#################################################################################"

  SHELL
  
end
